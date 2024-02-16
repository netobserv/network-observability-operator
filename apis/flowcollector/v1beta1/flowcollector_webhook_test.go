package v1beta1

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/pkg/helper"
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
				Timeout:    ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MinBackoff: ptr.To(metav1.Duration{Duration: 5 * time.Second}),
				BatchSize:  1000,
				BatchWait:  ptr.To(metav1.Duration{Duration: 10 * time.Second}),
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
	assert.Equal(v1beta2.LokiAuthForwardUserToken, converted.Spec.Loki.Manual.AuthToken)
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

	expectedDefaultMetrics := []v1beta2.FLPMetric{"namespace_egress_packets_total", "namespace_flows_total", "namespace_rtt_seconds", "namespace_drop_packets_total", "namespace_dns_latency_seconds"}
	assert.Equal([]v1beta2.FLPAlert{v1beta2.AlertLokiError}, converted.Spec.Processor.Metrics.DisableAlerts)
	assert.NotNil(converted.Spec.Processor.Metrics.IncludeList)
	assert.Equal(expectedDefaultMetrics, *converted.Spec.Processor.Metrics.IncludeList)

	// Other way
	var back FlowCollector
	err = back.ConvertFrom(&converted)
	assert.NoError(err)
	// Here, includeList is preserved; it takes precedence over ignoreTags
	var expectedBeta1 []FLPMetric
	for _, m := range expectedDefaultMetrics {
		expectedBeta1 = append(expectedBeta1, FLPMetric(m))
	}
	assert.Equal(expectedBeta1, *back.Spec.Processor.Metrics.IncludeList)
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
					IncludeList:   &[]v1beta2.FLPMetric{"namespace_egress_packets_total", "namespace_flows_total"},
				},
			},
		},
	}

	var converted FlowCollector
	err := converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.Equal([]FLPAlert{AlertLokiError}, converted.Spec.Processor.Metrics.DisableAlerts)
	assert.NotNil(converted.Spec.Processor.Metrics.IncludeList)
	assert.Equal([]FLPMetric{"namespace_egress_packets_total", "namespace_flows_total"}, *converted.Spec.Processor.Metrics.IncludeList)

	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)
	assert.Equal(initial.Spec.Processor.Metrics, back.Spec.Processor.Metrics)
}

func TestBeta2ConversionRoundtrip_Metrics_Default(t *testing.T) {
	// Testing beta2 -> beta1 -> beta2
	assert := assert.New(t)

	initial := v1beta2.FlowCollector{
		Spec: v1beta2.FlowCollectorSpec{
			Processor: v1beta2.FlowCollectorFLP{
				Metrics: v1beta2.FLPMetrics{
					DisableAlerts: []v1beta2.FLPAlert{v1beta2.AlertLokiError},
				},
			},
		},
	}

	var converted FlowCollector
	err := converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.Empty(converted.Spec.Processor.Metrics.IgnoreTags)
	assert.Nil(converted.Spec.Processor.Metrics.IncludeList)

	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)
	assert.Nil(back.Spec.Processor.Metrics.IncludeList)
}

func TestBeta1ConversionRoundtrip_DefaultAdvanced(t *testing.T) {
	// Testing beta1 -> beta2 -> beta1
	assert := assert.New(t)
	err := helper.SetCRDForTests("../../..")
	assert.NoError(err)

	initial := FlowCollector{
		Spec: FlowCollectorSpec{
			Processor: FlowCollectorFLP{},
			ConsolePlugin: FlowCollectorConsolePlugin{
				Register: ptr.To(true),
			},
			Loki: FlowCollectorLoki{
				MaxRetries: ptr.To(int32(2)),
			},
		},
	}

	var converted v1beta2.FlowCollector
	err = initial.ConvertTo(&converted)
	assert.NoError(err)

	assert.Nil(converted.Spec.ConsolePlugin.Advanced)
	assert.Nil(converted.Spec.Processor.Advanced)
	assert.Nil(converted.Spec.Agent.EBPF.Advanced)
	assert.Nil(converted.Spec.Loki.Advanced)

	pluginAdvanced := helper.GetAdvancedPluginConfig(converted.Spec.ConsolePlugin.Advanced)
	lokiAdvanced := helper.GetAdvancedLokiConfig(converted.Spec.Loki.Advanced)
	assert.True(*pluginAdvanced.Register)
	assert.Equal(int32(9001), *pluginAdvanced.Port)
	assert.Equal(int32(2), *lokiAdvanced.WriteMaxRetries)

	// Other way
	var back FlowCollector
	err = back.ConvertFrom(&converted)
	assert.NoError(err)

	assert.Nil(back.Spec.ConsolePlugin.Register)
	assert.Nil(back.Spec.Loki.MaxRetries)
}

