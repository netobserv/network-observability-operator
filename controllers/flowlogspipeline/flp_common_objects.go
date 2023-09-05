package flowlogspipeline

import (
	"embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/confgen"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/globals"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/filters"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

const (
	configVolume                = "config-volume"
	configPath                  = "/etc/flowlogs-pipeline"
	configFile                  = "config.json"
	metricsConfigDir            = "metrics_definitions"
	lokiToken                   = "loki-token"
	healthServiceName           = "health"
	prometheusServiceName       = "prometheus"
	profilePortName             = "pprof"
	healthTimeoutSeconds        = 5
	livenessPeriodSeconds       = 10
	startupFailureThreshold     = 5
	startupPeriodSeconds        = 10
	conntrackTerminatingTimeout = 5 * time.Second
	conntrackEndTimeout         = 10 * time.Second
	conntrackHeartbeatInterval  = 30 * time.Second
)

type ConfKind string

const (
	ConfMonolith         ConfKind = "allInOne"
	ConfKafkaIngester    ConfKind = "kafkaIngester"
	ConfKafkaTransformer ConfKind = "kafkaTransformer"
)

var FlpConfSuffix = map[ConfKind]string{
	ConfMonolith:         "",
	ConfKafkaIngester:    "-ingester",
	ConfKafkaTransformer: "-transformer",
}

type builder struct {
	info     *reconcilers.Instance
	labels   map[string]string
	selector map[string]string
	desired  *flowslatest.FlowCollectorSpec
	promTLS  *flowslatest.CertificateReference
	confKind ConfKind
	volumes  volumes.Builder
}

func newBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, ck ConfKind) builder {
	version := helper.ExtractVersion(info.Image)
	name := name(ck)
	var promTLS flowslatest.CertificateReference
	switch desired.Processor.Metrics.Server.TLS.Type {
	case flowslatest.ServerTLSProvided:
		promTLS = *desired.Processor.Metrics.Server.TLS.Provided
	case flowslatest.ServerTLSAuto:
		promTLS = flowslatest.CertificateReference{
			Type:     "secret",
			Name:     promServiceName(ck),
			CertFile: "tls.crt",
			CertKey:  "tls.key",
		}
	}
	return builder{
		info: info,
		labels: map[string]string{
			"app":     name,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			"app": name,
		},
		desired:  desired,
		confKind: ck,
		promTLS:  &promTLS,
	}
}

func name(ck ConfKind) string                 { return constants.FLPName + FlpConfSuffix[ck] }
func RoleBindingName(ck ConfKind) string      { return name(ck) + "-role" }
func RoleBindingMonoName(ck ConfKind) string  { return name(ck) + "-role-mono" }
func promServiceName(ck ConfKind) string      { return name(ck) + "-prom" }
func configMapName(ck ConfKind) string        { return name(ck) + "-config" }
func serviceMonitorName(ck ConfKind) string   { return name(ck) + "-monitor" }
func prometheusRuleName(ck ConfKind) string   { return name(ck) + "-alert" }
func (b *builder) name() string               { return name(b.confKind) }
func (b *builder) promServiceName() string    { return promServiceName(b.confKind) }
func (b *builder) configMapName() string      { return configMapName(b.confKind) }
func (b *builder) serviceMonitorName() string { return serviceMonitorName(b.confKind) }
func (b *builder) prometheusRuleName() string { return prometheusRuleName(b.confKind) }

func (b *builder) portProtocol() corev1.Protocol {
	if helper.UseEBPF(b.desired) {
		return corev1.ProtocolTCP
	}
	return corev1.ProtocolUDP
}

