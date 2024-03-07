package flp

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

var resources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	},
}
var image = "quay.io/netobserv/flowlogs-pipeline:dev"
var image2 = "quay.io/netobserv/flowlogs-pipeline:dev2"
var pullPolicy = corev1.PullIfNotPresent
var minReplicas = int32(1)
var maxReplicas = int32(5)
var targetCPU = int32(75)

const testNamespace = "flp"

func getConfig() flowslatest.FlowCollectorSpec {
	return flowslatest.FlowCollectorSpec{
		DeploymentModel: flowslatest.DeploymentModelDirect,
		Agent:           flowslatest.FlowCollectorAgent{Type: flowslatest.AgentEBPF},
		Processor: flowslatest.FlowCollectorFLP{
			ImagePullPolicy: string(pullPolicy),
			LogLevel:        "trace",
			Resources:       resources,
			Metrics: flowslatest.FLPMetrics{
				Server: flowslatest.MetricsServerConfig{
					Port: 9090,
					TLS: flowslatest.ServerTLS{
						Type: flowslatest.ServerTLSDisabled,
					},
				},
			},
			KafkaConsumerReplicas: ptr.To(int32(1)),
			KafkaConsumerAutoscaler: flowslatest.FlowCollectorHPA{
				Status:      flowslatest.HPAStatusEnabled,
				MinReplicas: &minReplicas,
				MaxReplicas: maxReplicas,
				Metrics: []ascv2.MetricSpec{{
					Type: ascv2.ResourceMetricSourceType,
					Resource: &ascv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: ascv2.MetricTarget{
							Type:               ascv2.UtilizationMetricType,
							AverageUtilization: &targetCPU,
						},
					},
				}},
			},
			LogTypes: &outputRecordTypes,
			Advanced: &flowslatest.AdvancedProcessorConfig{
				Port:       ptr.To(int32(2055)),
				HealthPort: ptr.To(int32(8080)),
			},
		},
		Loki: getLoki(),
		Kafka: flowslatest.FlowCollectorKafka{
			Address: "kafka",
			Topic:   "flp",
		},
	}
}

func getLoki() flowslatest.FlowCollectorLoki {
	return flowslatest.FlowCollectorLoki{
		Mode: flowslatest.LokiModeManual,
		Manual: flowslatest.LokiManualParams{
			IngesterURL: "http://loki:3100/",
		},
		Enable: ptr.To(true),
		WriteBatchWait: &metav1.Duration{
			Duration: 1,
		},
		WriteBatchSize: 102400,
		Advanced: &flowslatest.AdvancedLokiConfig{
			WriteMinBackoff: &metav1.Duration{
				Duration: 1,
			},
			WriteMaxBackoff: &metav1.Duration{
				Duration: 300,
			},
			WriteMaxRetries: ptr.To(int32(10)),
			StaticLabels:    map[string]string{"app": "netobserv-flowcollector"},
		},
	}
}

func useLokiStack(cfg *flowslatest.FlowCollectorSpec) {
	cfg.Loki.Mode = flowslatest.LokiModeLokiStack
	cfg.Loki.LokiStack = flowslatest.LokiStackRef{
		Name:      "lokistack",
		Namespace: "ls-namespace",
	}
}

func getConfigNoHPA() flowslatest.FlowCollectorSpec {
	cfg := getConfig()
	cfg.Processor.KafkaConsumerAutoscaler.Status = flowslatest.HPAStatusDisabled
	return cfg
}

func getAutoScalerSpecs() (ascv2.HorizontalPodAutoscaler, flowslatest.FlowCollectorHPA) {
	var autoScaler = ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: constants.FLPName,
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
			Metrics: []ascv2.MetricSpec{{
				Type: ascv2.ResourceMetricSourceType,
				Resource: &ascv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: ascv2.MetricTarget{
						Type:               ascv2.UtilizationMetricType,
						AverageUtilization: &targetCPU,
					},
				},
			}},
		},
	}

	return autoScaler, getConfig().Processor.KafkaConsumerAutoscaler
}

