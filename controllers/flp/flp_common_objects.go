package flp

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/volumes"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
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
	info            *reconcilers.Instance
	labels          map[string]string
	selector        map[string]string
	desired         *flowslatest.FlowCollectorSpec
	flowMetrics     *metricslatest.FlowMetricList
	detectedSubnets []flowslatest.SubnetLabel
	promTLS         *flowslatest.CertificateReference
	confKind        ConfKind
	volumes         volumes.Builder
	loki            *helper.LokiConfig
	pipeline        *PipelineBuilder
	isDownstream    bool
}

type builder = Builder

func NewBuilder(info *reconcilers.Instance, desired *flowslatest.FlowCollectorSpec, flowMetrics *metricslatest.FlowMetricList, detectedSubnets []flowslatest.SubnetLabel, ck ConfKind) (Builder, error) {
	version := helper.ExtractVersion(info.Images[constants.ControllerBaseImageIndex])
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
			"app":     name,
			"version": helper.MaxLabelLength(version),
		},
		selector: map[string]string{
			"app": name,
		},
		desired:         desired,
		flowMetrics:     flowMetrics,
		detectedSubnets: detectedSubnets,
		confKind:        ck,
		promTLS:         promTLS,
		loki:            info.Loki,
		isDownstream:    info.IsDownstream,
	}, nil
}

func name(ck ConfKind) string                   { return constants.FLPName + FlpConfSuffix[ck] }
func RoleBindingName(ck ConfKind) string        { return name(ck) + "-role" }
func RoleBindingMonoName(ck ConfKind) string    { return name(ck) + "-role-mono" }
func promServiceName(ck ConfKind) string        { return name(ck) + "-prom" }
func staticConfigMapName(ck ConfKind) string    { return name(ck) + "-config" }
func dynamicConfigMapName(ck ConfKind) string   { return name(ck) + "-config-dynamic" }
func serviceMonitorName(ck ConfKind) string     { return name(ck) + "-monitor" }
func prometheusRuleName(ck ConfKind) string     { return name(ck) + "-alert" }
func (b *builder) name() string                 { return name(b.confKind) }
func (b *builder) promServiceName() string      { return promServiceName(b.confKind) }
func (b *builder) staticConfigMapName() string  { return staticConfigMapName(b.confKind) }
func (b *builder) dynamicConfigMapName() string { return dynamicConfigMapName(b.confKind) }
func (b *builder) serviceMonitorName() string   { return serviceMonitorName(b.confKind) }
func (b *builder) prometheusRuleName() string   { return prometheusRuleName(b.confKind) }
func (b *builder) Pipeline() *PipelineBuilder   { return b.pipeline }

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
		GroupID:           b.name(), // Without groupid, each message is delivered to each consumers
		Decoder:           decoder,
		TLS:               getClientTLS(&b.desired.Kafka.TLS, "kafka-cert", &b.volumes),
		SASL:              getSASL(&b.desired.Kafka.SASL, "kafka-ingest", &b.volumes),
		PullQueueCapacity: b.desired.Processor.KafkaConsumerQueueCapacity,
		PullMaxBytes:      b.desired.Processor.KafkaConsumerBatchSize,
	}))
}

func (b *builder) initPipeline(ingest config.PipelineBuilderStage) PipelineBuilder {
	pipeline := newPipelineBuilder(b.desired, b.flowMetrics, b.detectedSubnets, b.info.Loki, b.info.ClusterInfo.ID, &b.volumes, &ingest)
	b.pipeline = &pipeline
	return pipeline
}

func (b *builder) podTemplate(hasHostPort, hostNetwork bool, annotations map[string]string) corev1.PodTemplateSpec {
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	var ports []corev1.ContainerPort
	if hasHostPort {
		ports = []corev1.ContainerPort{{
			Name:          constants.FLPPortName,
			HostPort:      *advancedConfig.Port,
			ContainerPort: *advancedConfig.Port,
			Protocol:      corev1.ProtocolTCP,
		}}
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          healthServiceName,
		ContainerPort: *advancedConfig.HealthPort,
	})
	ports = append(ports, corev1.ContainerPort{
		Name:          prometheusServiceName,
		ContainerPort: helper.GetFlowCollectorMetricsPort(b.desired),
	})

	if advancedConfig.ProfilePort != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          profilePortName,
			ContainerPort: *advancedConfig.ProfilePort,
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
					Name: b.staticConfigMapName(),
				},
			},
		},
	}})

	var envs []corev1.EnvVar
	// we need to sort env map to keep idempotency,
	// as equal maps could be iterated in different order
	for _, pair := range helper.KeySorted(advancedConfig.Env) {
		envs = append(envs, corev1.EnvVar{Name: pair[0], Value: pair[1]})
	}
	envs = append(envs, constants.EnvNoHTTP2)

	container := corev1.Container{
		Name:            constants.FLPName,
		Image:           b.info.Images[constants.ControllerBaseImageIndex],
		ImagePullPolicy: corev1.PullPolicy(b.desired.Processor.ImagePullPolicy),
		Args:            []string{fmt.Sprintf(`--config=%s/%s`, configPath, configFile)},
		Resources:       *b.desired.Processor.Resources.DeepCopy(),
		VolumeMounts:    volumeMounts,
		Ports:           ports,
		Env:             envs,
		SecurityContext: helper.ContainerDefaultSecurityContext(),
	}
	if *advancedConfig.EnableKubeProbes {
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
	annotations["prometheus.io/scrape_port"] = fmt.Sprint(helper.GetFlowCollectorMetricsPort(b.desired))
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      b.labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Volumes:            volumes,
			Containers:         []corev1.Container{container},
			ServiceAccountName: b.name(),
			HostNetwork:        hostNetwork,
			DNSPolicy:          dnsPolicy,
			NodeSelector:       advancedConfig.Scheduling.NodeSelector,
			Tolerations:        advancedConfig.Scheduling.Tolerations,
			Affinity:           advancedConfig.Scheduling.Affinity,
			PriorityClassName:  advancedConfig.Scheduling.PriorityClassName,
		},
	}
}