func (b *builder) podTemplate(hasHostPort, hostNetwork bool, annotations map[string]string) corev1.PodTemplateSpec {
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if hasHostPort {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName,
			HostPort:      b.desired.Processor.Port,
			ContainerPort: b.desired.Processor.Port,
			Protocol:      b.portProtocol(),
		}}
		// This allows deploying an instance in the master node, the same technique used in the
		// companion ovnkube-node daemonset definition
		tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: b.desired.Processor.HealthPort,
	})

	ports = append(ports, corev1.ContainerPort{
		Name:          prometheusServiceName,
		ContainerPort: b.desired.Processor.Metrics.Server.Port,
	})

	if b.desired.Processor.ProfilePort > 0 {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: b.desired.Processor.ProfilePort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	volumeMounts := b.volumes.AppendMounts([]corev1.VolumeMount{{
		MountPath: configPath,
		Name:      configVolume,
	}})
	volumes := b.volumes.AppendVolumes([]corev1.Volume{{
		Name: configVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: b.configMapName(),
				},
			},
		},
	}})

	var envs []corev1.EnvVar
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(b.desired.Processor.Debug.Env) {
		envs = append(envs, corev1.EnvVar{Name: pair[0], Value: pair[1]})
	}

	container := corev1.Container{
		Name:            constants.FLPName,
		Image:           b.info.Image,
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
	}
	if helper.PtrBool(b.desired.Processor.EnableKubeProbes) {
		container.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/live",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds: healthTimeoutSeconds,
			PeriodSeconds:  livenessPeriodSeconds,
		}
		container.StartupProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.FromString(healthServiceName),
				},
			},
			TimeoutSeconds:   healthTimeoutSeconds,
			PeriodSeconds:    startupPeriodSeconds,
			FailureThreshold: startupFailureThreshold,
		}
	}
	dnsPolicy := corev1.DNSClusterFirst
	if hostNetwork {
		dnsPolicy = corev1.DNSClusterFirstWithHostNet
	}
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/scrape_port"] = fmt.Sprint(b.desired.Processor.Metrics.Server.Port)

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Tolerations:        tolerations,
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: b.name(),
			HostNetwork:        hostNetwork,
			DNSPolicy:          dnsPolicy,
		},
	}
}

//go:embed metrics_definitions
var metricsConfigEmbed embed.FS

// obtainMetricsConfiguration returns the configuration info for the prometheus stage needed to
// supply the metrics for those metrics
func (b *builder) obtainMetricsConfiguration() (api.PromMetricsItems, error) {
	entries, err := metricsConfigEmbed.ReadDir(metricsConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access metrics_definitions directory: %w", err)
	}

	cg := confgen.NewConfGen(&confgen.Options{
		GenerateStages: []string{"encode_prom"},
		SkipWithTags:   b.desired.Processor.Metrics.IgnoreTags,
	})

	for _, entry := range entries {
		fileName := entry.Name()
		if fileName == "config.yaml" {
			continue
		}
		srcPath := filepath.Join(metricsConfigDir, fileName)

		input, err := metricsConfigEmbed.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("error reading metrics file %s; %w", srcPath, err)
		}
		err = cg.ParseDefinition(fileName, input)
		if err != nil {
			return nil, fmt.Errorf("error parsing metrics file %s; %w", srcPath, err)
		}
	}

	stages := cg.GenerateTruncatedConfig()
	if len(stages) != 1 {
		return nil, fmt.Errorf("error generating truncated config, 1 stage expected in %v", stages)
	}
	if stages[0].Encode == nil || stages[0].Encode.Prom == nil {
		return nil, fmt.Errorf("error generating truncated config, Encode expected in %v", stages)
	}
	return stages[0].Encode.Prom.Metrics, nil
}