func inProcessForImage(img string, ns string, cfg *flowslatest.FlowCollectorSpec) *Builder {
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: ns, Loki: &loki}
	b, _ := newInProcessBuilder(info.NewInstance(img, status.Instance{}), constants.FLPName, cfg, &metricslatest.FlowMetricList{})
	return b
}

func inProcessBuilder(ns string, cfg *flowslatest.FlowCollectorSpec) *Builder {
	return inProcessForImage(image, ns, cfg)
}

func transfBuilder(ns string, cfg *flowslatest.FlowCollectorSpec) *Builder {
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: ns, Loki: &loki}
	b, _ := newKafkaConsumerBuilder(info.NewInstance(image, status.Instance{}), cfg, &metricslatest.FlowMetricList{})
	return b
}

func annotate(digest string) map[string]string {
	return map[string]string{
		constants.PodConfigurationDigest: digest,
	}
}

func TestDeploymentNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := transfBuilder(ns, &cfg)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check no change
	cfg = getConfig()
	b = transfBuilder(ns, &cfg)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(annotate(digest))

	report := helper.NewChangeReport("")
	assert.False(helper.DeploymentChanged(first, second, constants.FLPName, !helper.HPAEnabled(&cfg.Processor.KafkaConsumerAutoscaler), *cfg.Processor.KafkaConsumerReplicas, &report))
	assert.Contains(report.String(), "no change")
}

func TestDeploymentChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := transfBuilder(ns, &cfg)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check probes enabled change
	cfg.Processor.Advanced.EnableKubeProbes = ptr.To(true)
	b = transfBuilder(ns, &cfg)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(annotate(digest))

	report := helper.NewChangeReport("")
	checkChanged := func(old, new *appsv1.Deployment, spec flowslatest.FlowCollectorSpec) bool {
		return helper.DeploymentChanged(old, new, constants.FLPName, !helper.HPAEnabled(&spec.Processor.KafkaConsumerAutoscaler), *spec.Processor.KafkaConsumerReplicas, &report)
	}

	assert.True(checkChanged(first, second, cfg))
	assert.Contains(report.String(), "probe changed")

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = transfBuilder(ns, &cfg)
	_, digest, err = b.configMap()
	assert.NoError(err)
	third := b.deployment(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(checkChanged(second, third, cfg))
	assert.Contains(report.String(), "config-digest")

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = transfBuilder(ns, &cfg)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fourth := b.deployment(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(checkChanged(third, fourth, cfg))
	assert.Contains(report.String(), "req/limit changed")

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = transfBuilder(ns, &cfg)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fifth := b.deployment(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(checkChanged(fourth, fifth, cfg))
	assert.Contains(report.String(), "req/limit changed")
	report = helper.NewChangeReport("")
	assert.False(checkChanged(third, fifth, cfg))
	assert.Contains(report.String(), "no change")

	// Check replicas didn't change because HPA is used
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = ptr.To(int32(5))
	b = transfBuilder(ns, &cfg2)
	_, digest, err = b.configMap()
	assert.NoError(err)
	sixth := b.deployment(annotate(digest))

	report = helper.NewChangeReport("")
	assert.False(checkChanged(fifth, sixth, cfg2))
	assert.Contains(report.String(), "no change")
}

func TestDeploymentChangedReplicasNoHPA(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfigNoHPA()
	b := transfBuilder(ns, &cfg)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check replicas changed (need to copy flp, as Spec.Replicas stores a pointer)
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = ptr.To(int32(5))
	b = transfBuilder(ns, &cfg2)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(annotate(digest))

	report := helper.NewChangeReport("")
	assert.True(helper.DeploymentChanged(first, second, constants.FLPName, !helper.HPAEnabled(&cfg2.Processor.KafkaConsumerAutoscaler), *cfg2.Processor.KafkaConsumerReplicas, &report))
	assert.Contains(report.String(), "Replicas changed")
}

func TestServiceNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := inProcessBuilder(ns, &cfg)
	first := b.promService()

	// Check no change
	newService := first.DeepCopy()

	report := helper.NewChangeReport("")
	assert.False(helper.ServiceChanged(first, newService, &report))
	assert.Contains(report.String(), "no change")
}

func TestServiceChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := inProcessBuilder(ns, &cfg)
	first := b.promService()

	// Check port changed
	cfg.Processor.Metrics.Server.Port = 9999
	b = inProcessBuilder(ns, &cfg)
	second := b.promService()

	report := helper.NewChangeReport("")
	assert.True(helper.ServiceChanged(first, second, &report))
	assert.Contains(report.String(), "Service spec changed")

	// Make sure non-service settings doesn't trigger service update
	cfg.Processor.LogLevel = "error"
	b = inProcessBuilder(ns, &cfg)
	third := b.promService()

	report = helper.NewChangeReport("")
	assert.False(helper.ServiceChanged(second, third, &report))
	assert.Contains(report.String(), "no change")

	// Check annotations change
	cfg.Processor.LogLevel = "error"
	b = inProcessBuilder(ns, &cfg)
	fourth := b.promService()
	fourth.ObjectMeta.Annotations = map[string]string{
		"name": "value",
	}

	report = helper.NewChangeReport("")
	assert.True(helper.ServiceChanged(third, fourth, &report))
	assert.Contains(report.String(), "Service annotations changed")
}

func TestServiceMonitorNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := inProcessBuilder(ns, &cfg)
	first := b.serviceMonitor()

	// Check no change
	newServiceMonitor := first.DeepCopy()

	report := helper.NewChangeReport("")
	assert.False(helper.ServiceMonitorChanged(first, newServiceMonitor, &report))
	assert.Contains(report.String(), "no change")
}

func TestServiceMonitorChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := inProcessBuilder(ns, &cfg)
	first := b.serviceMonitor()

	// Check namespace change
	b = inProcessBuilder("namespace2", &cfg)
	second := b.serviceMonitor()

	report := helper.NewChangeReport("")
	assert.True(helper.ServiceMonitorChanged(first, second, &report))
	assert.Contains(report.String(), "ServiceMonitor spec changed")

	// Check labels change
	b = inProcessForImage(image2, "namespace2", &cfg)
	third := b.serviceMonitor()

	report = helper.NewChangeReport("")
	assert.True(helper.ServiceMonitorChanged(second, third, &report))
	assert.Contains(report.String(), "ServiceMonitor labels changed")

	// Check scheme changed
	b = inProcessForImage(image2, "namespace2", &cfg)
	fourth := b.serviceMonitor()
	fourth.Spec.Endpoints[0].Scheme = "https"

	report = helper.NewChangeReport("")
	assert.True(helper.ServiceMonitorChanged(third, fourth, &report))
	assert.Contains(report.String(), "ServiceMonitor spec changed")
}

func TestPrometheusRuleNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := inProcessBuilder(ns, &cfg)
	first := b.prometheusRule()

	// Check no change
	newServiceMonitor := first.DeepCopy()

	report := helper.NewChangeReport("")
	assert.False(helper.PrometheusRuleChanged(first, newServiceMonitor, &report))
	assert.Contains(report.String(), "no change")
}

func TestPrometheusRuleChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	cfg := getConfig()
	b := inProcessBuilder("namespace", &cfg)
	first := b.prometheusRule()

	// Check enabled rule change
	cfg.Processor.Metrics.DisableAlerts = []flowslatest.FLPAlert{flowslatest.AlertNoFlows}
	b = inProcessBuilder("namespace", &cfg)
	second := b.prometheusRule()

	report := helper.NewChangeReport("")
	assert.True(helper.PrometheusRuleChanged(first, second, &report))
	assert.Contains(report.String(), "PrometheusRule spec changed")

	// Check labels change
	b = inProcessForImage(image2, "namespace2", &cfg)
	third := b.prometheusRule()

	report = helper.NewChangeReport("")
	assert.True(helper.PrometheusRuleChanged(second, third, &report))
	assert.Contains(report.String(), "PrometheusRule labels changed")
}

