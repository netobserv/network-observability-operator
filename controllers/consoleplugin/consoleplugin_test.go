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
	"k8s.io/utils/ptr"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
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
		Enable:          ptr.To(true),
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

func getServiceSpecs() corev1.Service {
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

	return service
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
	loki := helper.LokiConfig{
		LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://loki:3100/", TenantID: "netobserv"},
	}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin}
	builder := newBuilder(testNamespace, testImage, &spec, &loki)
	old := builder.deployment("digest")
	nEw := builder.deployment("digest")
	report := helper.NewChangeReport("")
	assert.False(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "no change")

	//wrong resources
	spec.ConsolePlugin.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("500Gi"),
	}
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "req/limit changed")
	old = nEw

	//new image
	builder.imageName = "quay.io/netobserv/network-observability-console-plugin:latest"
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Image changed")
	old = nEw

	//new pull policy
	spec.ConsolePlugin.ImagePullPolicy = string(corev1.PullAlways)
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Pull policy changed")
	old = nEw

	//new log level
	spec.ConsolePlugin.LogLevel = "debug"
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Args changed")
	old = nEw

	//new loki config
	loki = helper.LokiConfig{
		LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://loki:3100/", TenantID: "netobserv", TLS: flowslatest.ClientTLS{
			Enable: true,
			CACert: flowslatest.CertificateReference{
				Type:     "configmap",
				Name:     "cm-name",
				CertFile: "ca.crt",
			},
		}}}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
	old = nEw

	//new loki cert name
	loki.LokiManualParams.TLS.CACert.Name = "cm-name-2"
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.True(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "Volumes changed")
	old = nEw

	//test again no change
	loki.LokiManualParams.TLS.CACert.Name = "cm-name-2"
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw = builder.deployment("digest")
	report = helper.NewChangeReport("")
	assert.False(helper.PodChanged(&old.Spec.Template, &nEw.Spec.Template, constants.PluginName, &report))
	assert.Contains(report.String(), "no change")
}

func TestConfigMapUpdateCheck(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	plugin := getPluginConfig()
	loki := helper.LokiConfig{
		LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://loki:3100/", TenantID: "netobserv"},
	}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin}
	builder := newBuilder(testNamespace, testImage, &spec, &loki)
	old, _, _ := builder.configMap()
	nEw, _, _ := builder.configMap()
	assert.Equal(old.Data, nEw.Data)

	//update loki
	loki = helper.LokiConfig{
		LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://loki:3100/", TenantID: "netobserv", TLS: flowslatest.ClientTLS{
			Enable: true,
			CACert: flowslatest.CertificateReference{
				Type:     "configmap",
				Name:     "cm-name",
				CertFile: "ca.crt",
			},
		}},
	}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
	old = nEw

	//set status url and enable default tls
	loki.LokiManualParams.StatusURL = "http://loki.status:3100/"
	loki.LokiManualParams.StatusTLS.Enable = true
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
	old = nEw

	//update status ca cert
	loki.LokiManualParams.StatusTLS.CACert = flowslatest.CertificateReference{
		Type:     "configmap",
		Name:     "status-cm-name",
		CertFile: "status-ca.crt",
	}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
	old = nEw

	//update status user cert
	loki.LokiManualParams.StatusTLS.UserCert = flowslatest.CertificateReference{
		Type:     "secret",
		Name:     "sec-name",
		CertFile: "tls.crt",
		CertKey:  "tls.key",
	}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
}