// returns a configmap with a digest of its configuration contents, which will be used to
// detect any configuration change
func (b *builder) StaticConfigMap() (*corev1.ConfigMap, string, error) {
	configStr, err := b.GetStaticJSONConfig()
	if err != nil {
		return nil, "", err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.staticConfigMapName(),
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

func (b *builder) DynamicConfigMap() (*corev1.ConfigMap, error) {
	configStr, err := b.GetDynamicJSONConfig()
	if err != nil {
		return nil, err
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.dynamicConfigMapName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Data: map[string]string{
			configFile: configStr,
		},
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(configStr))
	return &configMap, nil
}

func (b *builder) GetStaticJSONConfig() (string, error) {
	metricsSettings := config.MetricsSettings{
		PromConnectionInfo: api.PromConnectionInfo{
			Port: int(helper.GetFlowCollectorMetricsPort(b.desired)),
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
	advancedConfig := helper.GetAdvancedProcessorConfig(b.desired.Processor.Advanced)
	config := map[string]interface{}{
		"log-level": b.desired.Processor.LogLevel,
		"health": map[string]interface{}{
			"port": *advancedConfig.HealthPort,
		},
		"pipeline":        b.pipeline.GetStages(),
		"parameters":      b.pipeline.GetStaticStageParams(),
		"metricsSettings": metricsSettings,
		"dynamicParameters": config.DynamicParameters{
			Namespace: b.info.Namespace,
			Name:      b.dynamicConfigMapName(),
			FileName:  configFile,
		},
	}
	if advancedConfig.ProfilePort != nil {
		config["profile"] = map[string]interface{}{
			"port": *advancedConfig.ProfilePort,
		}
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (b *builder) GetDynamicJSONConfig() (string, error) {
	config := map[string]interface{}{
		"parameters": b.pipeline.GetDynamicStageParams(),
	}

	bs, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(bs), nil

}
func (b *builder) promService() *corev1.Service {
	port := helper.GetFlowCollectorMetricsPort(b.desired)
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
				Port:     port,
				Protocol: corev1.ProtocolTCP,
				// Some Kubernetes versions might automatically set TargetPort to Port. We need to
				// explicitly set it here so the reconcile loop verifies that the owned service
				// is equal as the desired service
				TargetPort: intstr.FromInt32(port),
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

func (b *builder) serviceMonitor() *monitoringv1.ServiceMonitor {
	serverName := fmt.Sprintf("%s.%s.svc", b.promServiceName(), b.info.Namespace)
	scheme, smTLS := helper.GetServiceMonitorTLSConfig(&b.desired.Processor.Metrics.Server.TLS, serverName, b.isDownstream)
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.serviceMonitorName(),
			Namespace: b.info.Namespace,
			Labels:    b.labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:        prometheusServiceName,
					Interval:    "15s",
					Scheme:      scheme,
					TLSConfig:   smTLS,
					HonorLabels: true,
					MetricRelabelConfigs: []monitoringv1.RelabelConfig{
						{
							SourceLabels: []monitoringv1.LabelName{"__name__", "DstK8S_Namespace"},
							Separator:    ptr.To("@"),
							Regex:        "netobserv_(workload|namespace)_ingress_.*@(.*)",
							Replacement:  ptr.To("${2}"),
							TargetLabel:  "namespace",
							Action:       "replace",
						},
						{
							SourceLabels: []monitoringv1.LabelName{"__name__", "SrcK8S_Namespace"},
							Separator:    ptr.To("@"),
							Regex:        "netobserv_(workload|namespace)_(egress|flows|drop|rtt|dns)_.*@(.*)",
							Replacement:  ptr.To("${3}"),
							TargetLabel:  "namespace",
							Action:       "replace",
						},
					},
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
				"app":      "netobserv",
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