func TestConfigMapShouldDeserializeAsJSONWithLokiManual(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	cfg := getConfig()
	loki := cfg.Loki
	b := inProcessBuilder(ns, &cfg)
	cm, digest, err := b.configMap()
	assert.NoError(err)
	assert.NotEmpty(t, digest)

	assert.Equal("dev", cm.Labels["version"])

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded config.ConfigFileStruct
	err = json.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal("trace", decoded.LogLevel)

	params := decoded.Parameters
	assert.Len(params, 5)

	lokiCfg := params[2].Write.Loki
	assert.Equal(loki.Manual.IngesterURL, lokiCfg.URL)
	assert.Equal(cfg.Loki.WriteBatchWait.Duration.String(), lokiCfg.BatchWait)
	assert.EqualValues(cfg.Loki.WriteBatchSize, lokiCfg.BatchSize)
	assert.Equal(cfg.Loki.Advanced.WriteMinBackoff.Duration.String(), lokiCfg.MinBackoff)
	assert.Equal(cfg.Loki.Advanced.WriteMaxBackoff.Duration.String(), lokiCfg.MaxBackoff)
	assert.EqualValues(*cfg.Loki.Advanced.WriteMaxRetries, lokiCfg.MaxRetries)
	assert.EqualValues([]string{
		"SrcK8S_Namespace",
		"SrcK8S_OwnerName",
		"SrcK8S_Type",
		"DstK8S_Namespace",
		"DstK8S_OwnerName",
		"DstK8S_Type",
		"K8S_FlowLayer",
		"_RecordType",
		"FlowDirection",
	}, lokiCfg.Labels)
	assert.Equal(`{app="netobserv-flowcollector"}`, fmt.Sprintf("%v", lokiCfg.StaticLabels))

	assert.Equal(cfg.Processor.Metrics.Server.Port, int32(decoded.MetricsSettings.Port))
}