func (b *builder) addTransformStages(stage *config.PipelineBuilderStage) error {
	lastStage := *stage
	indexFields := constants.LokiIndexFields

	lastStage = b.addTransformFilter(lastStage)

	indexFields, lastStage = b.addConnectionTracking(indexFields, lastStage)

	// enrich stage (transform) configuration
	enrichedStage := lastStage.TransformNetwork("enrich", api.TransformNetwork{
		Rules: api.NetworkTransformRules{{
			Input:  "SrcAddr",
			Output: "SrcK8S",
			Type:   api.AddKubernetesRuleType,
		}, {
			Input:  "DstAddr",
			Output: "DstK8S",
			Type:   api.AddKubernetesRuleType,
		}, {
			Type: api.ReinterpretDirectionRuleType,
		}},
		DirectionInfo: api.NetworkTransformDirectionInfo{
			ReporterIPField:    "AgentIP",
			SrcHostField:       "SrcK8S_HostIP",
			DstHostField:       "DstK8S_HostIP",
			FlowDirectionField: "FlowDirection",
			IfDirectionField:   "IfDirection",
		},
	})

	// loki stage (write) configuration
	if helper.UseLoki(b.desired) {
		lokiWrite := api.WriteLoki{
			Labels:         indexFields,
			BatchSize:      int(b.desired.Loki.BatchSize),
			BatchWait:      helper.UnstructuredDuration(b.desired.Loki.BatchWait),
			MaxBackoff:     helper.UnstructuredDuration(b.desired.Loki.MaxBackoff),
			MaxRetries:     int(helper.PtrInt32(b.desired.Loki.MaxRetries)),
			MinBackoff:     helper.UnstructuredDuration(b.desired.Loki.MinBackoff),
			StaticLabels:   model.LabelSet{},
			Timeout:        helper.UnstructuredDuration(b.desired.Loki.Timeout),
			URL:            b.desired.Loki.URL,
			TimestampLabel: "TimeFlowEndMs",
			TimestampScale: "1ms",
			TenantID:       b.desired.Loki.TenantID,
		}

		for k, v := range b.desired.Loki.StaticLabels {
			lokiWrite.StaticLabels[model.LabelName(k)] = model.LabelValue(v)
		}

		var authorization *promConfig.Authorization
		if helper.LokiUseHostToken(&b.desired.Loki) || helper.LokiForwardUserToken(&b.desired.Loki) {
			b.volumes.AddToken(constants.FLPName)
			authorization = &promConfig.Authorization{
				Type:            "Bearer",
				CredentialsFile: constants.TokensPath + constants.FLPName,
			}
		}

		if b.desired.Loki.TLS.Enable {
			if b.desired.Loki.TLS.InsecureSkipVerify {
				lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
					Authorization: authorization,
					TLSConfig: promConfig.TLSConfig{
						InsecureSkipVerify: true,
					},
				}
			} else {
				caPath := b.volumes.AddCACertificate(&b.desired.Loki.TLS, "loki-certs")
				lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
					Authorization: authorization,
					TLSConfig: promConfig.TLSConfig{
						CAFile: caPath,
					},
				}
			}
		} else {
			lokiWrite.ClientConfig = &promConfig.HTTPClientConfig{
				Authorization: authorization,
			}
		}
		enrichedStage.WriteLoki("loki", lokiWrite)
	}

	// write on Stdout if logging trace enabled
	if b.desired.Processor.LogLevel == "trace" {
		enrichedStage.WriteStdout("stdout", api.WriteStdout{Format: "json"})
	}

	// obtain encode_prometheus stage from metrics_definitions
	promMetrics, err := b.obtainMetricsConfiguration()
	if err != nil {
		return err
	}

	if len(promMetrics) > 0 {
		// prometheus stage (encode) configuration
		promEncode := api.PromEncode{
			Prefix:  "netobserv_",
			Metrics: promMetrics,
		}
		enrichedStage.EncodePrometheus("prometheus", promEncode)
	}

	b.addCustomExportStages(&enrichedStage)
	return nil
}

