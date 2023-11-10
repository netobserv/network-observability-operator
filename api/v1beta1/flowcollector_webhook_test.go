package v1beta1

import (
	"testing"

	"github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestBeta1ConversionRoundtrip_Loki(t *testing.T) {
	// Testing beta1 -> beta2 -> beta1
	assert := assert.New(t)

	initial := FlowCollector{
		Spec: FlowCollectorSpec{
			Loki: FlowCollectorLoki{
				Enable:     ptr.To(true),
				URL:        "http://loki",
				StatusURL:  "http://loki/status",
				QuerierURL: "http://loki/querier",
				TenantID:   "tenant",
				AuthToken:  LokiAuthForwardUserToken,
				TLS: ClientTLS{
					Enable:             true,
					InsecureSkipVerify: true,
				},
				StatusTLS: ClientTLS{
					Enable:             true,
					InsecureSkipVerify: true,
				},
				BatchSize: 1000,
			},
		},
	}

	var converted v1beta2.FlowCollector
	err := initial.ConvertTo(&converted)
	assert.NoError(err)

	assert.Equal(v1beta2.LokiModeManual, converted.Spec.Loki.Mode)
	assert.True(*converted.Spec.Loki.Enable)
	assert.Equal("http://loki", converted.Spec.Loki.Manual.IngesterURL)
	assert.Equal("http://loki/status", converted.Spec.Loki.Manual.StatusURL)
	assert.Equal("http://loki/querier", converted.Spec.Loki.Manual.QuerierURL)
	assert.Equal("tenant", converted.Spec.Loki.Manual.TenantID)
	assert.Equal(LokiAuthForwardUserToken, converted.Spec.Loki.Manual.AuthToken)
	assert.True(converted.Spec.Loki.Manual.TLS.Enable)
	assert.True(converted.Spec.Loki.Manual.TLS.InsecureSkipVerify)
	assert.True(converted.Spec.Loki.Manual.StatusTLS.Enable)
	assert.True(converted.Spec.Loki.Manual.StatusTLS.InsecureSkipVerify)

	// Other way
	var back FlowCollector
	err = back.ConvertFrom(&converted)
	assert.NoError(err)
	assert.Equal(initial.Spec.Loki, back.Spec.Loki)
}

func TestBeta2ConversionRoundtrip_Loki(t *testing.T) {
	// Testing beta2 -> beta1 -> beta2
	assert := assert.New(t)

	initial := v1beta2.FlowCollector{
		Spec: v1beta2.FlowCollectorSpec{
			Loki: v1beta2.FlowCollectorLoki{
				Enable: ptr.To(true),
				Mode:   v1beta2.LokiModeLokiStack,
				LokiStack: v1beta2.LokiStackRef{
					Name:      "lokiii",
					Namespace: "lokins",
				},
				BatchSize: 1000,
			},
		},
	}

	var converted FlowCollector
	err := converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.True(*converted.Spec.Loki.Enable)
	assert.Equal("https://lokiii-gateway-http.lokins.svc:8080/api/logs/v1/network/", converted.Spec.Loki.URL)
	assert.Equal("https://lokiii-query-frontend-http.lokins.svc:3100/", converted.Spec.Loki.StatusURL)
	assert.Equal("https://lokiii-gateway-http.lokins.svc:8080/api/logs/v1/network/", converted.Spec.Loki.QuerierURL)
	assert.Equal("network", converted.Spec.Loki.TenantID)
	assert.Equal(LokiAuthForwardUserToken, converted.Spec.Loki.AuthToken)
	assert.True(converted.Spec.Loki.TLS.Enable)
	assert.False(converted.Spec.Loki.TLS.InsecureSkipVerify)
	assert.True(converted.Spec.Loki.StatusTLS.Enable)
	assert.False(converted.Spec.Loki.StatusTLS.InsecureSkipVerify)

	// Other way
	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)
	assert.Equal(initial.Spec.Loki, back.Spec.Loki)
}

func TestBeta1ConversionRoundtrip_Metrics(t *testing.T) {
	// Testing beta1 -> beta2 -> beta1
	assert := assert.New(t)

	initial := FlowCollector{
		Spec: FlowCollectorSpec{
			Processor: FlowCollectorFLP{
				Metrics: FLPMetrics{
					DisableAlerts: []FLPAlert{AlertLokiError},
					IgnoreTags:    []string{"nodes", "workloads", "bytes", "ingress"},
				},
			},
		},
	}

	var converted v1beta2.FlowCollector
	err := initial.ConvertTo(&converted)
	assert.NoError(err)

	expectedDefaultMetrics := []string{"namespace_egress_packets_total", "namespace_flows_total", "namespace_rtt_seconds", "namespace_drop_packets_total", "namespace_dns_latency_seconds"}
	assert.Equal([]v1beta2.FLPAlert{v1beta2.AlertLokiError}, converted.Spec.Processor.Metrics.DisableAlerts)
	assert.NotNil(converted.Spec.Processor.Metrics.IncludeList)
	assert.Equal(expectedDefaultMetrics, *converted.Spec.Processor.Metrics.IncludeList)

	// Other way
	var back FlowCollector
	err = back.ConvertFrom(&converted)
	assert.NoError(err)
	// Here, includeList is preserved; it takes precedence over ignoreTags
	assert.Equal(expectedDefaultMetrics, *back.Spec.Processor.Metrics.IncludeList)
	assert.Equal(initial.Spec.Processor.Metrics.DisableAlerts, back.Spec.Processor.Metrics.DisableAlerts)
	assert.Equal(initial.Spec.Processor.Metrics.Server, back.Spec.Processor.Metrics.Server)
}

func TestBeta2ConversionRoundtrip_Metrics(t *testing.T) {
	// Testing beta2 -> beta1 -> beta2
	assert := assert.New(t)

	initial := v1beta2.FlowCollector{
		Spec: v1beta2.FlowCollectorSpec{
			Processor: v1beta2.FlowCollectorFLP{
				Metrics: v1beta2.FLPMetrics{
					DisableAlerts: []v1beta2.FLPAlert{v1beta2.AlertLokiError},
					IncludeList:   &[]string{"namespace_egress_packets_total", "namespace_flows_total"},
				},
			},
		},
	}

	var converted FlowCollector
	err := converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.Equal([]FLPAlert{AlertLokiError}, converted.Spec.Processor.Metrics.DisableAlerts)
	assert.NotNil(converted.Spec.Processor.Metrics.IncludeList)
	assert.Equal([]string{"namespace_egress_packets_total", "namespace_flows_total"}, *converted.Spec.Processor.Metrics.IncludeList)

	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)
	assert.Equal(initial.Spec.Processor.Metrics, back.Spec.Processor.Metrics)
}
