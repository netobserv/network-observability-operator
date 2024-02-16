package flp

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
)

const (
	configVolume            = "config-volume"
	configPath              = "/etc/flowlogs-pipeline"
	configFile              = "config.json"
	healthServiceName       = "health"
	prometheusServiceName   = "prometheus"
	profilePortName         = "pprof"
	healthTimeoutSeconds    = 5
	livenessPeriodSeconds   = 10
	startupFailureThreshold = 5
	startupPeriodSeconds    = 10
	appLabel                = "app"
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

type Builder struct {
	info        *reconcilers.Instance
	labels      map[string]string
	selector    map[string]string
	desired     *flowslatest.FlowCollectorSpec
	flowMetrics *metricslatest.FlowMetricList
	promTLS     *flowslatest.CertificateReference
	confKind    ConfKind
	volumes     volumes.Builder
	loki        *helper.LokiConfig
	pipeline    *PipelineBuilder
}

type builder = Builder

func NewBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, ck ConfKind) (Builder, error) {
	version := helper.ExtractVersion(info.Image)
	name := name(ck)
	var promTLS *flowslatest.CertificateReference
	switch desired.Processor.Metrics.Server.TLS.Type {
	case flowslatest.ServerTLSProvided:
		promTLS = desired.Processor.Metrics.Server.TLS.Provided
		if promTLS == nil {
			return builder{}, fmt.Errorf("processor tls configuration set to provided but none is provided")
		}
	case flowslatest.ServerTLSAuto:
		promTLS = &flowslatest.CertificateReference{
			Type:     "secret",
			Name:     promServiceName(ck),
			CertFile: "tls.crt",
			CertKey:  "tls.key",
		}
	case flowslatest.ServerTLSDisabled:
		// nothing to do there
	}
	return builder{
		info: info,
		labels: map[string]string{
			appLabel:  name,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			appLabel: name,
		},
		desired:     desired,
		flowMetrics: flowMetrics,
		confKind:    ck,
		promTLS:     promTLS,
		loki:        info.Loki,
	}, nil
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
func (b *builder) Pipeline() *PipelineBuilder { return b.pipeline }

func (b *builder) NewGRPCPipeline() PipelineBuilder {
	return b.initPipeline(config.NewGRPCPipeline("grpc", api.IngestGRPCProto{
		Port: int(*helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced).Port),
	}))
}

func (b *builder) NewKafkaPipeline() PipelineBuilder {
	decoder := api.Decoder{Type: "protobuf"}
	return b.initPipeline(config.NewKafkaPipeline("kafka-read", api.IngestKafka{
		Brokers:           []string{b.desired.Kafka.Address},
		Topic:             b.desired.Kafka.Topic,
		GroupId:           b.name(), // Without groupid, each message is delivered to each consumers
		Decoder:           decoder,
		TLS:               getKafkaTLS(&b.desired.Kafka.TLS, "kafka-cert", &b.volumes),
		SASL:              getKafkaSASL(&b.desired.Kafka.SASL, "kafka-ingest", &b.volumes),
		PullQueueCapacity: b.desired.Processor.KafkaConsumerQueueCapacity,
		PullMaxBytes:      b.desired.Processor.KafkaConsumerBatchSize,
	}))
}

func (b *builder) NewInProcessPipeline() PipelineBuilder {
	return b.initPipeline(config.NewPresetIngesterPipeline())
}

func (b *builder) initPipeline(ingest config.PipelineBuilderStage) PipelineBuilder {
	pipeline := newPipelineBuilder(b.desired, b.flowMetrics, b.info.Loki, b.info.ClusterID, &b.volumes, &ingest)
	b.pipeline = &pipeline
	return pipeline
}

func (b *builder) overrideApp(app string) {
	b.labels[appLabel] = app
	b.selector[appLabel] = app
}