func (b *builder) addConnectionTracking(indexFields []string, lastStage config.PipelineBuilderStage) ([]string, config.PipelineBuilderStage) {
	outputFields := []api.OutputField{
		{
			Name:      "Bytes",
			Operation: "sum",
		},
		{
			Name:      "Bytes",
			Operation: "sum",
			SplitAB:   true,
		},
		{
			Name:      "Packets",
			Operation: "sum",
		},
		{
			Name:      "Packets",
			Operation: "sum",
			SplitAB:   true,
		},
		{
			Name:      "numFlowLogs",
			Operation: "count",
		},
		{
			Name:          "TimeFlowStartMs",
			Operation:     "min",
			ReportMissing: true,
		},
		{
			Name:          "TimeFlowEndMs",
			Operation:     "max",
			ReportMissing: true,
		},
		{
			Name:          "FlowDirection",
			Operation:     "first",
			ReportMissing: true,
		},
		{
			Name:          "IfDirection",
			Operation:     "first",
			ReportMissing: true,
		},
		{
			Name:          "AgentIP",
			Operation:     "first",
			ReportMissing: true,
		},
	}

	if helper.IsPktDropEnabled(&b.desired.Agent.EBPF) {
		outputPktDropFields := []api.OutputField{
			{
				Name:      "PktDropBytes",
				Operation: "sum",
			},
			{
				Name:      "PktDropBytes",
				Operation: "sum",
				SplitAB:   true,
			},
			{
				Name:      "PktDropPackets",
				Operation: "sum",
			},
			{
				Name:      "PktDropPackets",
				Operation: "sum",
				SplitAB:   true,
			},
			{
				Name:      "PktDropLatestState",
				Operation: "last",
			},
			{
				Name:      "PktDropLatestDropCause",
				Operation: "last",
			},
		}
		outputFields = append(outputFields, outputPktDropFields...)
	}

	if helper.IsDNSTrackingEnabled(&b.desired.Agent.EBPF) {
		outDNSTrackingFields := []api.OutputField{
			{
				Name:      "DnsFlagsResponseCode",
				Operation: "last",
			},
			{
				Name:      "DnsLatencyMs",
				Operation: "max",
			},
		}
		outputFields = append(outputFields, outDNSTrackingFields...)
	}

	if helper.IsFlowRTTEnabled(&b.desired.Agent.EBPF) {
		outputFields = append(outputFields, api.OutputField{
			Name:      "MaxTimeFlowRttNs",
			Operation: "max",
			Input:     "TimeFlowRttNs",
		})
	}

	// Connection tracking stage (only if LogTypes is not FLOWS)
	if b.desired.Processor.LogTypes != nil && *b.desired.Processor.LogTypes != flowslatest.LogTypeFlows {
		indexFields = append(indexFields, constants.LokiConnectionIndexFields...)
		outputRecordTypes := helper.GetRecordTypes(&b.desired.Processor)

		terminatingTimeout := conntrackTerminatingTimeout
		if b.desired.Processor.ConversationTerminatingTimeout != nil {
			terminatingTimeout = b.desired.Processor.ConversationTerminatingTimeout.Duration
		}

		endTimeout := conntrackEndTimeout
		if b.desired.Processor.ConversationEndTimeout != nil {
			endTimeout = b.desired.Processor.ConversationEndTimeout.Duration
		}

		heartbeatInterval := conntrackHeartbeatInterval
		if b.desired.Processor.ConversationHeartbeatInterval != nil {
			heartbeatInterval = b.desired.Processor.ConversationHeartbeatInterval.Duration
		}

		lastStage = lastStage.ConnTrack("extract_conntrack", api.ConnTrack{
			KeyDefinition: api.KeyDefinition{
				FieldGroups: []api.FieldGroup{
					{Name: "src", Fields: []string{"SrcAddr", "SrcPort"}},
					{Name: "dst", Fields: []string{"DstAddr", "DstPort"}},
					{Name: "common", Fields: []string{"Proto"}},
				},
				Hash: api.ConnTrackHash{
					FieldGroupRefs: []string{
						"common",
					},
					FieldGroupARef: "src",
					FieldGroupBRef: "dst",
				},
			},
			OutputRecordTypes: outputRecordTypes,
			OutputFields:      outputFields,
			Scheduling: []api.ConnTrackSchedulingGroup{
				{
					Selector:             nil, // Default group. Match all flowlogs
					HeartbeatInterval:    api.Duration{Duration: heartbeatInterval},
					EndConnectionTimeout: api.Duration{Duration: endTimeout},
					TerminatingTimeout:   api.Duration{Duration: terminatingTimeout},
				},
			},
			TCPFlags: api.ConnTrackTCPFlags{
				FieldName:           "Flags",
				DetectEndConnection: true,
				SwapAB:              true,
			},
		})

	}
	return indexFields, lastStage
}

