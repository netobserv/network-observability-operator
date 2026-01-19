package alerts

import (
	"context"
	"slices"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
)

func allTemplates() []flowslatest.HealthRuleTemplate {
	return []flowslatest.HealthRuleTemplate{
		flowslatest.AlertLokiError,
		flowslatest.AlertNoFlows,
		flowslatest.HealthRulePacketDropsByKernel,
		flowslatest.HealthRulePacketDropsByDevice,
		flowslatest.HealthRuleDNSErrors,
		flowslatest.HealthRuleDNSNxDomain,
		flowslatest.HealthRuleIPsecErrors,
		flowslatest.HealthRuleLatencyHighTrend,
		flowslatest.HealthRuleNetpolDenied,
		flowslatest.HealthRuleExternalEgressHighTrend,
		flowslatest.HealthRuleExternalIngressHighTrend,
	}
}

func allTemplatesBut(tpls ...flowslatest.AlertTemplate) []flowslatest.AlertTemplate {
	var ret []flowslatest.AlertTemplate
	for _, tpl := range allTemplates() {
		if !slices.Contains(tpls, tpl) {
			ret = append(ret, tpl)
		}
	}
	return ret
}

func TestBuildRules_DefaultWithDisabled(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: []flowslatest.AlertTemplate{
					flowslatest.AlertLokiError,
					flowslatest.HealthRulePacketDropsByDevice,
					flowslatest.HealthRuleExternalEgressHighTrend,
					flowslatest.HealthRuleExternalIngressHighTrend,
				},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
}

func allNames(rules []monitoringv1.Rule) []string {
	var names []string
	for _, r := range rules {
		names = append(names, r.Alert)
	}
	return names
}

