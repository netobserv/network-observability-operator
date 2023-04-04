package consoleplugin

import (
	"encoding/json"
	"testing"

	promConfig "github.com/prometheus/common/config"
	"github.com/stretchr/testify/assert"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const testImage = "quay.io/netobserv/network-observability-console-plugin:dev"
const testNamespace = constants.PluginName

var testPullPolicy = corev1.PullIfNotPresent
var testResources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("512Mi"),
	},
}

func getPluginConfig() flowslatest.FlowCollectorConsolePlugin {
	return flowslatest.FlowCollectorConsolePlugin{
		Port:            9001,
		ImagePullPolicy: string(testPullPolicy),
		Resources:       testResources,
		Autoscaler: flowslatest.FlowCollectorHPA{
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
		LogLevel: "info",
	}
}

func getServiceSpecs() (corev1.Service, flowslatest.FlowCollectorConsolePlugin) {
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

func getAutoScalerSpecs() (ascv2.HorizontalPodAutoscaler, flowslatest.FlowCollectorConsolePlugin) {
	var autoScaler = ascv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
		},
		Spec: ascv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: ascv2.CrossVersionObjectReference{
				Kind: "Deployment",
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
	plugin := getPluginConfig()
	loki := flowslatest.FlowCollectorLoki{URL: "http://loki:3100/", TenantID: "netobserv"}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder := newBuilder(testNamespace, testImage, &spec)
	old := builder.deployment("digest")
	new := builder.deployment("digest")
	report := helper.NewChangeReport("")
	assert.False(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "no change")

	//wrong resources
	spec.ConsolePlugin.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "req/limit changed")
	old = new

	//new image
	builder.imageName = "quay.io/netobserv/network-observability-console-plugin:latest"
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Image changed")
	old = new

	//new pull policy
	spec.ConsolePlugin.ImagePullPolicy = string(corev1.PullAlways)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Pull policy changed")
	old = new

	//new log level
	spec.ConsolePlugin.LogLevel = "debug"
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Args changed")
	old = new

	//new loki config
	loki = flowslatest.FlowCollectorLoki{URL: "http://loki:3100/", TenantID: "netobserv", TLS: flowslatest.ClientTLS{
		Enable: true,
		CACert: flowslatest.CertificateReference{
			Type:     "configmap",
			Name:     "cm-name",
			CertFile: "ca.crt",
		},
	}}
	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
	old = new

	//new loki cert name
	loki.TLS.CACert.Name = "cm-name-2"
	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
	old = new

	//test again no change
	loki.TLS.CACert.Name = "cm-name-2"
	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.False(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "no change")
	old = new

	//set status url and enable default tls
	loki.StatusURL = "http://loki.status:3100/"
	loki.StatusTLS.Enable = true

	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Container changed")
	old = new

	//update status ca cert
	loki.StatusTLS.CACert = flowslatest.CertificateReference{
		Type:     "configmap",
		Name:     "status-cm-name",
		CertFile: "status-ca.crt",
	}

	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
	old = new

	//update status user cert
	loki.StatusTLS.UserCert = flowslatest.CertificateReference{
		Type:     "secret",
		Name:     "sec-name",
		CertFile: "tls.crt",
		CertKey:  "tls.key",
	}

	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder = newBuilder(testNamespace, testImage, &spec)
	new = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &new.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
}

func TestServiceUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	serviceSpec, containerConfig := getServiceSpecs()
	report := helper.NewChangeReport("")
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, &report), false)
	assert.Contains(report.String(), "no change")

	//wrong port protocol
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "UDP"
	report = helper.NewChangeReport("")
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, &report), true)
	assert.Contains(report.String(), "Port change")

	//wrong port number
	serviceSpec, containerConfig = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Port = 8080
	report = helper.NewChangeReport("")
	assert.Equal(serviceNeedsUpdate(&serviceSpec, &containerConfig, &report), true)
	assert.Contains(report.String(), "Port change")
}

func TestBuiltService(t *testing.T) {
	assert := assert.New(t)

	//newly created service should not need update
	plugin := getPluginConfig()
	loki := flowslatest.FlowCollectorLoki{URL: "http://foo:1234"}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder := newBuilder(testNamespace, testImage, &spec)
	newService := builder.service(nil)
	report := helper.NewChangeReport("")
	assert.Equal(serviceNeedsUpdate(newService, &plugin, &report), false)
	assert.Contains(report.String(), "no change")
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	plugin := getPluginConfig()
	loki := flowslatest.FlowCollectorLoki{URL: "http://foo:1234"}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: loki}
	builder := newBuilder(testNamespace, testImage, &spec)

	// Deployment
	depl := builder.deployment("digest")
	assert.Equal("netobserv-plugin", depl.Labels["app"])
	assert.Equal("netobserv-plugin", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// Service
	svc := builder.service(nil)
	assert.Equal("netobserv-plugin", svc.Labels["app"])
	assert.Equal("netobserv-plugin", svc.Spec.Selector["app"])
	assert.Equal("dev", svc.Labels["version"])
	assert.Empty(svc.Spec.Selector["version"])
}

func TestAutoScalerUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	autoScalerSpec, plugin := getAutoScalerSpecs()
	report := helper.NewChangeReport("")
	assert.Equal(helper.AutoScalerChanged(&autoScalerSpec, plugin.Autoscaler, &report), false)
	assert.Contains(report.String(), "no change")

	//wrong max replicas
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.MaxReplicas = 10
	report = helper.NewChangeReport("")
	assert.Equal(helper.AutoScalerChanged(&autoScalerSpec, plugin.Autoscaler, &report), true)
	assert.Contains(report.String(), "Max replicas changed")

	//missing min replicas
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.MinReplicas = nil
	report = helper.NewChangeReport("")
	assert.Equal(helper.AutoScalerChanged(&autoScalerSpec, plugin.Autoscaler, &report), true)
	assert.Contains(report.String(), "Min replicas changed")

	//missing metrics
	autoScalerSpec, plugin = getAutoScalerSpecs()
	autoScalerSpec.Spec.Metrics = []ascv2.MetricSpec{}
	report = helper.NewChangeReport("")
	assert.Equal(helper.AutoScalerChanged(&autoScalerSpec, plugin.Autoscaler, &report), true)
	assert.Contains(report.String(), "Metrics changed")
}

// ensure HTTPClientConfig Marshal / Unmarshal works as expected for ProxyURL *URL
// ProxyURL should not be set when only TLSConfig.InsecureSkipVerify is specified
func TestHTTPClientConfig(t *testing.T) {
	config := promConfig.HTTPClientConfig{
		TLSConfig: promConfig.TLSConfig{
			InsecureSkipVerify: true,
		},
	}
	err := config.Validate()
	assert.Nil(t, err)

	bs, _ := json.Marshal(config)
	assert.Equal(t, string(bs), `{"proxy_url":null,"tls_config":{"insecure_skip_verify":true},"follow_redirects":false}`)

	config2 := promConfig.HTTPClientConfig{}
	err = json.Unmarshal(bs, &config2)
	assert.Nil(t, err)
	assert.Equal(t, config2.TLSConfig.InsecureSkipVerify, true)
	assert.Equal(t, config2.ProxyURL, promConfig.URL{})

	err = config2.Validate()
	assert.Nil(t, err)
	assert.Equal(t, config2.TLSConfig.InsecureSkipVerify, true)
	assert.Nil(t, config2.ProxyURL.URL, nil)
}
