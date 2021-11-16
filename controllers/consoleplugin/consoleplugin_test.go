package consoleplugin

import (
	"testing"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const testImage = "quay.io/netobserv/network-observability-console-plugin:dev"
const testNamespace = pluginName

var testPullPolicy = corev1.PullIfNotPresent
var testResources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	},
}

func getPluginConfig() flowsv1alpha1.FlowCollectorConsolePlugin {
	return flowsv1alpha1.FlowCollectorConsolePlugin{
		Port:            9001,
		Image:           testImage,
		ImagePullPolicy: string(testPullPolicy),
		Resources:       testResources,
	}
}

func getContainerSpecs() (corev1.PodSpec, flowsv1alpha1.FlowCollectorConsolePlugin) {
	var podSpec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            pluginName,
				Image:           testImage,
				Resources:       testResources,
				ImagePullPolicy: testPullPolicy,
			},
		},
	}

	return podSpec, getPluginConfig()
}

func getServiceSpecs() (corev1.Service, flowsv1alpha1.FlowCollectorConsolePlugin) {
	var service = corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     9001,
					Protocol: "TCP",
				},
			},
		},
	}

	return service, getPluginConfig()
}

func TestContainerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	podSpec, containerConfig := getContainerSpecs()
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig), false)

	//wrong resources
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig), true)

	//new image
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.Image = "quay.io/netobserv/network-observability-console-plugin:latest"
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig), true)

	//new pull policy
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.ImagePullPolicy = string(corev1.PullAlways)
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig), true)

}

func TestServiceUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	serviceSpec, containerConfig := getServiceSpecs()
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig), false)

	//wrong port protocol
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "UDP"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig), true)

	//wrong port number
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Port = 8080
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig), true)

}

func TestBuiltContainer(t *testing.T) {
	assert := assert.New(t)

	//newly created containers should not need update
	containerConfig := getPluginConfig()
	newContainer := buildPodTemplate(&containerConfig)
	assert.Equal(containerNeedsUpdate(&newContainer.Spec, &containerConfig), false)
}

func TestBuiltService(t *testing.T) {
	assert := assert.New(t)

	//newly created service should not need update
	containerConfig := getPluginConfig()
	newService := buildService(&containerConfig, testNamespace)
	assert.Equal(serviceNeedsUpdate(newService, &containerConfig), false)

}
