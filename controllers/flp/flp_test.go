/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flp

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/cluster"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

var rs = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	},
}
var image = map[reconcilers.ImageRef]string{reconcilers.MainImage: "quay.io/netobserv/flowlogs-pipeline:dev"}
var image2 = map[reconcilers.ImageRef]string{reconcilers.MainImage: "quay.io/netobserv/flowlogs-pipeline:dev2"}
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
			Resources:       rs,
			Metrics: flowslatest.FLPMetrics{
				Server: flowslatest.MetricsServerConfig{
					Port: ptr.To(int32(9090)),
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

func monoBuilder(ns string, cfg *flowslatest.FlowCollectorSpec) monolithBuilder {
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: ns, Loki: &loki, ClusterInfo: &cluster.Info{}}
	b, _ := newMonolithBuilder(info.NewInstance(image, status.Instance{}), cfg, &metricslatest.FlowMetricList{}, nil)
	return b
}

func transfBuilder(ns string, cfg *flowslatest.FlowCollectorSpec) transfoBuilder {
	loki := helper.NewLokiConfig(&cfg.Loki, "any")
	info := reconcilers.Common{Namespace: ns, Loki: &loki, ClusterInfo: &cluster.Info{}}
	b, _ := newTransfoBuilder(info.NewInstance(image, status.Instance{}), cfg, &metricslatest.FlowMetricList{}, nil)
	return b
}

func annotate(digest string) map[string]string {
	return map[string]string{
		constants.PodConfigurationDigest: digest,
	}
}

func TestDaemonSetNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := monoBuilder(ns, &cfg)
	_, digest, _, err := b.configMaps()
	assert.NoError(err)
	first := b.daemonSet(annotate(digest))

	// Check no change
	cfg = getConfig()
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	second := b.daemonSet(annotate(digest))

	report := helper.NewChangeReport("")
	assert.False(helper.PodChanged(&first.Spec.Template, &second.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "no change")
}

func TestDaemonSetChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := monoBuilder(ns, &cfg)
	_, digest, _, err := b.configMaps()
	assert.NoError(err)
	first := b.daemonSet(annotate(digest))

	// Check probes enabled change
	cfg.Processor.Advanced.EnableKubeProbes = ptr.To(true)
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	second := b.daemonSet(annotate(digest))

	report := helper.NewChangeReport("")
	assert.True(helper.PodChanged(&first.Spec.Template, &second.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "probe changed")

	// Check probes DON'T change infinitely (bc DeepEqual/Derivative checks won't work there)
	assert.NoError(err)
	secondBis := b.daemonSet(annotate(digest))
	secondBis.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
		FailureThreshold: 3,
		PeriodSeconds:    10,
		SuccessThreshold: 1,
		TimeoutSeconds:   5,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/live",
				Port:   intstr.FromString("health"),
				Scheme: "http",
			},
		},
	}
	report = helper.NewChangeReport("")
	assert.False(helper.PodChanged(&second.Spec.Template, &secondBis.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "no change")

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	third := b.daemonSet(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&second.Spec.Template, &third.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "config-digest")

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	fourth := b.daemonSet(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&third.Spec.Template, &fourth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "req/limit changed")

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	fifth := b.daemonSet(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&fourth.Spec.Template, &fifth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "req/limit changed")
	report = helper.NewChangeReport("")
	assert.False(helper.PodChanged(&third.Spec.Template, &fifth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "no change")

	// Check Loki config change
	cfg.Loki.Manual.TLS = flowslatest.ClientTLS{
		Enable: true,
		CACert: flowslatest.CertificateReference{
			Type:     "configmap",
			Name:     "loki-cert",
			CertFile: "ca.crt",
		},
	}
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	sixth := b.daemonSet(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&fifth.Spec.Template, &sixth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "config-digest")

	// Check volumes change
	cfg.Loki.Manual.TLS = flowslatest.ClientTLS{
		Enable: true,
		CACert: flowslatest.CertificateReference{
			Type:     "configmap",
			Name:     "loki-cert-2",
			CertFile: "ca.crt",
		},
	}
	b = monoBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	seventh := b.daemonSet(annotate(digest))

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&sixth.Spec.Template, &seventh.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "Volumes changed")
}

func TestDeploymentNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := transfBuilder(ns, &cfg)
	_, digest, _, err := b.configMaps()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check no change
	cfg = getConfig()
	b = transfBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
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
	_, digest, _, err := b.configMaps()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check probes enabled change
	cfg.Processor.Advanced.EnableKubeProbes = ptr.To(true)
	b = transfBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
	assert.NoError(err)
	second := b.deployment(annotate(digest))

	report := helper.NewChangeReport("")
	checkChanged := func(old, newd *appsv1.Deployment, spec flowslatest.FlowCollectorSpec) bool {
		return helper.DeploymentChanged(old, newd, constants.FLPName, !helper.HPAEnabled(&spec.Processor.KafkaConsumerAutoscaler), *spec.Processor.KafkaConsumerReplicas, &report)
	}

	assert.True(checkChanged(first, second, cfg))
	assert.Contains(report.String(), "probe changed")

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = transfBuilder(ns, &cfg)
	_, digest, _, err = b.configMaps()
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
	_, digest, _, err = b.configMaps()
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
	_, digest, _, err = b.configMaps()
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
	_, digest, _, err = b.configMaps()
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
	_, digest, _, err := b.configMaps()
	assert.NoError(err)
	first := b.deployment(annotate(digest))

	// Check replicas changed (need to copy flp, as Spec.Replicas stores a pointer)
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = ptr.To(int32(5))
	b = transfBuilder(ns, &cfg2)
	_, digest, _, err = b.configMaps()
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
	b := monoBuilder(ns, &cfg)
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
	b := monoBuilder(ns, &cfg)
	first := b.promService()

	// Check port changed
	cfg.Processor.Metrics.Server.Port = ptr.To(int32(9999))
	b = monoBuilder(ns, &cfg)
	second := b.promService()

	report := helper.NewChangeReport("")
	assert.True(helper.ServiceChanged(first, second, &report))
	assert.Contains(report.String(), "Service spec changed")

	// Make sure non-service settings doesn't trigger service update
	cfg.Processor.LogLevel = "error"
	b = monoBuilder(ns, &cfg)
	third := b.promService()

	report = helper.NewChangeReport("")
	assert.False(helper.ServiceChanged(second, third, &report))
	assert.Contains(report.String(), "no change")

	// Check annotations change
	cfg.Processor.LogLevel = "error"
	b = monoBuilder(ns, &cfg)
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
	b := monoBuilder(ns, &cfg)
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
	b := monoBuilder(ns, &cfg)
	first := b.serviceMonitor()

	// Check namespace change
	b = monoBuilder("namespace2", &cfg)
	second := b.serviceMonitor()

	report := helper.NewChangeReport("")
	assert.True(helper.ServiceMonitorChanged(first, second, &report))
	assert.Contains(report.String(), "ServiceMonitor spec changed")

	// Check labels change
	info := reconcilers.Common{Namespace: "namespace2", ClusterInfo: &cluster.Info{}}
	b, _ = newMonolithBuilder(info.NewInstance(image2, status.Instance{}), &cfg, b.flowMetrics, nil)
	third := b.serviceMonitor()

	report = helper.NewChangeReport("")
	assert.True(helper.ServiceMonitorChanged(second, third, &report))
	assert.Contains(report.String(), "ServiceMonitor labels changed")

	// Check scheme changed
	b, _ = newMonolithBuilder(info.NewInstance(image2, status.Instance{}), &cfg, b.flowMetrics, nil)
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
	b := monoBuilder(ns, &cfg)
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
	b := monoBuilder("namespace", &cfg)
	first := b.prometheusRule()

	// Check enabled rule change
	cfg.Processor.Metrics.DisableAlerts = []flowslatest.FLPAlert{flowslatest.AlertNoFlows}
	b = monoBuilder("namespace", &cfg)
	second := b.prometheusRule()

	report := helper.NewChangeReport("")
	assert.True(helper.PrometheusRuleChanged(first, second, &report))
	assert.Contains(report.String(), "PrometheusRule spec changed")

	// Check labels change
	info := reconcilers.Common{Namespace: "namespace2", ClusterInfo: &cluster.Info{}}
	b, _ = newMonolithBuilder(info.NewInstance(image2, status.Instance{}), &cfg, b.flowMetrics, nil)
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
	b := monoBuilder(ns, &cfg)
	cm, digest, _, err := b.configMaps()
	assert.NoError(err)
	assert.NotEmpty(t, digest)

	assert.Equal("flowlogs-pipeline", cm.Labels["app"])

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded config.ConfigFileStruct
	err = json.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal("trace", decoded.LogLevel)

	params := decoded.Parameters
	assert.Len(params, 5)
	assert.Equal(*cfg.Processor.Advanced.Port, int32(params[0].Ingest.GRPC.Port))

	lokiCfg := params[3].Write.Loki
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
		"FlowDirection",
		"UdnId",
		"_RecordType",
	}, lokiCfg.Labels)
	assert.Equal(`{app="netobserv-flowcollector"}`, fmt.Sprintf("%v", lokiCfg.StaticLabels))

	assert.Equal(*cfg.Processor.Metrics.Server.Port, int32(decoded.MetricsSettings.Port))
}