func TestConfigMapUpdateWithLokistackMode(t *testing.T) {
	assert := assert.New(t)

	//equals specs
	plugin := getPluginConfig()
	lokiSpec := flowslatest.FlowCollectorLoki{
		Mode:      flowslatest.LokiModeLokiStack,
		LokiStack: flowslatest.LokiStackRef{Name: "lokistack", Namespace: "ls-namespace"},
	}
	loki := helper.NewLokiConfig(&lokiSpec, "any")
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: lokiSpec}
	builder := newBuilder(testNamespace, testImage, &spec, &loki)
	old, _, _ := builder.configMap()
	nEw, _, _ := builder.configMap()
	assert.Equal(old.Data, nEw.Data)

	//update lokistack name
	lokiSpec.LokiStack.Name = "lokistack-updated"
	loki = helper.NewLokiConfig(&lokiSpec, "any")

	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: lokiSpec}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
	old = nEw

	//update lokistack namespace
	lokiSpec.LokiStack.Namespace = "ls-namespace-updated"
	loki = helper.NewLokiConfig(&lokiSpec, "any")

	spec = flowslatest.FlowCollectorSpec{ConsolePlugin: plugin, Loki: lokiSpec}
	builder = newBuilder(testNamespace, testImage, &spec, &loki)
	nEw, _, _ = builder.configMap()
	assert.NotEqual(old.Data, nEw.Data)
}

func TestServiceUpdateCheck(t *testing.T) {
	assert := assert.New(t)
	old := getServiceSpecs()

	//equals specs
	serviceSpec := getServiceSpecs()
	report := helper.NewChangeReport("")
	assert.Equal(helper.ServiceChanged(&old, &serviceSpec, &report), false)
	assert.Contains(report.String(), "no change")

	//wrong port protocol
	serviceSpec = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Protocol = "UDP"
	report = helper.NewChangeReport("")
	assert.Equal(helper.ServiceChanged(&old, &serviceSpec, &report), true)
	assert.Contains(report.String(), "Service spec changed")

	//wrong port number
	serviceSpec = getServiceSpecs()
	serviceSpec.Spec.Ports[0].Port = 8080
	report = helper.NewChangeReport("")
	assert.Equal(helper.ServiceChanged(&old, &serviceSpec, &report), true)
	assert.Contains(report.String(), "Service spec changed")
}

func TestBuiltService(t *testing.T) {
	assert := assert.New(t)

	//newly created service should not need update
	plugin := getPluginConfig()
	loki := helper.LokiConfig{LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://foo:1234"}}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin}
	builder := newBuilder(testNamespace, testImage, &spec, &loki)
	old := builder.mainService()
	nEw := builder.mainService()
	report := helper.NewChangeReport("")
	assert.Equal(helper.ServiceChanged(old, nEw, &report), false)
	assert.Contains(report.String(), "no change")
}

func TestLabels(t *testing.T) {
	assert := assert.New(t)

	plugin := getPluginConfig()
	loki := helper.LokiConfig{LokiManualParams: flowslatest.LokiManualParams{IngesterURL: "http://foo:1234"}}
	spec := flowslatest.FlowCollectorSpec{ConsolePlugin: plugin}
	builder := newBuilder(testNamespace, testImage, &spec, &loki)

	// Deployment
	depl := builder.deployment("digest")
	assert.Equal("netobserv-plugin", depl.Labels["app"])
	assert.Equal("netobserv-plugin", depl.Spec.Template.Labels["app"])
	assert.Equal("dev", depl.Labels["version"])
	assert.Equal("dev", depl.Spec.Template.Labels["version"])

	// Service
	svc := builder.mainService()
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
	assert.Equal(t, `{"tls_config":{"insecure_skip_verify":true},"follow_redirects":false,"enable_http2":false,"proxy_url":null}`, string(bs))

	config2 := promConfig.HTTPClientConfig{}
	err = json.Unmarshal(bs, &config2)
	assert.Nil(t, err)
	assert.True(t, config2.TLSConfig.InsecureSkipVerify)
	assert.Equal(t, promConfig.URL{}, config2.ProxyURL)

	err = config2.Validate()
	assert.Nil(t, err)
	assert.True(t, config2.TLSConfig.InsecureSkipVerify)
	assert.Nil(t, config2.ProxyURL.URL)
}