func (b *builder) addTransformFilter(lastStage config.PipelineBuilderStage) config.PipelineBuilderStage {
	var clusterName string
	transformFilterRules := []api.TransformFilterRule{}

	if b.desired.Processor.ClusterName != "" {
		clusterName = b.desired.Processor.ClusterName
	} else {
		//take clustername from openshift
		clusterName = string(globals.DefaultClusterID)
	}
	if clusterName != "" {
		transformFilterRules = []api.TransformFilterRule{
			{
				Input: "K8S_ClusterName",
				Type:  "add_field_if_doesnt_exist",
				Value: clusterName,
			},
		}
	}

	// Filter-out unused fields?
	if helper.PtrBool(b.desired.Processor.DropUnusedFields) {
		if helper.UseIPFIX(b.desired) {
			rules := filters.GetOVSGoflowUnusedRules()
			transformFilterRules = append(transformFilterRules, rules...)
		}
		// Else: nothing for eBPF at the moment
	}
	if len(transformFilterRules) > 0 {
		lastStage = lastStage.TransformFilter("filter", api.TransformFilter{
			Rules: transformFilterRules,
		})
	}
	return lastStage
}

func (b *builder) addCustomExportStages(enrichedStage *config.PipelineBuilderStage) {
	for i, exporter := range b.desired.Exporters {
		if exporter.Type == flowslatest.KafkaExporter {
			b.createKafkaWriteStage(fmt.Sprintf("kafka-export-%d", i), &exporter.Kafka, enrichedStage)
		}
		if exporter.Type == flowslatest.IpfixExporter {
			createIPFIXWriteStage(fmt.Sprintf("IPFIX-export-%d", i), &exporter.IPFIX, enrichedStage)
		}
	}
}

func (b *builder) createKafkaWriteStage(name string, spec *flowslatest.FlowCollectorKafka, fromStage *config.PipelineBuilderStage) config.PipelineBuilderStage {
	return fromStage.EncodeKafka(name, api.EncodeKafka{
		Address: spec.Address,
		Topic:   spec.Topic,
		TLS:     b.getKafkaTLS(&spec.TLS, name),
		SASL:    b.getKafkaSASL(&spec.SASL, name),
	})
}

func createIPFIXWriteStage(name string, spec *flowslatest.FlowCollectorIPFIXReceiver, fromStage *config.PipelineBuilderStage) config.PipelineBuilderStage {
	return fromStage.WriteIpfix(name, api.WriteIpfix{
		TargetHost:   spec.TargetHost,
		TargetPort:   spec.TargetPort,
		Transport:    getIPFIXTransport(spec.Transport),
		EnterpriseID: 2,
	})
}

func (b *builder) getKafkaTLS(tls *flowslatest.ClientTLS, volumeName string) *api.ClientTLS {
	if tls.Enable {
		caPath, userCertPath, userKeyPath := b.volumes.AddMutualTLSCertificates(tls, volumeName)
		return &api.ClientTLS{
			InsecureSkipVerify: tls.InsecureSkipVerify,
			CACertPath:         caPath,
			UserCertPath:       userCertPath,
			UserKeyPath:        userKeyPath,
		}
	}
	return nil
}

func (b *builder) getKafkaSASL(sasl *flowslatest.SASLConfig, volumePrefix string) *api.SASLConfig {
	if !helper.UseSASL(sasl) {
		return nil
	}
	t := "plain"
	if sasl.Type == flowslatest.SASLScramSHA512 {
		t = "scramSHA512"
	}
	basePath := b.volumes.AddVolume(&sasl.Reference, volumePrefix+"-sasl")
	return &api.SASLConfig{
		Type:             t,
		ClientIDPath:     path.Join(basePath, sasl.ClientIDKey),
		ClientSecretPath: path.Join(basePath, sasl.ClientSecretKey),
	}
}

