package alerts

import (
	"context"
	"slices"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
)

func allTemplates() []flowslatest.AlertTemplate {
	return []flowslatest.AlertTemplate{
		flowslatest.AlertLokiError,
		flowslatest.AlertNoFlows,
		flowslatest.AlertPacketDropsByKernel,
		flowslatest.AlertPacketDropsByDevice,
		flowslatest.AlertDNSErrors,
		flowslatest.AlertIPsecErrors,
		flowslatest.AlertLatencyHighTrend,
		flowslatest.AlertNetpolDenied,
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
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertPacketDropsByDevice},
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
				Features:   []flowslatest.AgentFeature{flowslatest.FlowRTT, flowslatest.DNSTracking, flowslatest.IPSec, flowslatest.NetworkEvents, flowslatest.PacketDrop},
			},
		},
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Equal(t, []string{
		"PacketDropsByKernel_PerNamespaceWarning",
		"PacketDropsByKernel_PerNamespaceInfo",
		"PacketDropsByKernel_PerNodeWarning",
		"PacketDropsByKernel_PerNodeInfo",
		"PacketDropsByDevice_PerNodeWarning",
		"IPsecErrors_Critical",
		"IPsecErrors_PerNodeCritical",
		"DNSErrors_Warning",
		"DNSErrors_PerNamespaceWarning",
		"DNSErrors_PerNamespaceInfo",
		"NetpolDenied_PerNamespaceWarning",
		"NetpolDenied_PerNamespaceInfo",
		"LatencyHighTrend_PerNamespaceInfo",
		"NetObservNoFlows",
	}, allNames(rules))
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 20% of packets dropped by the kernel [namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"20","unit":"%"}`, rules[0].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"10","unit":"%"}`, rules[1].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[2].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [node={{ $labels.node }}]")
	assert.Contains(t, rules[4].Annotations["description"], "node-exporter is detecting more than 5% of dropped packets [node={{ $labels.instance }}]")
	assert.Contains(t, rules[13].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: allTemplatesBut(flowslatest.AlertPacketDropsByKernel),
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertPacketDropsByKernel,
						Variants: []flowslatest.AlertVariant{
							{
								Thresholds: flowslatest.AlertThresholds{
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
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel [workload={{ $labels.workload }} ({{ $labels.kind }})]")
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
				DisableAlerts: allTemplatesBut(flowslatest.AlertPacketDropsByKernel),
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertPacketDropsByKernel,
						Variants: []flowslatest.AlertVariant{
							{
								Thresholds: flowslatest.AlertThresholds{
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
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertPacketDropsByKernel,
						Variants: []flowslatest.AlertVariant{
							{
								Thresholds: flowslatest.AlertThresholds{
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

func TestAggregateSourceDest(t *testing.T) {
	pql := aggregateSourceDest("rate(my_metric[1m])", flowslatest.GroupByNode, "")
	assert.Equal(t,
		`(sum(label_replace(rate(my_metric[1m]), "node", "$1", "SrcK8S_HostName", "(.*)")) by (node)`+
			` + sum(label_replace(rate(my_metric[1m]), "node", "$1", "DstK8S_HostName", "(.*)")) by (node))`+
			` OR sum(label_replace(rate(my_metric[1m]), "node", "$1", "SrcK8S_HostName", "(.*)")) by (node)`+
			` OR sum(label_replace(rate(my_metric[1m]), "node", "$1", "DstK8S_HostName", "(.*)")) by (node)`,
		pql,
	)

	pql = aggregateSourceDest("rate(my_metric[1m])", flowslatest.GroupByWorkload, "")
	assert.Equal(t,
		`(sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)"), "workload", "$1", "SrcK8S_OwnerName", "(.*)"), "kind", "$1", "SrcK8S_OwnerType", "(.*)")) by (namespace,workload,kind)`+
			` + sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)"), "workload", "$1", "DstK8S_OwnerName", "(.*)"), "kind", "$1", "DstK8S_OwnerType", "(.*)")) by (namespace,workload,kind))`+
			` OR sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)"), "workload", "$1", "SrcK8S_OwnerName", "(.*)"), "kind", "$1", "SrcK8S_OwnerType", "(.*)")) by (namespace,workload,kind)`+
			` OR sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)"), "workload", "$1", "DstK8S_OwnerName", "(.*)"), "kind", "$1", "DstK8S_OwnerType", "(.*)")) by (namespace,workload,kind)`,
		pql,
	)

	pql = aggregateSourceDest("rate(my_metric[1m])", "", "")
	assert.Equal(t, `sum(rate(my_metric[1m]))`, pql)
}

func TestPercentagePromQL(t *testing.T) {
	pql := percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m]))) > 10 < 20", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "2")
	assert.Equal(t, "100 * (sum(rate(my_metric[1m]))) / (sum(rate(my_total[1m])) > 2) > 10 < 20", pql)
}

func TestLatencyPromql(t *testing.T) {
	variant := flowslatest.AlertVariant{
		GroupBy: flowslatest.GroupByNamespace,
		Thresholds: flowslatest.AlertThresholds{
			Info: "100",
		},
	}
	rules, err := convertToRules(flowslatest.AlertLatencyHighTrend, &variant, []string{"namespace_rtt_seconds"})
	assert.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting TCP latency increased by more than 100% [namespace={{ $labels.namespace }}], compared to baseline (offset: 24h).")
	// Pretty dense query, which stands for:
	// 100 * (<current latency> - <past latency>) / <past latency>
	// Both <current latency> and <past latency> are:
	// (Latency by source namespace + Latency by dest namespace) OR Latency by source namespace OR Latency by dest namespace
	// (this "OR" trick is necessary, otherwise prometheus eliminates data that isn't retrieved both as source and destination)
	assert.Equal(t,
		`100 * `+
			`((histogram_quantile(0.9, `+
			`(sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` + sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2m]), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le))`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2m]), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le)))`+
			` - (histogram_quantile(0.9, `+
			`(sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` + sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le))`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le))))`+
			` / (histogram_quantile(0.9, `+
			`(sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` + sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le))`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace,le)`+
			` OR sum(label_replace(rate(netobserv_namespace_rtt_seconds_bucket[2h] offset 24h), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace,le)))`+
			` > 100`,
		rules[0].Expr.StrVal,
	)
}
