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

package goflowkube

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	ascv1 "k8s.io/api/autoscaling/v1"
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
var commands = []string{
	"/bin/sh",
	"-c",
	`/goflow-kube -loglevel "trace" -config /etc/goflow-kube/config.yaml -healthport 8080`,
}
var image = "quay.io/netobserv/goflow2-kube:dev"
var pullPolicy = corev1.PullIfNotPresent
var minReplicas = int32(1)
var maxReplicas = int32(5)
var targetCPU = int32(75)

const testNamespace = "goflowkube"

func getGoflowKubeConfig() flowsv1alpha1.FlowCollectorGoflowKube {
	return flowsv1alpha1.FlowCollectorGoflowKube{
		Port:            2055,
		Image:           image,
		ImagePullPolicy: string(pullPolicy),
		LogLevel:        "trace",
		Resources:       resources,
		HPA: &flowsv1alpha1.FlowCollectorHPA{
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &targetCPU,
		},
		PrintOutput: false,
		HealthPort:  8080,
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

func getContainerSpecs() (corev1.PodSpec, flowsv1alpha1.FlowCollectorGoflowKube) {
	var podSpec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            constants.GoflowKubeName,
				Image:           image,
				Command:         commands,
				Resources:       resources,
				ImagePullPolicy: pullPolicy,
			},
		},
	}

	return podSpec, getGoflowKubeConfig()
}

func getServiceSpecs() (corev1.Service, flowsv1alpha1.FlowCollectorGoflowKube) {
	var service = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     2055,
					Protocol: "UDP",
				},
			},
		},
	}

	return service, getGoflowKubeConfig()
}

func TestBuildMainCommand(t *testing.T) {
	assert := assert.New(t)

	_, goflowKube := getContainerSpecs()
	cmd := buildMainCommand(&goflowKube)
	assert.Equal(commands[2], cmd)
}

func getAutoScalerSpecs() (ascv1.HorizontalPodAutoscaler, flowsv1alpha1.FlowCollectorGoflowKube) {
	var autoScaler = ascv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: ascv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv1.CrossVersionObjectReference{
				Kind: constants.DeploymentKind,
				Name: constants.GoflowKubeName,
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &targetCPU,
		},
	}

	return autoScaler, getGoflowKubeConfig()
}

func TestContainerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	podSpec, goflowKube := getContainerSpecs()
	assert.Equal(containerNeedsUpdate(&podSpec, &goflowKube), false)

	//wrong command
	podSpec, goflowKube = getContainerSpecs()
	podSpec.Containers[0].Command = []string{"/bin/sh"}
	assert.Equal(containerNeedsUpdate(&podSpec, &goflowKube), true)

	//wrong log level
	podSpec, goflowKube = getContainerSpecs()
	goflowKube.LogLevel = "info"
	assert.Equal(containerNeedsUpdate(&podSpec, &goflowKube), true)

	//wrong resources
	podSpec, goflowKube = getContainerSpecs()
	goflowKube.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	assert.Equal(containerNeedsUpdate(&podSpec, &goflowKube), true)
}

func TestServiceUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	serviceSpec, goflowKube := getServiceSpecs()
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &goflowKube, testNamespace), false)

	//wrong port protocol
	serviceSpec, goflowKube = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "TCP"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &goflowKube, testNamespace), true)

	//wrong namespace
	serviceSpec, goflowKube = getServiceSpecs()
	serviceSpec.Namespace = "NewNamespace"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &goflowKube, testNamespace), true)
}

func TestConfigMapShouldDeserializeAsYAML(t *testing.T) {
	assert := assert.New(t)

	ns := "namespace"
	goflowKube := getGoflowKubeConfig()
	loki := getLokiConfig()
	b := newBuilder(ns, &goflowKube, &loki)
	cm, digest := b.configMap()
	assert.NotEmpty(t, digest)

	assert.Equal("dev", cm.Labels["version"])

	data, ok := cm.Data[configFile]
	assert.True(ok)

	var decoded map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &decoded)

	assert.Nil(err)
	assert.Equal(fmt.Sprintf("netflow://:%d", goflowKube.Port), decoded["listen"])
	assert.Equal(goflowKube.PrintOutput, decoded["printOutput"])

	lokiCfg := decoded["loki"].(map[interface{}]interface{})
	assert.Equal(loki.URL, lokiCfg["url"])
	assert.Equal(loki.BatchWait.Duration.String(), lokiCfg["batchWait"])
	assert.Equal(loki.MinBackoff.Duration.String(), lokiCfg["minBackoff"])
	assert.Equal(loki.MaxBackoff.Duration.String(), lokiCfg["maxBackoff"])
	assert.EqualValues(loki.MaxRetries, lokiCfg["maxRetries"])
	assert.EqualValues(loki.BatchSize, lokiCfg["batchSize"])
	assert.EqualValues([]interface{}{"SrcNamespace", "SrcWorkload", "DstNamespace", "DstWorkload", "FlowDirection"}, lokiCfg["labels"])
	assert.Equal(fmt.Sprintf("%v", loki.StaticLabels), fmt.Sprintf("%v", lokiCfg["staticLabels"]))
}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, goflowKube := getAutoScalerSpecs()
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &goflowKube, testNamespace), false)

	//wrong max replicas
	autoScalerSpec, goflowKube = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &goflowKube, testNamespace), true)

	//missing min replicas
	autoScalerSpec, goflowKube = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &goflowKube, testNamespace), true)

	//missing min target CPU
	autoScalerSpec, goflowKube = getAutoScalerSpecs()
	autoScalerSpec.Spec.TargetCPUUtilizationPercentage = nil
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &goflowKube, testNamespace), true)

	//wrong namespace
	autoScalerSpec, goflowKube = getAutoScalerSpecs()
	autoScalerSpec.Namespace = "NewNamespace"
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &goflowKube, testNamespace), true)
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	gfk := getGoflowKubeConfig()
	builder := newBuilder("ns", &gfk, nil)

	// Deployment
	depl := builder.deployment("digest")
	assert.Equal("goflow-kube", depl.Labels["app"])
	assert.Equal("goflow-kube", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// DaemonSet
	ds := builder.daemonSet("digest")
	assert.Equal("goflow-kube", ds.Labels["app"])
	assert.Equal("goflow-kube", ds.Spec.Template.Labels["app"])
	assert.Equal("dev", ds.Labels["version"])
	assert.Equal("dev", ds.Spec.Template.Labels["version"])

	// Service
	svc := builder.service(nil)
	assert.Equal("goflow-kube", svc.Labels["app"])
	assert.Equal("goflow-kube", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])
}
