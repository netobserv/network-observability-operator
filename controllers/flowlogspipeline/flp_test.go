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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
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

func getFLPConfig() flowsv1alpha1.FlowCollectorFLP {
	return flowsv1alpha1.FlowCollectorFLP{
		Replicas:        1,
		Port:            2055,
		Image:           image,
		ImagePullPolicy: string(pullPolicy),
		LogLevel:        "trace",
		Resources:       resources,
		HPA: &flowsv1alpha1.FlowCollectorHPA{
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
		HealthPort:     8080,
		PrometheusPort: 9090,
	}
}

func getFLPConfigNoHPA() flowsv1alpha1.FlowCollectorFLP {
	return flowsv1alpha1.FlowCollectorFLP{
		Replicas:        1,
		Port:            2055,
		Image:           image,
		ImagePullPolicy: string(pullPolicy),
		LogLevel:        "trace",
		Resources:       resources,
		HealthPort:      8080,
	}
}

func getLokiConfig() flowsv1alpha1.FlowCollectorLoki {
	return flowsv1alpha1.FlowCollectorLoki{
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
	}
}

func getAutoScalerSpecs() (ascv2.HorizontalPodAutoscaler, flowsv1alpha1.FlowCollectorFLP) {
	var autoScaler = ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind: constants.DeploymentKind,
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

	return autoScaler, getFLPConfig()
}

func TestDaemonSetNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest := b.configMap()
	first := b.daemonSet(digest)

	// Check no change
	flp = getFLPConfig()
	loki = getLokiConfig()
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()

	assert.False(daemonSetNeedsUpdate(first, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
}

func TestDaemonSetChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest := b.configMap()
	first := b.daemonSet(digest)

	// Check probes enabled change
	flp.EnableKubeProbes = true
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	second := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(first, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check log level change
	flp.LogLevel = "info"
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	third := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(second, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check resource change
	flp.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	fourth := b.daemonSet(digest)

	assert.True(daemonSetNeedsUpdate(third, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check reverting limits
	flp.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()

	assert.True(daemonSetNeedsUpdate(fourth, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
	assert.False(daemonSetNeedsUpdate(third, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
}

func TestDeploymentNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest := b.configMap()
	first := b.deployment(digest)

	// Check no change
	flp = getFLPConfig()
	loki = getLokiConfig()
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()

	assert.False(deploymentNeedsUpdate(first, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
}

func TestDeploymentChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest := b.configMap()
	first := b.deployment(digest)

	// Check probes enabled change
	flp.EnableKubeProbes = true
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	second := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(first, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check log level change
	flp.LogLevel = "info"
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	third := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(second, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check resource change
	flp.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	fourth := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(third, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check reverting limits
	flp.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	}
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest = b.configMap()
	fifth := b.deployment(digest)

	assert.True(deploymentNeedsUpdate(fourth, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
	assert.False(deploymentNeedsUpdate(third, &flp, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))

	// Check replicas didn't change because HPA is used
	flp2 := flp
	flp2.Replicas = 5
	b = newBuilder(ns, corev1.ProtocolUDP, &flp2, &loki, ConfSingle, true)
	_, digest = b.configMap()

	assert.False(deploymentNeedsUpdate(fifth, &flp2, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
}

func TestDeploymentChangedReplicasNoHPA(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfigNoHPA()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	_, digest := b.configMap()
	first := b.deployment(digest)

	// Check replicas changed (need to copy flp, as Spec.Replicas stores a pointer)
	flp2 := flp
	flp2.Replicas = 5
	b = newBuilder(ns, corev1.ProtocolUDP, &flp2, &loki, ConfSingle, true)
	_, digest = b.configMap()

	assert.True(deploymentNeedsUpdate(first, &flp2, digest, constants.FLPName+FlpConfSuffix[ConfSingle]))
}

func TestServiceNoChange(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	first := b.service(nil)

	// Check no change
	newService := first.DeepCopy()

	assert.False(serviceNeedsUpdate(first, newService))
}

func TestServiceChanged(t *testing.T) {
	assert := assert.New(t)

	// Get first
	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	first := b.service(nil)

	// Check port changed
	flp.Port = 9999
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	second := b.service(first)

	assert.True(serviceNeedsUpdate(first, second))

	// Make sure non-service settings doesn't trigger service update
	flp.LogLevel = "error"
	b = newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	third := b.service(first)

	assert.False(serviceNeedsUpdate(second, third))
}

func TestConfigMapShouldDeserializeAsYAML(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	flp := getFLPConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, corev1.ProtocolUDP, &flp, &loki, ConfSingle, true)
	cm, digest := b.configMap()
	assert.NotEmpty(t, digest)

	assert.Equal("dev", cm.Labels["version"])

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal("trace", decoded["log-level"])

	parameters := decoded["parameters"].([]interface{})
	ingest := parameters[0].(map[interface{}]interface{})["ingest"].(map[interface{}]interface{})
	collector := ingest["collector"].(map[interface{}]interface{})
	assert.Equal(flp.Port, int32(collector["port"].(int)))

	lokiCfg := parameters[3].(map[interface{}]interface{})["write"].(map[interface{}]interface{})["loki"].(map[interface{}]interface{})
	assert.Equal(loki.URL, lokiCfg["url"])
	assert.Equal(loki.BatchWait.Duration.String(), lokiCfg["batchWait"])
	assert.Equal(loki.MinBackoff.Duration.String(), lokiCfg["minBackoff"])
	assert.Equal(loki.MaxBackoff.Duration.String(), lokiCfg["maxBackoff"])
	assert.EqualValues(loki.MaxRetries, lokiCfg["maxRetries"])
	assert.EqualValues(loki.BatchSize, lokiCfg["batchSize"])
	assert.EqualValues([]interface{}{"SrcK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_Namespace", "DstK8S_OwnerName", "FlowDirection"}, lokiCfg["labels"])
	assert.Equal(fmt.Sprintf("%v", loki.StaticLabels), fmt.Sprintf("%v", lokiCfg["staticLabels"]))

	encode := parameters[5].(map[interface{}]interface{})["encode"].(map[interface{}]interface{})
	prom := encode["prom"].(map[interface{}]interface{})
	assert.Equal(flp.PrometheusPort, int32(prom["port"].(int)))

}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, flp := getAutoScalerSpecs()
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &flp, testNamespace), false)

	//wrong max replicas
	autoScalerSpec, flp = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &flp, testNamespace), true)

	//missing min replicas
	autoScalerSpec, flp = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &flp, testNamespace), true)

	//missing metrics
	autoScalerSpec, flp = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &flp, testNamespace), true)

	//wrong namespace
	autoScalerSpec, flp = getAutoScalerSpecs()
	autoScalerSpec.Namespace = "NewNamespace"
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &flp, testNamespace), true)
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	flpk := getFLPConfig()
	builder := newBuilder("ns", corev1.ProtocolUDP, &flpk, nil, ConfSingle, true)

	// Deployment
	depl := builder.deployment("digest")
	assert.Equal("flowlogs-pipeline", depl.Labels["app"])
	assert.Equal("flowlogs-pipeline", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// DaemonSet
	ds := builder.daemonSet("digest")
	assert.Equal("flowlogs-pipeline", ds.Labels["app"])
	assert.Equal("flowlogs-pipeline", ds.Spec.Template.Labels["app"])
	assert.Equal("dev", ds.Labels["version"])
	assert.Equal("dev", ds.Spec.Template.Labels["version"])

	// Service
	svc := builder.service(nil)
	assert.Equal("flowlogs-pipeline", svc.Labels["app"])
	assert.Equal("flowlogs-pipeline", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])
}

func TestDeployNeeded(t *testing.T) {
	assert := assert.New(t)

	kafka := flowsv1alpha1.FlowCollectorKafka{Enable: false, Address: "loaclhost:9092", Topic: "FLP"}
	// Kafka not configured
	res, err := checkDeployNeeded(kafka, ConfSingle)
	assert.True(res)
	assert.NoError(err)
	res, err = checkDeployNeeded(kafka, ConfKafkaIngester)
	assert.False(res)
	assert.NoError(err)
	res, err = checkDeployNeeded(kafka, ConfKafkaTransformer)
	assert.False(res)
	assert.NoError(err)

	// Kafka configured
	kafka.Enable = true
	res, err = checkDeployNeeded(kafka, ConfSingle)
	assert.False(res)
	assert.NoError(err)
	res, err = checkDeployNeeded(kafka, ConfKafkaIngester)
	assert.True(res)
	assert.NoError(err)
	res, err = checkDeployNeeded(kafka, ConfKafkaTransformer)
	assert.True(res)
	assert.NoError(err)

}