func TestBeta2ConversionRoundtrip_DefaultAdvanced(t *testing.T) {
	// Testing beta2 -> beta1 -> beta2
	assert := assert.New(t)
	err := helper.SetCRDForTests("../../..")
	assert.NoError(err)

	initial := v1beta2.FlowCollector{
		Spec: v1beta2.FlowCollectorSpec{
			ConsolePlugin: v1beta2.FlowCollectorConsolePlugin{
				Advanced: &v1beta2.AdvancedPluginConfig{
					Register: ptr.To(true),
				},
			},
			Loki: v1beta2.FlowCollectorLoki{
				Advanced: &v1beta2.AdvancedLokiConfig{
					WriteMaxRetries: ptr.To(int32(2)),
				},
			},
		},
	}

	var converted FlowCollector
	err = converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.True(*converted.Spec.ConsolePlugin.Register)
	assert.Equal(int32(2), *converted.Spec.Loki.MaxRetries)

	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)

	assert.Nil(back.Spec.ConsolePlugin.Advanced)
	assert.Nil(back.Spec.Processor.Advanced)
	assert.Nil(back.Spec.Agent.EBPF.Advanced)
	assert.Nil(back.Spec.Loki.Advanced)
}

func TestBeta1ConversionRoundtrip_Advanced(t *testing.T) {
	// Testing beta1 -> beta2 -> beta1
	assert := assert.New(t)
	err := helper.SetCRDForTests("../../..")
	assert.NoError(err)

	initial := FlowCollector{
		Spec: FlowCollectorSpec{
			Processor: FlowCollectorFLP{
				HealthPort:                     999,
				ProfilePort:                    998,
				ConversationEndTimeout:         &metav1.Duration{Duration: time.Second},
				ConversationHeartbeatInterval:  &metav1.Duration{Duration: time.Minute},
				ConversationTerminatingTimeout: &metav1.Duration{Duration: time.Hour},
			},
			ConsolePlugin: FlowCollectorConsolePlugin{
				Register: ptr.To(false),
				Port:     1000,
			},
			Loki: FlowCollectorLoki{
				BatchWait:  &metav1.Duration{Duration: time.Minute},
				MinBackoff: &metav1.Duration{Duration: time.Minute},
				MaxBackoff: &metav1.Duration{Duration: time.Hour},
				MaxRetries: ptr.To(int32(10)),
			},
		},
	}

	var converted v1beta2.FlowCollector
	err = initial.ConvertTo(&converted)
	assert.NoError(err)

	assert.False(*converted.Spec.ConsolePlugin.Advanced.Register)
	assert.Equal(int32(1000), *converted.Spec.ConsolePlugin.Advanced.Port)
	assert.Equal(int32(999), *converted.Spec.Processor.Advanced.HealthPort)
	assert.Equal(int32(998), *converted.Spec.Processor.Advanced.ProfilePort)
	assert.Equal(time.Second, converted.Spec.Processor.Advanced.ConversationEndTimeout.Duration)
	assert.Equal(time.Minute, converted.Spec.Processor.Advanced.ConversationHeartbeatInterval.Duration)
	assert.Equal(time.Hour, converted.Spec.Processor.Advanced.ConversationTerminatingTimeout.Duration)
	assert.Equal(time.Minute, converted.Spec.Loki.WriteBatchWait.Duration)
	assert.Equal(time.Minute, converted.Spec.Loki.Advanced.WriteMinBackoff.Duration)
	assert.Equal(time.Hour, converted.Spec.Loki.Advanced.WriteMaxBackoff.Duration)
	assert.Equal(int32(10), *converted.Spec.Loki.Advanced.WriteMaxRetries)

	// Other way
	var back FlowCollector
	err = back.ConvertFrom(&converted)
	assert.NoError(err)

	assert.False(*back.Spec.ConsolePlugin.Register)
	assert.Equal(int32(1000), back.Spec.ConsolePlugin.Port)
	assert.Equal(int32(999), back.Spec.Processor.HealthPort)
	assert.Equal(int32(998), back.Spec.Processor.ProfilePort)
	assert.Equal(time.Second, back.Spec.Processor.ConversationEndTimeout.Duration)
	assert.Equal(time.Minute, back.Spec.Processor.ConversationHeartbeatInterval.Duration)
	assert.Equal(time.Hour, back.Spec.Processor.ConversationTerminatingTimeout.Duration)
	assert.Equal(time.Minute, back.Spec.Loki.BatchWait.Duration)
	assert.Equal(time.Minute, back.Spec.Loki.MinBackoff.Duration)
	assert.Equal(time.Hour, back.Spec.Loki.MaxBackoff.Duration)
	assert.Equal(int32(10), *back.Spec.Loki.MaxRetries)
}

