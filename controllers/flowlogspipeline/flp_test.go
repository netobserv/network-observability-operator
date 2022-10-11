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
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
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

const testNamespace = "flp"

func getConfig() v1alpha1.FlowCollectorSpec {
	return v1alpha1.FlowCollectorSpec{
		DeploymentModel: v1alpha1.DeploymentModelDirect,
		Agent:           v1alpha1.FlowCollectorAgent{Type: v1alpha1.AgentIPFIX},
		Processor: v1alpha1.FlowCollectorFLP{
			Port:            2055,
			Image:           image,
			ImagePullPolicy: string(pullPolicy),
			LogLevel:        "trace",
			Resources:       resources,
			HealthPort:      8080,
			MetricsServer: v1alpha1.MetricsServerConfig{
				Port: 9090,
				TLS: v1alpha1.ServerTLS{
					Type: v1alpha1.ServerTLSDisabled,
				},
			},
			KafkaConsumerReplicas: 1,
			KafkaConsumerAutoscaler: &v1alpha1.FlowCollectorHPA{
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
		},
		Loki: v1alpha1.FlowCollectorLoki{
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
		Kafka: v1alpha1.FlowCollectorKafka{
			Address: "kafka",
			Topic:   "flp",
		},
	}
}

func getConfigNoHPA() v1alpha1.FlowCollectorSpec {
	cfg := getConfig()
	cfg.Processor.KafkaConsumerAutoscaler = nil
	return cfg
}

func getAutoScalerSpecs() (ascv2.HorizontalPodAutoscaler, v1alpha1.FlowCollectorHPA) {
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

	return autoScaler, *getConfig().Processor.KafkaConsumerAutoscaler
}

func TestDaemonSetNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, &cfg, true)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.daemonSet(digest)

	// Check no change
	cfg = getConfig()
	b = newMonolithBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)

	assert.False(daemonSetNeedsUpdate(first, &cfg.Processor, digest))
}

func TestDaemonSetChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, &cfg, true)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.daemonSet(digest)

	// Check probes enabled change
	cfg.Processor.EnableKubeProbes = true
	b = newMonolithBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(first, &cfg.Processor, digest))

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = newMonolithBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	third := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(second, &cfg.Processor, digest))

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newMonolithBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fourth := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(third, &cfg.Processor, digest))

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newMonolithBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)

	assert.True(daemonSetNeedsUpdate(fourth, &cfg.Processor, digest))
	assert.False(daemonSetNeedsUpdate(third, &cfg.Processor, digest))
}

func TestDeploymentNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newTransfoBuilder(ns, &cfg, true)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check no change
	cfg = getConfig()
	b = newTransfoBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)

	assert.False(deploymentNeedsUpdate(first, &cfg.Processor, digest))
}

func TestDeploymentChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newTransfoBuilder(ns, &cfg, true)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check probes enabled change
	cfg.Processor.EnableKubeProbes = true
	b = newTransfoBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	second := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(first, &cfg.Processor, digest))

	// Check log level change
	cfg.Processor.LogLevel = "info"
	b = newTransfoBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	third := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(second, &cfg.Processor, digest))

	// Check resource change
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newTransfoBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fourth := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(third, &cfg.Processor, digest))

	// Check reverting limits
	cfg.Processor.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newTransfoBuilder(ns, &cfg, true)
	_, digest, err = b.configMap()
	assert.NoError(err)
	fifth := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(fourth, &cfg.Processor, digest))
	assert.False(deploymentNeedsUpdate(third, &cfg.Processor, digest))

	// Check replicas didn't change because HPA is used
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = 5
	b = newTransfoBuilder(ns, &cfg2, true)
	_, digest, err = b.configMap()
	assert.NoError(err)

	assert.False(deploymentNeedsUpdate(fifth, &cfg2.Processor, digest))
}

func TestDeploymentChangedReplicasNoHPA(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfigNoHPA()
	b := newTransfoBuilder(ns, &cfg, true)
	_, digest, err := b.configMap()
	assert.NoError(err)
	first := b.deployment(digest)

	// Check replicas changed (need to copy flp, as Spec.Replicas stores a pointer)
	cfg2 := cfg
	cfg2.Processor.KafkaConsumerReplicas = 5
	b = newTransfoBuilder(ns, &cfg2, true)
	_, digest, err = b.configMap()
	assert.NoError(err)

	assert.True(deploymentNeedsUpdate(first, &cfg2.Processor, digest))
}

func TestServiceNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, &cfg, true)
	first := b.newPromService()

	// Check no change
	newService := first.DeepCopy()

	assert.False(serviceNeedsUpdate(first, newService))
}

func TestServiceChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	cfg := getConfig()
	b := newMonolithBuilder(ns, &cfg, true)
	first := b.newPromService()

	// Check port changed
	cfg.Processor.MetricsServer.Port = 9999
	b = newMonolithBuilder(ns, &cfg, true)
	second := b.fromPromService(first)

	assert.True(serviceNeedsUpdate(first, second))

	// Make sure non-service settings doesn't trigger service update
	cfg.Processor.LogLevel = "error"
	b = newMonolithBuilder(ns, &cfg, true)
	third := b.fromPromService(first)

	assert.False(serviceNeedsUpdate(second, third))
}

