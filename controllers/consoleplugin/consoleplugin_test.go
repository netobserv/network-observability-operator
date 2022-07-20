package consoleplugin

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	ascv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"

	promConfig "github.com/prometheus/common/config"
)

const testImage = "quay.io/netobserv/network-observability-console-plugin:dev"
const testNamespace = constants.PluginName

var testArgs = []string{
	"-cert", "/var/serving-cert/tls.crt",
	"-key", "/var/serving-cert/tls.key",
	"-loki", "http://loki:3100/",
	"-loki-labels", "SrcK8S_Namespace,SrcK8S_OwnerName,DstK8S_Namespace,DstK8S_OwnerName,FlowDirection",
	"-loki-tenant-id", "netobserv",
	"-loki-skip-tls", "true",
	"-loglevel", "info",
	"-frontend-config", "/opt/app-root/config.yaml",
}
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
		LogLevel: "info",
	}
}

func getContainerSpecs() (corev1.PodSpec, flowsv1alpha1.FlowCollectorConsolePlugin) {
	var podSpec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            constants.PluginName,
				Image:           testImage,
				Resources:       testResources,
				ImagePullPolicy: testPullPolicy,
				Args:            testArgs,
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

var minReplicas = int32(1)
var maxReplicas = int32(5)
var targetCPU = int32(75)

func getAutoScalerSpecs() (ascv2.HorizontalPodAutoscaler, flowsv1alpha1.FlowCollectorConsolePlugin) {
	var autoScaler = ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind: constants.DeploymentKind,
				Name: constants.PluginName,
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

	return autoScaler, getPluginConfig()
}

func TestContainerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	podSpec, containerConfig := getContainerSpecs()
	loki := &flowsv1alpha1.FlowCollectorLoki{URL: "http://loki:3100/", TenantID: "netobserv"}
	fmt.Printf("%v\n", buildArgs(&containerConfig, loki))
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig, loki), false)

	//wrong resources
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig, loki), true)

	//new image
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.Image = "quay.io/netobserv/network-observability-console-plugin:latest"
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig, loki), true)

	//new pull policy
	podSpec, containerConfig = getContainerSpecs()
	containerConfig.ImagePullPolicy = string(corev1.PullAlways)
	assert.Equal(containerNeedsUpdate(&podSpec, &containerConfig, loki), true)

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
	loki := &flowsv1alpha1.FlowCollectorLoki{URL: "http://foo:1234", TenantID: "netobserv"}
	builder := newBuilder(testNamespace, &plugin, loki)
	newContainer := builder.podTemplate("digest")
	assert.Equal(containerNeedsUpdate(&newContainer.Spec, &plugin, loki), false)
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
	depl := builder.deployment("digest")
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

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, plugin := getAutoScalerSpecs()
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &plugin, testNamespace), false)

	//wrong max replicas
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &plugin, testNamespace), true)

	//missing min replicas
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &plugin, testNamespace), true)

	//missing metrics
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &plugin, testNamespace), true)

	//wrong namespace
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Namespace = "NewNamespace"
	assert.Equal(autoScalerNeedsUpdate(&autoScalerSpec, &plugin, testNamespace), true)
}

//ensure HTTPClientConfig Marshal / Unmarshal works as expected for ProxyURL *URL
//ProxyURL should not be set when only TLSConfig.InsecureSkipVerify is specified
func TestHTTPClientConfig(t *testing.T) {
	config := promConfig.HTTPClientConfig{
		TLSConfig: promConfig.TLSConfig{
			InsecureSkipVerify: true,
		},
	}
	bs, _ := json.Marshal(config)
	assert.Equal(t, string(bs), `{"tls_config":{"insecure_skip_verify":true}}`)

	config2 := promConfig.HTTPClientConfig{}
	json.Unmarshal(bs, &config2)
	assert.Equal(t, config2.TLSConfig.InsecureSkipVerify, true)
	assert.Nil(t, config2.ProxyURL)
}