func TestBeta2ConversionRoundtrip_Advanced(t *testing.T) {
	// Testing beta2 -> beta1 -> beta2
	assert := assert.New(t)
	err := helper.SetCRDForTests("../../..")
	assert.NoError(err)

	affinityExample := v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values: []string{
									"ok",
								},
							},
						},
					},
				},
			},
		},
	}

	initial := v1beta2.FlowCollector{
		Spec: v1beta2.FlowCollectorSpec{
			Agent: v1beta2.FlowCollectorAgent{
				EBPF: v1beta2.FlowCollectorEBPF{
					Advanced: &v1beta2.AdvancedAgentConfig{
						PriorityClassName: "pcn",
					},
				},
			},
			Processor: v1beta2.FlowCollectorFLP{
				Advanced: &v1beta2.AdvancedProcessorConfig{
					HealthPort:                     ptr.To(int32(999)),
					ProfilePort:                    ptr.To(int32(998)),
					ConversationEndTimeout:         &metav1.Duration{Duration: time.Second},
					ConversationHeartbeatInterval:  &metav1.Duration{Duration: time.Minute},
					ConversationTerminatingTimeout: &metav1.Duration{Duration: time.Hour},
					NodeSelector:                   map[string]string{"test": "ok"},
				},
			},
			ConsolePlugin: v1beta2.FlowCollectorConsolePlugin{
				Advanced: &v1beta2.AdvancedPluginConfig{
					Register: ptr.To(false),
					Port:     ptr.To(int32(1000)),
					Affinity: &affinityExample,
				},
			},
			Loki: v1beta2.FlowCollectorLoki{
				WriteBatchWait: &metav1.Duration{Duration: time.Minute},
				Advanced: &v1beta2.AdvancedLokiConfig{
					WriteMinBackoff: &metav1.Duration{Duration: time.Minute},
					WriteMaxBackoff: &metav1.Duration{Duration: time.Hour},
					WriteMaxRetries: ptr.To(int32(10)),
				},
			},
		},
	}

	var converted FlowCollector
	err = converted.ConvertFrom(&initial)
	assert.NoError(err)

	assert.False(*converted.Spec.ConsolePlugin.Register)
	assert.Equal(int32(1000), converted.Spec.ConsolePlugin.Port)
	assert.Equal(int32(999), converted.Spec.Processor.HealthPort)
	assert.Equal(int32(998), converted.Spec.Processor.ProfilePort)
	assert.Equal(time.Second, converted.Spec.Processor.ConversationEndTimeout.Duration)
	assert.Equal(time.Minute, converted.Spec.Processor.ConversationHeartbeatInterval.Duration)
	assert.Equal(time.Hour, converted.Spec.Processor.ConversationTerminatingTimeout.Duration)
	assert.Equal(time.Minute, converted.Spec.Loki.BatchWait.Duration)
	assert.Equal(time.Minute, converted.Spec.Loki.MinBackoff.Duration)
	assert.Equal(time.Hour, converted.Spec.Loki.MaxBackoff.Duration)
	assert.Equal(int32(10), *converted.Spec.Loki.MaxRetries)

	var back v1beta2.FlowCollector
	err = converted.ConvertTo(&back)
	assert.NoError(err)

	assert.Equal("pcn", back.Spec.Agent.EBPF.Advanced.PriorityClassName)
	assert.False(*back.Spec.ConsolePlugin.Advanced.Register)
	assert.Equal(int32(1000), *back.Spec.ConsolePlugin.Advanced.Port)
	assert.Equal(&affinityExample, back.Spec.ConsolePlugin.Advanced.Affinity)
	assert.Equal(int32(999), *back.Spec.Processor.Advanced.HealthPort)
	assert.Equal(int32(998), *back.Spec.Processor.Advanced.ProfilePort)
	assert.Equal(time.Second, back.Spec.Processor.Advanced.ConversationEndTimeout.Duration)
	assert.Equal(time.Minute, back.Spec.Processor.Advanced.ConversationHeartbeatInterval.Duration)
	assert.Equal(time.Hour, back.Spec.Processor.Advanced.ConversationTerminatingTimeout.Duration)
	assert.Equal(map[string]string{"test": "ok"}, back.Spec.Processor.Advanced.NodeSelector)
	assert.Equal(time.Minute, back.Spec.Loki.WriteBatchWait.Duration)
	assert.Equal(time.Minute, back.Spec.Loki.Advanced.WriteMinBackoff.Duration)
	assert.Equal(time.Hour, back.Spec.Loki.Advanced.WriteMaxBackoff.Duration)
	assert.Equal(int32(10), *back.Spec.Loki.Advanced.WriteMaxRetries)
}