func TestConfigMapShouldDeserializeAsJSON(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	cfg := getConfig()
	loki := cfg.Loki
	b := newMonolithBuilder(ns, &cfg, true)
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
	assert.Equal(cfg.Processor.Port, int32(params[0].Ingest.Collector.Port))

	lokiCfg := params[2].Write.Loki
	assert.Equal(loki.URL, lokiCfg.URL)
	assert.Equal(loki.BatchWait.Duration.String(), lokiCfg.BatchWait)
	assert.Equal(loki.MinBackoff.Duration.String(), lokiCfg.MinBackoff)
	assert.Equal(loki.MaxBackoff.Duration.String(), lokiCfg.MaxBackoff)
	assert.EqualValues(loki.MaxRetries, lokiCfg.MaxRetries)
	assert.EqualValues(loki.BatchSize, lokiCfg.BatchSize)
	assert.EqualValues([]string{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection"}, lokiCfg.Labels)
	assert.Equal(`{app="netobserv-flowcollector"}`, fmt.Sprintf("%v", lokiCfg.StaticLabels))

	assert.Equal(cfg.Processor.MetricsServer.Port, int32(params[4].Encode.Prom.Port))
}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, hpa := getAutoScalerSpecs()
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &hpa, testNamespace), false)

	//wrong max replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &hpa, testNamespace), true)

	//missing min replicas
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &hpa, testNamespace), true)

	//missing metrics
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &hpa, testNamespace), true)

	//wrong namespace
	autoScalerSpec, hpa = getAutoScalerSpecs()
	autoScalerSpec.Namespace = "NewNamespace"
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &hpa, testNamespace), true)
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	builder := newMonolithBuilder("ns", &cfg, true)
	tBuilder := newTransfoBuilder("ns", &cfg, true)
	iBuilder := newIngestBuilder("ns", &cfg, true)

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
	b := newMonolithBuilder(ns, &cfg, true)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"enrich","follows":"ipfix"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))

	// Kafka Ingester
	cfg.DeploymentModel = v1alpha1.DeploymentModelKafka
	bi := newIngestBuilder(ns, &cfg, true)
	stages, parameters, err = bi.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ = json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"kafka-write","follows":"ipfix"}]`, string(jsonStages))

	// Kafka Transformer
	bt := newTransfoBuilder(ns, &cfg, true)
	stages, parameters, err = bt.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ = json.Marshal(stages)
	assert.Equal(`[{"name":"kafka-read"},{"name":"enrich","follows":"kafka-read"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
}

func TestPipelineConfigDropUnused(t *testing.T) {
	assert := assert.New(t)

	// Single config
	ns := "namespace"
	cfg := getConfig()
	cfg.Processor.LogLevel = "info"
	cfg.Processor.DropUnusedFields = true
	b := newMonolithBuilder(ns, &cfg, true)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"filter","follows":"ipfix"},{"name":"enrich","follows":"filter"},{"name":"loki","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))

	jsonParams, _ := json.Marshal(parameters[1].Transform.Filter)
	assert.Contains(string(jsonParams), `{"input":"CustomBytes1","type":"remove_field"}`)
	assert.Contains(string(jsonParams), `{"input":"CustomInteger5","type":"remove_field"}`)
	assert.Contains(string(jsonParams), `{"input":"MPLS1Label","type":"remove_field"}`)
}

func TestPipelineTraceStage(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := newMonolithBuilder("namespace", &cfg, true)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"enrich","follows":"ipfix"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
}

func TestMergeMetricsConfigurationNoIgnore(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()

	b := newMonolithBuilder("namespace", &cfg, true)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"enrich","follows":"ipfix"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
	assert.Len(parameters[4].Encode.Prom.Metrics, 6)
	assert.Equal("node_received_bytes_total", parameters[4].Encode.Prom.Metrics[0].Name)
	assert.Equal("node_sent_bytes_total", parameters[4].Encode.Prom.Metrics[1].Name)
	assert.Equal("received_bytes_total", parameters[4].Encode.Prom.Metrics[2].Name)
	assert.Equal("received_packets_total", parameters[4].Encode.Prom.Metrics[3].Name)
	assert.Equal("sent_bytes_total", parameters[4].Encode.Prom.Metrics[4].Name)
	assert.Equal("sent_packets_total", parameters[4].Encode.Prom.Metrics[5].Name)
	assert.Equal("netobserv_", parameters[4].Encode.Prom.Prefix)
}

func TestMergeMetricsConfigurationWithIgnore(t *testing.T) {
	assert := assert.New(t)

	cfg := getConfig()
	cfg.Processor.IgnoreMetrics = []string{"nodes"}

	b := newMonolithBuilder("namespace", &cfg, true)
	stages, parameters, err := b.buildPipelineConfig()
	assert.NoError(err)
	assert.True(validatePipelineConfig(stages, parameters))
	jsonStages, _ := json.Marshal(stages)
	assert.Equal(`[{"name":"ipfix"},{"name":"enrich","follows":"ipfix"},{"name":"loki","follows":"enrich"},{"name":"stdout","follows":"enrich"},{"name":"prometheus","follows":"enrich"}]`, string(jsonStages))
	assert.Len(parameters[4].Encode.Prom.Metrics, 4)
	assert.Equal("received_bytes_total", parameters[4].Encode.Prom.Metrics[0].Name)
	assert.Equal("netobserv_", parameters[4].Encode.Prom.Prefix)
}