func TestBuildRules_DefaultWithFeaturesAndDisabled(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features: []flowslatest.AgentFeature{
					flowslatest.FlowRTT,
					flowslatest.DNSTracking,
					flowslatest.IPSec,
					flowslatest.NetworkEvents,
					flowslatest.PacketDrop,
				},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: []flowslatest.AlertTemplate{
					flowslatest.AlertLokiError,
					flowslatest.HealthRuleExternalEgressHighTrend,
					flowslatest.HealthRuleExternalIngressHighTrend,
				},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Equal(t, []string{
		"PacketDropsByKernel_PerSrcNamespaceWarning",
		"PacketDropsByKernel_PerDstNamespaceWarning",
		"PacketDropsByKernel_PerSrcNamespaceInfo",
		"PacketDropsByKernel_PerDstNamespaceInfo",
		"PacketDropsByKernel_PerSrcNodeWarning",
		"PacketDropsByKernel_PerDstNodeWarning",
		"PacketDropsByKernel_PerSrcNodeInfo",
		"PacketDropsByKernel_PerDstNodeInfo",
		"PacketDropsByDevice_PerNodeWarning",
		"IPsecErrors_Critical",
		"IPsecErrors_PerSrcNodeCritical",
		"IPsecErrors_PerDstNodeCritical",
		"DNSErrors_Warning",
		"DNSErrors_PerDstNamespaceWarning",
		"DNSErrors_PerDstNamespaceInfo",
		"DNSNxDomain_PerDstNamespaceWarning",
		"DNSNxDomain_PerDstNamespaceInfo",
		"NetpolDenied_PerSrcNamespaceWarning",
		"NetpolDenied_PerDstNamespaceWarning",
		"NetpolDenied_PerSrcNamespaceInfo",
		"NetpolDenied_PerDstNamespaceInfo",
		"LatencyHighTrend_PerSrcNamespaceInfo",
		"LatencyHighTrend_PerDstNamespaceInfo",
		"NetObservNoFlows",
	}, allNames(rules))
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 20% of packets dropped by the kernel [source namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"20","unit":"%"}`, rules[0].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[3].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [dest. namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"10","unit":"%"}`, rules[3].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[4].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [source node={{ $labels.node }}]")
	assert.Contains(t, rules[8].Annotations["description"], "node-exporter is detecting more than 5% of dropped packets [node={{ $labels.instance }}]")
	assert.Contains(t, rules[len(rules)-1].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
}

func TestBuildRules_DefaultWithFeaturesAndAllDisabled(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features:   []flowslatest.AgentFeature{flowslatest.FlowRTT, flowslatest.DNSTracking, flowslatest.IPSec, flowslatest.NetworkEvents, flowslatest.PacketDrop},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: allTemplates(),
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Empty(t, rules)
}

func TestBuildRules_Overidden(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features:   []flowslatest.AgentFeature{flowslatest.FlowRTT, flowslatest.DNSTracking, flowslatest.IPSec, flowslatest.NetworkEvents, flowslatest.PacketDrop},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: allTemplatesBut(flowslatest.HealthRulePacketDropsByKernel),
				HealthRules: &[]flowslatest.FLPHealthRule{
					{
						Template: flowslatest.HealthRulePacketDropsByKernel,
						Variants: []flowslatest.HealthRuleVariant{
							{
								Thresholds: flowslatest.HealthRuleThresholds{
									Critical: "50",
								},
								GroupBy: flowslatest.GroupByWorkload,
							},
						},
					},
				},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel [source workload={{ $labels.workload }} ({{ $labels.kind }})]")
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel [dest. workload={{ $labels.workload }} ({{ $labels.kind }})]")
}

func TestBuildRules_Global(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features:   []flowslatest.AgentFeature{flowslatest.FlowRTT, flowslatest.DNSTracking, flowslatest.IPSec, flowslatest.NetworkEvents, flowslatest.PacketDrop},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: allTemplatesBut(flowslatest.HealthRulePacketDropsByKernel),
				HealthRules: &[]flowslatest.FLPHealthRule{
					{
						Template: flowslatest.HealthRulePacketDropsByKernel,
						Variants: []flowslatest.HealthRuleVariant{
							{
								Thresholds: flowslatest.HealthRuleThresholds{
									Critical: "50",
								},
							},
						},
					},
				},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel.")
	assert.Equal(t, "100 * (sum(rate(netobserv_namespace_drop_packets_total[2m]))) / (sum(rate(netobserv_namespace_ingress_packets_total[2m]))) > 50", rules[0].Expr.StrVal)
}

func TestBuildRules_DisableTakesPrecedence(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features:   []flowslatest.AgentFeature{flowslatest.FlowRTT, flowslatest.DNSTracking, flowslatest.IPSec, flowslatest.NetworkEvents, flowslatest.PacketDrop},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: allTemplates(),
				HealthRules: &[]flowslatest.FLPHealthRule{
					{
						Template: flowslatest.HealthRulePacketDropsByKernel,
						Variants: []flowslatest.HealthRuleVariant{
							{
								Thresholds: flowslatest.HealthRuleThresholds{
									Critical: "50",
								},
								GroupBy: flowslatest.GroupByWorkload,
							},
						},
					},
				},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Empty(t, rules)
}

func TestLatencyPromql(t *testing.T) {
	variant := flowslatest.HealthRuleVariant{
		GroupBy: flowslatest.GroupByNamespace,
		Thresholds: flowslatest.HealthRuleThresholds{
			Info: "100",
		},
	}
	rules, err := convertToRules(flowslatest.HealthRuleLatencyHighTrend, flowslatest.ModeAlert, &variant, []string{"namespace_rtt_seconds"})
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting TCP latency increased by more than 100% [source namespace={{ $labels.namespace }}], compared to baseline (offset: 24h).")
	// The pattern is:
	// 100 * (<current latency> - <past latency>) / <past latency>
	assert.Equal(t,
		`100 * `+
			`((histogram_quantile(0.9, `+
			`sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket{SrcK8S_Namespace!=""}[2m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)))`+
			` - (histogram_quantile(0.9, `+
			`sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket{SrcK8S_Namespace!=""}[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le))))`+
			` / (histogram_quantile(0.9, `+
			`sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket{SrcK8S_Namespace!=""}[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)))`+
			` > 100`,
		rules[0].Expr.StrVal,
	)
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"100","unit":"%","upperBound":"500"}`, rules[0].Annotations["netobserv_io_network_health"])
}

func TestAllAlertsHaveRunbookURL(t *testing.T) {
	// Create a FlowCollector with all features enabled
	fc := flowslatest.FlowCollectorSpec{
		Agent: flowslatest.FlowCollectorAgent{
			EBPF: flowslatest.FlowCollectorEBPF{
				Privileged: true,
				Features: []flowslatest.AgentFeature{
					flowslatest.FlowRTT,
					flowslatest.DNSTracking,
					flowslatest.IPSec,
					flowslatest.NetworkEvents,
					flowslatest.PacketDrop,
				},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{},
		},
	}

	rules := BuildRules(context.Background(), &fc)

	// Verify all rules have a runbook_url annotation
	for _, rule := range rules {
		url, exists := rule.Annotations["runbook_url"]
		assert.True(t, exists, "Alert %s is missing runbook_url annotation", rule.Alert)
		assert.NotEmpty(t, url, "Alert %s has empty runbook_url", rule.Alert)
		assert.Contains(t, url, "https://github.com/netobserv/runbooks/blob/main/alerts/network-observability-operator/",
			"Alert %s has invalid runbook_url: %s", rule.Alert, url)
		assert.Contains(t, url, ".md", "Alert %s runbook_url doesn't end with .md: %s", rule.Alert, url)
	}
}