func TestConfigMapShouldDeserializeAsJSONWithLokiStack(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	cfg := getConfig()
	useLokiStack(&cfg)
	cfg.Agent.Type = flowslatest.AgentEBPF
	b := inProcessBuilder(ns, &cfg)
	cm, digest, err := b.configMap()
	assert.NoError(err)
	assert.NotEmpty(t, digest)

	assert.Equal("dev", cm.Labels["version"])

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded config.ConfigFileStruct
	err = json.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal("trace", decoded.LogLevel)

	params := decoded.Parameters
	assert.Len(params, 5)

	lokiCfg := params[2].Write.Loki
	assert.Equal("https://lokistack-gateway-http.ls-namespace.svc:8080/api/logs/v1/network/", lokiCfg.URL)
	assert.Equal("network", lokiCfg.TenantID)
	assert.Equal("Bearer", lokiCfg.ClientConfig.Authorization.Type)
	assert.Equal("/var/run/secrets/tokens/flowlogs-pipeline", lokiCfg.ClientConfig.Authorization.CredentialsFile)
	assert.Equal(false, lokiCfg.ClientConfig.TLSConfig.InsecureSkipVerify)
	assert.Equal("/var/loki-certs-ca/service-ca.crt", lokiCfg.ClientConfig.TLSConfig.CAFile)
	assert.Equal("", lokiCfg.ClientConfig.TLSConfig.CertFile)
	assert.Equal("", lokiCfg.ClientConfig.TLSConfig.KeyFile)
	assert.Equal(cfg.Loki.WriteBatchWait.Duration.String(), lokiCfg.BatchWait)
	assert.EqualValues(cfg.Loki.WriteBatchSize, lokiCfg.BatchSize)
	assert.Equal(cfg.Loki.Advanced.WriteMinBackoff.Duration.String(), lokiCfg.MinBackoff)
	assert.Equal(cfg.Loki.Advanced.WriteMaxBackoff.Duration.String(), lokiCfg.MaxBackoff)
	assert.EqualValues(*cfg.Loki.Advanced.WriteMaxRetries, lokiCfg.MaxRetries)
	assert.EqualValues([]string{
		"SrcK8S_Namespace",
		"SrcK8S_OwnerName",
		"SrcK8S_Type",
		"DstK8S_Namespace",
		"DstK8S_OwnerName",
		"DstK8S_Type",
		"K8S_FlowLayer",
		"_RecordType",
		"FlowDirection",
	}, lokiCfg.Labels)
	assert.Equal(`{app="netobserv-flowcollector"}`, fmt.Sprintf("%v", lokiCfg.StaticLabels))

	assert.Equal(cfg.Processor.Metrics.Server.Port, int32(decoded.MetricsSettings.Port))
}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, hpa := getAutoScalerSpecs()
	report := helper.NewChangeReport("")
	assert.False(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "no change")

	//wrong max replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Max replicas changed")

	//missing min replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Min replicas changed")

	//missing metrics
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Metrics changed")
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	builder := inProcessBuilder("ns", &cfg)
	tBuilder := transfBuilder("ns", &cfg)

	// Deployment
	depl := tBuilder.deployment(annotate("digest"))
	assert.Equal("flowlogs-pipeline", depl.Labels["app"])
	assert.Equal("flowlogs-pipeline", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// Service
	svc := builder.promService()
	assert.Equal("flowlogs-pipeline", svc.Labels["app"])
	assert.Equal("flowlogs-pipeline", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])

	// ServiceMonitor
	smMono := builder.serviceMonitor()
	assert.Equal("flowlogs-pipeline-monitor", smMono.Name)
	assert.Equal("flowlogs-pipeline", smMono.Spec.Selector.MatchLabels["app"])
	smTrans := tBuilder.serviceMonitor()
	assert.Equal("flowlogs-pipeline-monitor", smTrans.Name)
	assert.Equal("flowlogs-pipeline", smTrans.Spec.Selector.MatchLabels["app"])
}

// This function validate that each stage has its matching parameter
func validatePipelineConfig(t *testing.T, cm *corev1.ConfigMap) (*config.ConfigFileStruct, string) {
	var cfs config.ConfigFileStruct
	err := json.Unmarshal([]byte(cm.Data[configFile]), &cfs)
	assert.NoError(t, err)

	for _, stage := range cfs.Pipeline {
		assert.NotEmpty(t, stage.Name)
		exist := false
		for _, parameter := range cfs.Parameters {
			if stage.Name == parameter.Name {
				exist = true
				break
			}
		}
		assert.True(t, exist, "stage params not found", stage.Name)
	}
	b, err := json.Marshal(cfs.Pipeline)
	assert.NoError(t, err)
	return &cfs, string(b)
}

