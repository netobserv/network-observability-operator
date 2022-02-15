package consoleplugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
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
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
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
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, testNamespace), false)

	//wrong port protocol
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "UDP"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, testNamespace), true)

	//wrong port number
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Port = 8080
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, testNamespace), true)

	//wrong namespace
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Namespace = "OldNamespace"
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, testNamespace), true)

}

func TestBuiltContainer(t *testing.T) {
	assert := assert.New(t)

	//newly created containers should not need update
	plugin := getPluginConfig()
	loki := &flowsv1alpha1.FlowCollectorLoki{URL: "http://foo:1234"}
	builder := newBuilder(testNamespace, &plugin, loki)
	newContainer := builder.podTemplate()
	assert.Equal(containerNeedsUpdate(&newContainer.Spec, &plugin), false)
}

func TestBuiltService(t *testing.T) {
	assert := assert.New(t)

	//newly created service should not need update
	plugin := getPluginConfig()
	builder := newBuilder(testNamespace, &plugin, nil)
	newService := builder.service(nil)
	assert.Equal(serviceNeedsUpdate(newService, &plugin, testNamespace), false)
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	plugin := getPluginConfig()
	loki := &flowsv1alpha1.FlowCollectorLoki{URL: "http://foo:1234"}
	builder := newBuilder(testNamespace, &plugin, loki)

	// Deployment
	depl := builder.deployment()
	assert.Equal("network-observability-plugin", depl.Labels["app"])
	assert.Equal("network-observability-plugin", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// Service
	svc := builder.service(nil)
	assert.Equal("network-observability-plugin", svc.Labels["app"])
	assert.Equal("network-observability-plugin", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])
}