func getIPFIXTransport(transport string) string {
	switch transport {
	case "UDP":
		return "udp"
	default:
		return "tcp" //always fallback on tcp
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) configMap(stages []config.Stage, parameters []config.StageParam) (*corev1.ConfigMap, string, error) {
	metricsSettings := config.MetricsSettings{
		Port:    int(b.desired.Processor.Metrics.Server.Port),
		Prefix:  "netobserv_",
		NoPanic: true,
	}
	if b.desired.Processor.Metrics.Server.TLS.Type != flowslatest.ServerTLSDisabled {
		cert, key := b.volumes.AddCertificate(b.promTLS, "prom-certs")
		metricsSettings.TLS = &api.PromTLSConf{
			CertPath: cert,
			KeyPath:  key,
		}
	}
	config := map[string]interface{}{
		"log-level": b.desired.Processor.LogLevel,
		"health": map[string]interface{}{
			"port": b.desired.Processor.HealthPort,
		},
		"pipeline":        stages,
		"parameters":      parameters,
		"metricsSettings": metricsSettings,
	}
	if b.desired.Processor.ProfilePort > 0 {
		config["profile"] = map[string]interface{}{
			"port": b.desired.Processor.ProfilePort,
		}
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return nil, "", err
	}
	configStr := string(bs)

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.configMapName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Data: map[string]string{
			configFile: configStr,
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(configStr))
	digest := strconv.FormatUint(hasher.Sum64(), 36)
	return &configMap, digest, nil
}

func (b *builder) promService() *corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.promServiceName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: b.selector,
			Ports: []corev1.ServicePort{{
				Name:     prometheusServiceName,
				Port:     b.desired.Processor.Metrics.Server.Port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt(int(b.desired.Processor.Metrics.Server.Port)),
			}},
		},
	}
	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		svc.ObjectMeta.Annotations = map[string]string{
			constants.OpenShiftCertificateAnnotation: b.promServiceName(),
		}
	}
	return &svc
}

func (b *builder) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name(),
			Namespace: b.info.Namespace,
			Labels: map[string]string{
				"app": b.name(),
			},
		},
	}
}

func (b *builder) clusterRoleBinding(ck ConfKind, mono bool) *rbacv1.ClusterRoleBinding {
	var rbName string
	if mono {
		rbName = RoleBindingMonoName(ck)
	} else {
		rbName = RoleBindingName(ck)
	}
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rbName,
			Labels: map[string]string{
				"app": b.name(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name(ck),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      b.name(),
			Namespace: b.info.Namespace,
		}},
	}
}

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", b.promServiceName(), b.info.Namespace)
	flpServiceMonitorObject := monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.serviceMonitorName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     prometheusServiceName,
					Interval: "15s",
					Scheme:   "http",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{
					b.info.Namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: b.selector,
			},
		},
	}
	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSAuto {
		flpServiceMonitorObject.Spec.Endpoints[0].Scheme = "https"
		flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				ServerName: serverName,
			},
			CAFile: "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt",
		}
	}

	if b.desired.Processor.Metrics.Server.TLS.Type == flowslatest.ServerTLSProvided {
		flpServiceMonitorObject.Spec.Endpoints[0].Scheme = "https"
		flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig = &monitoringv1.TLSConfig{
			SafeTLSConfig: monitoringv1.SafeTLSConfig{
				ServerName:         serverName,
				InsecureSkipVerify: true,
			},
		}
	}

	return &flpServiceMonitorObject
}

func shouldAddAlert(name flowslatest.FLPAlert, disabledList []flowslatest.FLPAlert) bool {
	for _, disabledAlert := range disabledList {
		if name == disabledAlert {
			return false
		}
	}
	return true
}

func (b *builder) prometheusRule() *monitoringv1.PrometheusRule {
	rules := []monitoringv1.Rule{}

	// Not receiving flows
	if shouldAddAlert(flowslatest.AlertNoFlows, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: flowslatest.AlertNoFlows,
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
				"summary":     "NetObserv flowlogs-pipeline is not receiving any flow",
			},
			Expr: intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m])) == 0"),
			For:  "10m",
			Labels: map[string]string{
				"severity": "warning",
				"app":      "netobserv",
			},
		})
	}

	// Flows getting dropped by loki library
	if shouldAddAlert(flowslatest.AlertLokiError, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: flowslatest.AlertLokiError,
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is dropping flows because of loki errors, loki may be down or having issues ingesting every flows. Please check loki and flowlogs-pipeline logs.",
				"summary":     "NetObserv flowlogs-pipeline is dropping flows because of loki errors",
			},
			Expr: intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
			For:  "10m",
			Labels: map[string]string{
				"severity": "warning",
				"app":      "netobserv",
			},
		})
	}

	flpPrometheusRuleObject := monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.prometheusRuleName(),
			Labels:    b.labels,
			Namespace: b.info.Namespace,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name:  "NetobservFlowLogsPipeline",
					Rules: rules,
				},
			},
		},
	}
	return &flpPrometheusRuleObject
}