func TestPipelineConfig(t *testing.T) {
	assert := assert.New(t)

	// Single config
	ns := "namespace"
	cfg := getConfig()
	cfg.Processor.LogLevel = "info"
	b := inProcessBuilder(ns, &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, cm)
	assert.Equal(
		`[{"name":"extract_conntrack","follows":"preset-ingester"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)

	// Kafka Transformer
	cfg.DeploymentModel = flowslatest.DeploymentModelKafka
	bt := transfBuilder(ns, &cfg)
	cm, _, err = bt.configMap()
	assert.NoError(err)
	_, pipeline = validatePipelineConfig(t, cm)
	assert.Equal(
		`[{"name":"kafka-read"},{"name":"extract_conntrack","follows":"kafka-read"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}

func TestPipelineTraceStage(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, cm)
	assert.Equal(
		`[{"name":"extract_conntrack","follows":"preset-ingester"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}

func getSortedMetricsNames(m []api.MetricsItem) []string {
	ret := []string{}
	for i := range m {
		ret = append(ret, m[i].Name)
	}
	sort.Strings(ret)
	return ret
}

func TestMergeMetricsConfiguration_Default(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, cm)
	names := getSortedMetricsNames(cfs.Parameters[4].Encode.Prom.Metrics)
	assert.Equal([]string{
		"namespace_flows_total",
		"node_ingress_bytes_total",
		"workload_ingress_bytes_total",
	}, names)
	assert.Equal("netobserv_", cfs.Parameters[4].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_DefaultWithFeatures(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Agent.EBPF.Privileged = true
	cfg.Agent.EBPF.Features = []flowslatest.AgentFeature{flowslatest.DNSTracking, flowslatest.FlowRTT, flowslatest.PacketDrop}

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, cm)
	names := getSortedMetricsNames(cfs.Parameters[4].Encode.Prom.Metrics)
	assert.Equal([]string{
		"namespace_dns_latency_seconds",
		"namespace_drop_packets_total",
		"namespace_flows_total",
		"namespace_rtt_seconds",
		"node_ingress_bytes_total",
		"workload_ingress_bytes_total",
	}, names)
	assert.Equal("netobserv_", cfs.Parameters[4].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_WithList(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.Metrics.IncludeList = &[]flowslatest.FLPMetric{"namespace_egress_bytes_total", "namespace_ingress_bytes_total"}

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, cm)
	names := getSortedMetricsNames(cfs.Parameters[4].Encode.Prom.Metrics)
	assert.Len(names, 2)
	assert.Equal("namespace_egress_bytes_total", names[0])
	assert.Equal("namespace_ingress_bytes_total", names[1])
	assert.Equal("netobserv_", cfs.Parameters[4].Encode.Prom.Prefix)
}

func TestMergeMetricsConfiguration_EmptyList(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.Metrics.IncludeList = &[]flowslatest.FLPMetric{}

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	cfs, _ := validatePipelineConfig(t, cm)
	assert.Len(cfs.Parameters, 4)
}

func TestPipelineWithExporter(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Exporters = append(cfg.Exporters, &flowslatest.FlowCollectorExporter{
		Type:  flowslatest.KafkaExporter,
		Kafka: flowslatest.FlowCollectorKafka{Address: "kafka-test", Topic: "topic-test"},
	})

	cfg.Exporters = append(cfg.Exporters, &flowslatest.FlowCollectorExporter{
		Type: flowslatest.IpfixExporter,
		IPFIX: flowslatest.FlowCollectorIPFIXReceiver{
			TargetHost: "ipfix-receiver-test",
			TargetPort: 9999,
			Transport:  "TCP",
		},
	})

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	cfs, pipeline := validatePipelineConfig(t, cm)
	assert.Equal(
		`[{"name":"extract_conntrack","follows":"preset-ingester"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"},{"name":"kafka-export-0","follows":"enrich"},{"name":"IPFIX-export-1","follows":"enrich"}]`,
		pipeline,
	)

	assert.Equal("kafka-test", cfs.Parameters[5].Encode.Kafka.Address)
	assert.Equal("topic-test", cfs.Parameters[5].Encode.Kafka.Topic)

	assert.Equal("ipfix-receiver-test", cfs.Parameters[6].Write.Ipfix.TargetHost)
	assert.Equal(9999, cfs.Parameters[6].Write.Ipfix.TargetPort)
	assert.Equal("tcp", cfs.Parameters[6].Write.Ipfix.Transport)
}

func TestPipelineWithoutLoki(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Loki.Enable = ptr.To(false)

	b := inProcessBuilder("namespace", &cfg)
	cm, _, err := b.configMap()
	assert.NoError(err)
	_, pipeline := validatePipelineConfig(t, cm)
	assert.Equal(
		`[{"name":"extract_conntrack","follows":"preset-ingester"},{"name":"enrich","follows":"extract_conntrack"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`,
		pipeline,
	)
}