func TestConfigMapShouldDeserializeAsJSONWithLokiStack(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	cfg := getConfig()
	useLokiStack(&cfg)
	cfg.Agent.Type = flowslatest.AgentEBPF
	b := monoBuilder(ns, &cfg)
	cm, digest, _, err := b.configMaps()
	assert.NoError(err)
	assert.NotEmpty(t, digest)

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded config.ConfigFileStruct
	err = json.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal("trace", decoded.LogLevel)

	params := decoded.Parameters
	assert.Len(params, 5)

	lokiCfg := params[3].Write.Loki
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
		"FlowDirection",
		"UdnId",
		"_RecordType",
	}, lokiCfg.Labels)
	assert.Equal(`{app="netobserv-flowcollector"}`, fmt.Sprintf("%v", lokiCfg.StaticLabels))

	assert.Equal(*cfg.Processor.Metrics.Server.Port, int32(decoded.MetricsSettings.Port))
}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	// Equals specs
	autoScalerSpec, hpa := getAutoScalerSpecs()
	report := helper.NewChangeReport("")
	assert.False(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "no change")

	// Wrong max replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Max replicas changed")

	// Missing min replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Min replicas changed")

	// Missing metrics
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	report = helper.NewChangeReport("")
	assert.True(helper.AutoScalerChanged(&autoScalerSpec, hpa, &report))
	assert.Contains(report.String(), "Metrics changed")
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	info := reconcilers.Common{Namespace: "ns", ClusterInfo: &cluster.Info{}}
	builder, _ := newMonolithBuilder(info.NewInstance(image, status.Instance{}), &cfg, &metricslatest.FlowMetricList{}, nil)
	tBuilder, _ := newTransfoBuilder(info.NewInstance(image, status.Instance{}), &cfg, &metricslatest.FlowMetricList{}, nil)

	// Deployment
	depl := tBuilder.deployment(annotate("digest"))
	assert.Equal("flowlogs-pipeline-transformer", depl.Labels["app"])
	assert.Equal("flowlogs-pipeline-transformer", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// DaemonSet
	ds := builder.daemonSet(annotate("digest"))
	assert.Equal("flowlogs-pipeline", ds.Labels["app"])
	assert.Equal("flowlogs-pipeline", ds.Spec.Template.Labels["app"])
	assert.Equal("dev", ds.Labels["version"])
	assert.Equal("dev", ds.Spec.Template.Labels["version"])

	// Service
	svc := builder.promService()
	assert.Equal("flowlogs-pipeline", svc.Labels["app"])
	assert.Equal("flowlogs-pipeline", svc.Spec.Selector["app"])
	assert.Empty(svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])

	// ServiceMonitor
	smMono := builder.serviceMonitor()
	assert.Equal("flowlogs-pipeline-monitor", smMono.Name)
	assert.Equal("flowlogs-pipeline", smMono.Spec.Selector.MatchLabels["app"])
	smTrans := tBuilder.serviceMonitor()
	assert.Equal("flowlogs-pipeline-transformer-monitor", smTrans.Name)
	assert.Equal("flowlogs-pipeline-transformer", smTrans.Spec.Selector.MatchLabels["app"])
}
