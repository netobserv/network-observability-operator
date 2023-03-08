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

package flowlogspipeline

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

var resources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	},
}
var image = "quay.io/netobserv/flowlogs-pipeline:dev"
var pullPolicy = corev1.PullIfNotPresent
var minReplicas = int32(1)
var maxReplicas = int32(5)
var targetCPU = int32(75)
var certWatcher = watchers.NewCertificatesWatcher()
var outputRecordTypes = flowslatest.OutputRecordAll

const testNamespace = "flp"

func getConfig() flowslatest.FlowCollectorSpec {
	return flowslatest.FlowCollectorSpec{
		DeploymentModel: flowslatest.DeploymentModelDirect,
		Agent:           flowslatest.FlowCollectorAgent{Type: flowslatest.AgentIPFIX},
		Processor: flowslatest.FlowCollectorFLP{
			Port:            2055,
			ImagePullPolicy: string(pullPolicy),
			LogLevel:        "trace",
			Resources:       resources,
			HealthPort:      8080,
			Metrics: flowslatest.FLPMetrics{
				Server: flowslatest.MetricsServerConfig{
					Port: 9090,
					TLS: flowslatest.ServerTLS{
						Type: flowslatest.ServerTLSDisabled,
					},
				},
			},
			KafkaConsumerReplicas: 1,
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
			OutputRecordTypes: &outputRecordTypes,
			ConnectionHeartbeatInterval: &metav1.Duration{
				Duration: conntrackHeartbeatInterval,
			},
			ConnectionEndTimeout: &metav1.Duration{
				Duration: conntrackEndTimeout,
			},
		},
		Loki: flowslatest.FlowCollectorLoki{
			URL: "http://loki:3100/",
			BatchWait: metav1.Duration{
				Duration: 1,
			},
			BatchSize: 102400,
			MinBackoff: metav1.Duration{
				Duration: 1,
			},
			MaxBackoff: metav1.Duration{
				Duration: 300,
			},
			MaxRetries:   10,
			StaticLabels: map[string]string{"app": "netobserv-flowcollector"},
		},
		Kafka: flowslatest.FlowCollectorKafka{
			Address: "kafka",
			Topic:   "flp",
		},
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

func TestDaemonSetNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.daemonSet(digest)

	// Check no change
	cfg = getConfig()
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.daemonSet(digest)

	report := helper.NewChangeReport("")
	assert.False(helper.PodChanged(&first.Spec.Template, &second.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "no change")
}

func TestDaemonSetChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.daemonSet(digest)

	// Check probes enabled change
	cfg.Processor.EnableKubeProbes = true
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.daemonSet(digest)

	report := helper.NewChangeReport("")
	assert.True(helper.PodChanged(&first.Spec.Template, &second.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "probe changed")

	// Check probes DON'T change infinitely (bc DeepEqual/Derivative checks won't work there)
	assert.NoError(err)
	secondBis := b.daemonSet(digest)
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
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	third := b.daemonSet(digest)

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&second.Spec.Template, &third.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "config-digest")

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fourth := b.daemonSet(digest)

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&third.Spec.Template, &fourth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "req/limit changed")

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fifth := b.daemonSet(digest)

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&fourth.Spec.Template, &fifth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "req/limit changed")
	report = helper.NewChangeReport("")
	assert.False(helper.PodChanged(&third.Spec.Template, &fifth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "no change")

	// Check Loki config change
	cfg.Loki.TLS = flowslatest.ClientTLS{
		Enable: true,
		CACert: flowslatest.CertificateReference{
			Type:     "configmap",
			Name:     "loki-cert",
			CertFile: "ca.crt",
		},
	}
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	sixth := b.daemonSet(digest)

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&fifth.Spec.Template, &sixth.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "config-digest")

	// Check volumes change
	cfg.Loki.TLS = flowslatest.ClientTLS{
		Enable: true,
		CACert: flowslatest.CertificateReference{
			Type:     "configmap",
			Name:     "loki-cert-2",
			CertFile: "ca.crt",
		},
	}
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	seventh := b.daemonSet(digest)

	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&sixth.Spec.Template, &seventh.Spec.Template, constants.FLPName, &report))
	assert.Contains(report.String(), "Volumes changed")
}

func TestDeploymentNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check no change
	cfg = getConfig()
	b = newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(digest)

	report := helper.NewChangeReport("")
	assert.False(helper.DeploymentChanged(first, second, constants.FLPName, helper.HPADisabled(&cfg.Processor.KafkaConsumerAutoscaler), cfg.Processor.KafkaConsumerReplicas, &report))
	assert.Contains(report.String(), "no change")
}

func TestDeploymentChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check probes enabled change
	cfg.Processor.EnableKubeProbes = true
	b = newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(digest)

	report := helper.NewChangeReport("")
	checkChanged := func(old, new *appsv1.Deployment, spec flowslatest.FlowCollectorSpec) bool {
		return helper.DeploymentChanged(old, new, constants.FLPName, helper.HPADisabled(&spec.Processor.KafkaConsumerAutoscaler), spec.Processor.KafkaConsumerReplicas, &report)
	}

	assert.True(checkChanged(first, second, cfg))
	assert.Contains(report.String(), "probe changed")

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	third := b.deployment(digest)

	report = helper.NewChangeReport("")
	assert.True(checkChanged(second, third, cfg))
	assert.Contains(report.String(), "config-digest")

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fourth := b.deployment(digest)

	report = helper.NewChangeReport("")
	assert.True(checkChanged(third, fourth, cfg))
	assert.Contains(report.String(), "req/limit changed")

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fifth := b.deployment(digest)

	report = helper.NewChangeReport("")
	assert.True(checkChanged(fourth, fifth, cfg))
	assert.Contains(report.String(), "req/limit changed")
	report = helper.NewChangeReport("")
	assert.False(checkChanged(third, fifth, cfg))
	assert.Contains(report.String(), "no change")

	// Check replicas didn't change because HPA is used
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = 5
	b = newTransfoBuilder(ns, image, &cfg2, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	sixth := b.deployment(digest)

	report = helper.NewChangeReport("")
	assert.False(checkChanged(fifth, sixth, cfg2))
	assert.Contains(report.String(), "no change")
}

func TestDeploymentChangedReplicasNoHPA(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfigNoHPA()
	b := newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check replicas changed (need to copy flp, as Spec.Replicas stores a pointer)
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = 5
	b = newTransfoBuilder(ns, image, &cfg2, true, &certWatcher)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(digest)

	report := helper.NewChangeReport("")
	assert.True(helper.DeploymentChanged(first, second, constants.FLPName, helper.HPADisabled(&cfg2.Processor.KafkaConsumerAutoscaler), cfg2.Processor.KafkaConsumerReplicas, &report))
	assert.Contains(report.String(), "Replicas changed")
}

func TestServiceNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.newPromService()

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
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.newPromService()

	// Check port changed
	cfg.Processor.Metrics.Server.Port = 9999
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	second := b.fromPromService(first)

	report := helper.NewChangeReport("")
	assert.True(helper.ServiceChanged(first, second, &report))
	assert.Contains(report.String(), "Service spec changed")

	// Make sure non-service settings doesn't trigger service update
	cfg.Processor.LogLevel = "error"
	b = newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	third := b.fromPromService(first)

	report = helper.NewChangeReport("")
	assert.False(helper.ServiceChanged(second, third, &report))
	assert.Contains(report.String(), "no change")
}

func TestServiceMonitorNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.generic.serviceMonitor()

	// Check no change
	newServiceMonitor := first.DeepCopy()

	assert.False(helper.ServiceMonitorChanged(first, newServiceMonitor))
}

func TestServiceMonitorChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.generic.serviceMonitor()

	// Check namespace change
	cfg.Processor.Metrics.Server.Port = 9999
	b = newMonolithBuilder("namespace2", image, &cfg, true, &certWatcher)
	second := b.generic.serviceMonitor()

	assert.True(helper.ServiceMonitorChanged(first, second))
}

func TestPrometheusRuleNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.generic.prometheusRule()

	// Check no change
	newServiceMonitor := first.DeepCopy()

	assert.False(helper.PrometheusRuleChanged(first, newServiceMonitor))
}

func TestPrometheusRuleChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	first := b.generic.prometheusRule()

	// Check namespace change
	cfg.Processor.Metrics.Server.Port = 9999
	b = newMonolithBuilder("namespace2", image, &cfg, true, &certWatcher)
	second := b.generic.prometheusRule()

	assert.True(helper.PrometheusRuleChanged(first, second))
}

func TestConfigMapShouldDeserializeAsJSON(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	cfg := getConfig()
	loki := cfg.Loki
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
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
	assert.Len(params, 6)
	assert.Equal(cfg.Processor.Port, int32(params[0].Ingest.Collector.Port))

	lokiCfg := params[3].Write.Loki
	assert.Equal(loki.URL, lokiCfg.URL)
	assert.Equal(loki.BatchWait.Duration.String(), lokiCfg.BatchWait)
	assert.Equal(loki.MinBackoff.Duration.String(), lokiCfg.MinBackoff)
	assert.Equal(loki.MaxBackoff.Duration.String(), lokiCfg.MaxBackoff)
	assert.EqualValues(loki.MaxRetries, lokiCfg.MaxRetries)
	assert.EqualValues(loki.BatchSize, lokiCfg.BatchSize)
	assert.EqualValues([]string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection", "_RecordType"}, lokiCfg.Labels)
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
	builder := newMonolithBuilder("ns", image, &cfg, true, &certWatcher)
	tBuilder := newTransfoBuilder("ns", image, &cfg, true, &certWatcher)
	iBuilder := newIngestBuilder("ns", image, &cfg, true, &certWatcher)

	// Deployment
	depl := tBuilder.deployment("digest")
	assert.Equal("flowlogs-pipeline-transformer", depl.Labels["app"])
	assert.Equal("flowlogs-pipeline-transformer", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// DaemonSet
	ds := builder.daemonSet("digest")
	assert.Equal("flowlogs-pipeline", ds.Labels["app"])
	assert.Equal("flowlogs-pipeline", ds.Spec.Template.Labels["app"])
	assert.Equal("dev", ds.Labels["version"])
	assert.Equal("dev", ds.Spec.Template.Labels["version"])

	// DaemonSet (ingester)
	ds2 := iBuilder.daemonSet("digest")
	assert.Equal("flowlogs-pipeline-ingester", ds2.Labels["app"])
	assert.Equal("flowlogs-pipeline-ingester", ds2.Spec.Template.Labels["app"])
	assert.Equal("dev", ds2.Labels["version"])
	assert.Equal("dev", ds2.Spec.Template.Labels["version"])

	// Service
	svc := builder.newPromService()
	assert.Equal("flowlogs-pipeline", svc.Labels["app"])
	assert.Equal("flowlogs-pipeline", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])

	// ServiceMonitor
	smMono := builder.generic.serviceMonitor()
	assert.Equal("flowlogs-pipeline-monitor", smMono.Name)
	assert.Equal("flowlogs-pipeline", smMono.Spec.Selector.MatchLabels["app"])
	smTrans := tBuilder.generic.serviceMonitor()
	assert.Equal("flowlogs-pipeline-transformer-monitor", smTrans.Name)
	assert.Equal("flowlogs-pipeline-transformer", smTrans.Spec.Selector.MatchLabels["app"])
	smIng := iBuilder.generic.serviceMonitor()
	assert.Equal("flowlogs-pipeline-ingester-monitor", smIng.Name)
	assert.Equal("flowlogs-pipeline-ingester", smIng.Spec.Selector.MatchLabels["app"])
}

// This function validate that each stage has its matching parameter
func validatePipelineConfig(stages []config.Stage, parameters []config.StageParam) bool {
	for _, stage := range stages {
		if stage.Name == "" {
			return false
		}
		exist := false
		for _, parameter := range parameters {
			if stage.Name == parameter.Name {
				exist = true
				break
			}
		}
		if !exist {
			return false
		}
	}
	return true
}

func TestPipelineConfig(t *testing.T) {
	assert := assert.New(t)

	// Single config
	ns := "namespace"
	cfg := getConfig()
	cfg.Processor.LogLevel = "info"
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"extract_conntrack","follows":"ipfix"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))

	// Kafka Ingester
	cfg.DeploymentModel = flowslatest.DeploymentModelKafka
	bi := newIngestBuilder(ns, image, &cfg, true, &certWatcher)
	stages, parameters, err = bi.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ = json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"kafka-write","follows":"ipfix"}]`, string(jsonStages))

	// Kafka Transformer
	bt := newTransfoBuilder(ns, image, &cfg, true, &certWatcher)
	stages, parameters, err = bt.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ = json.Marshal(stages)
	assert.Equal(`[{"name":"kafka-read"},{"name":"extract_conntrack","follows":"kafka-read"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
}

func TestPipelineConfigDropUnused(t *testing.T) {
	assert := assert.New(t)

	// Single config
	ns := "namespace"
	cfg := getConfig()
	cfg.Processor.LogLevel = "info"
	cfg.Processor.DropUnusedFields = true
	b := newMonolithBuilder(ns, image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"filter","follows":"ipfix"},{"name":"extract_conntrack","follows":"filter"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))

	jsonParams, _ := json.Marshal(parameters[1].Transform.Filter)
	assert.Contains(string(jsonParams), `{"input":"CustomBytes1","type":"remove_field"}`)
	assert.Contains(string(jsonParams), `{"input":"CustomInteger5","type":"remove_field"}`)
	assert.Contains(string(jsonParams), `{"input":"MPLS1Label","type":"remove_field"}`)
}