func (b *builder) podTemplate(hasHostPort, hostNetwork bool, annotations map[string]string) corev1.PodTemplateSpec {
	debugConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	var ports []corev1.ContainerPort
	var tolerations []corev1.Toleration
	if hasHostPort {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName,
			HostPort:      *debugConfig.Port,
			ContainerPort: *debugConfig.Port,
			Protocol:      corev1.ProtocolTCP,
		}}
		// This allows deploying an instance in the master node, the same technique used in the
		// companion ovnkube-node daemonset definition
		tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: *debugConfig.HealthPort,
	})

	ports = append(ports, corev1.ContainerPort{
		Name:          prometheusServiceName,
		ContainerPort: b.desired.Processor.Metrics.Server.Port,
	})

	if debugConfig.ProfilePort != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: *debugConfig.ProfilePort,
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
	debugConfig = helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(debugConfig.Env) {
		envs = append(envs, corev1.EnvVar{Name: pair[0], Value: pair[1]})
	}
	envs = append(envs, constants.EnvNoHTTP2)

	container := corev1.Container{
		Name:            constants.FLPName,
		Image:           b.info.Image,
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
		SecurityContext: helper.ContainerDefaultSecurityContext(),
	}
	if *debugConfig.EnableKubeProbes {
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

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) ConfigMap() (*corev1.ConfigMap, string, error) {
	configStr, err := b.GetJSONConfig()
	if err != nil {
		return nil, "", err
	}

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

func (b *builder) GetJSONConfig() (string, error) {
	metricsSettings := config.MetricsSettings{
		PromConnectionInfo: api.PromConnectionInfo{
			Port: int(b.desired.Processor.Metrics.Server.Port),
		},
		Prefix:  "netobserv_",
		NoPanic: true,
	}
	if b.desired.Processor.Metrics.Server.TLS.Type != flowslatest.ServerTLSDisabled {
		cert, key := b.volumes.AddCertificate(b.promTLS, "prom-certs")
		if cert != "" && key != "" {
			metricsSettings.TLS = &api.PromTLSConf{
				CertPath: cert,
				KeyPath:  key,
			}
		}
	}
	debugConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	config := map[string]interface{}{
		"log-level": b.desired.Processor.LogLevel,
		"health": map[string]interface{}{
			"port": *debugConfig.HealthPort,
		},
		"pipeline":        b.pipeline.GetStages(),
		"parameters":      b.pipeline.GetStageParams(),
		"metricsSettings": metricsSettings,
	}
	if debugConfig.ProfilePort != nil {
		config["profile"] = map[string]interface{}{
			"port": *debugConfig.ProfilePort,
		}
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bs), nil
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
				TargetPort: intstr.FromInt32(b.desired.Processor.Metrics.Server.Port),
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
				appLabel: b.name(),
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
				appLabel: b.name(),
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
				InsecureSkipVerify: b.desired.Processor.Metrics.Server.TLS.InsecureSkipVerify,
			},
		}
		if !b.desired.Processor.Metrics.Server.TLS.InsecureSkipVerify &&
			b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile != nil &&
			b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile.File != "" {
			flpServiceMonitorObject.Spec.Endpoints[0].TLSConfig.SafeTLSConfig.CA = helper.GetSecretOrConfigMap(b.desired.Processor.Metrics.Server.TLS.ProvidedCaFile)
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
	d := monitoringv1.Duration("10m")

	// Not receiving flows
	if shouldAddAlert(flowslatest.AlertNoFlows, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: string(flowslatest.AlertNoFlows),
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
				"summary":     "NetObserv flowlogs-pipeline is not receiving any flow",
			},
			Expr: intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m])) == 0"),
			For:  &d,
			Labels: map[string]string{
				"severity": "warning",
				appLabel:   "netobserv",
			},
		})
	}

	// Flows getting dropped by loki library
	if shouldAddAlert(flowslatest.AlertLokiError, b.desired.Processor.Metrics.DisableAlerts) {
		rules = append(rules, monitoringv1.Rule{
			Alert: string(flowslatest.AlertLokiError),
			Annotations: map[string]string{
				"description": "NetObserv flowlogs-pipeline is dropping flows because of loki errors, loki may be down or having issues ingesting every flows. Please check loki and flowlogs-pipeline logs.",
				"summary":     "NetObserv flowlogs-pipeline is dropping flows because of loki errors",
			},
			Expr: intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
			For:  &d,
			Labels: map[string]string{
				"severity": "warning",
				appLabel:   "netobserv",
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

func buildClusterRoleIngester(useOpenShiftSCC bool) *rbacv1.ClusterRole {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name(ConfKafkaIngester),
		},
		Rules: []rbacv1.PolicyRule{},
	}
	if useOpenShiftSCC {
		cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"security.openshift.io"},
			Verbs:         []string{"use"},
			Resources:     []string{"securitycontextconstraints"},
			ResourceNames: []string{"hostnetwork"},
		})
	}
	return &cr
}
