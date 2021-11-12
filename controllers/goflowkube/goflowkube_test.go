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

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	`/goflow-kube -loglevel "trace" -config /etc/goflow-kube/config.yaml`,
}
var image = "quay.io/netobserv/goflow2-kube:dev"
var pullPolicy = corev1.PullIfNotPresent

func getGoflowKubeConfig() flowsv1alpha1.FlowCollectorGoflowKube {
	return flowsv1alpha1.FlowCollectorGoflowKube{
		Port:            2055,
		Image:           image,
		ImagePullPolicy: string(pullPolicy),
		LogLevel:        "trace",
		Resources:       resources,
		PrintOutput:     false,
	}
}

func getLokiConfig() flowsv1alpha1.FlowCollectorLoki {
	return flowsv1alpha1.FlowCollectorLoki{
		URL: "http://loki:3100/",
		BatchWait: v1.Duration{
			Duration: 1,
		},
		BatchSize: 102400,
		MinBackoff: v1.Duration{
			Duration: 1,
		},
		MaxBackoff: v1.Duration{
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
				Name:            "goflow-kube",
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
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &goflowKube), false)

	//wrong port protocol
	serviceSpec, goflowKube = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "TCP"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &goflowKube), true)
}

func TestConfigMapShouldDeserializeAsYAML(t *testing.T) {
	assert := assert.New(t)

	goflowKube := getGoflowKubeConfig()
	loki := getLokiConfig()
	cm := buildConfigMap(&goflowKube, &loki, "namespace")
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
	assert.EqualValues([]interface{}{"SrcNamespace", "SrcWorkload", "DstNamespace", "DstWorkload"}, lokiCfg["labels"])
	assert.Equal(fmt.Sprintf("%v", loki.StaticLabels), fmt.Sprintf("%v", lokiCfg["staticLabels"]))
}