func TestPipelineTraceStage(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := newMonolithBuilder("namespace", image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"extract_conntrack","follows":"ipfix"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
}

func TestMergeMetricsConfigurationNoIgnore(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := newMonolithBuilder("namespace", image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"extract_conntrack","follows":"ipfix"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
	assert.Len(parameters[5].Encode.Prom.Metrics, 7)
	assert.Equal("namespace_flows_total", parameters[5].Encode.Prom.Metrics[0].Name)
	assert.Equal("node_egress_bytes_total", parameters[5].Encode.Prom.Metrics[1].Name)
	assert.Equal("node_ingress_bytes_total", parameters[5].Encode.Prom.Metrics[2].Name)
	assert.Equal("workload_egress_bytes_total", parameters[5].Encode.Prom.Metrics[3].Name)
	assert.Equal("workload_egress_packets_total", parameters[5].Encode.Prom.Metrics[4].Name)
	assert.Equal("workload_ingress_bytes_total", parameters[5].Encode.Prom.Metrics[5].Name)
	assert.Equal("workload_ingress_packets_total", parameters[5].Encode.Prom.Metrics[6].Name)
	assert.Equal("netobserv_", parameters[5].Encode.Prom.Prefix)
}

func TestMergeMetricsConfigurationWithIgnore(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.Metrics.IgnoreTags = []string{"nodes"}

	b := newMonolithBuilder("namespace", image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"extract_conntrack","follows":"ipfix"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
	assert.Len(parameters[5].Encode.Prom.Metrics, 5)
	assert.Equal("namespace_flows_total", parameters[5].Encode.Prom.Metrics[0].Name)
	assert.Equal("netobserv_", parameters[5].Encode.Prom.Prefix)
}

func TestPipelineWithExporter(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Exporters = append(cfg.Exporters, &flowslatest.FlowCollectorExporter{
		Type:  flowslatest.KafkaExporter,
		Kafka: flowslatest.FlowCollectorKafka{Address: "kafka-test", Topic: "topic-test"},
	})

	b := newMonolithBuilder("namespace", image, &cfg, true, &certWatcher)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"extract_conntrack","follows":"ipfix"},{"name":"enrich","follows":"extract_conntrack"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"},{"name":"kafka-export-0","follows":"enrich"}]`, string(jsonStages))

	assert.Equal("kafka-test", parameters[6].Encode.Kafka.Address)
	assert.Equal("topic-test", parameters[6].Encode.Kafka.Topic)
}
