package alerts

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
)

func TestBuildRules_DefaultWithDisabled(t *testing.T) {
	fc := flowslatest.FlowCollectorSpec{
		Processor: flowslatest.FlowCollectorFLP{
			Metrics: flowslatest.FLPMetrics{
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertTooManyDeviceDrops},
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
		"TooManyKernelDrops_WNamespace",
		"TooManyKernelDrops_INamespace",
		"TooManyKernelDrops_WNode",
		"TooManyKernelDrops_INode",
		"TooManyDeviceDrops_WNode",
		"NetObservNoFlows",
	}, allNames(rules))
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 20% of packets dropped by the kernel [namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"20","unit":"%"}`, rules[0].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [namespace={{ $labels.namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["namespace"],"threshold":"10","unit":"%"}`, rules[1].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[2].Annotations["description"], "NetObserv is detecting more than 10% of packets dropped by the kernel [node={{ $labels.node }}]")
	assert.Contains(t, rules[4].Annotations["description"], "node-exporter is detecting more than 5% of dropped packets [node={{ $labels.instance }}]")
	assert.Contains(t, rules[5].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertTooManyKernelDrops, flowslatest.AlertTooManyDeviceDrops},
			},
		},
	}
	rules := BuildRules(context.Background(), &fc)
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertTooManyDeviceDrops},
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertTooManyKernelDrops,
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
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel [workload={{ $labels.workload }} ({{ $labels.kind }})]")
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertTooManyDeviceDrops},
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertTooManyKernelDrops,
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
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of packets dropped by the kernel.")
	assert.Equal(t, "100 * sum(rate(netobserv_namespace_drop_packets_total[2m]) OR rate(netobserv_workload_drop_packets_total[2m]) OR rate(netobserv_node_drop_packets_total[2m])) / (sum(rate(netobserv_namespace_ingress_packets_total[2m]) OR rate(netobserv_workload_ingress_packets_total[2m]) OR rate(netobserv_node_ingress_packets_total[2m]) OR rate(netobserv_namespace_egress_packets_total[2m]) OR rate(netobserv_workload_egress_packets_total[2m]) OR rate(netobserv_node_egress_packets_total[2m]))) > 50", rules[0].Expr.StrVal)
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: []flowslatest.AlertTemplate{flowslatest.AlertLokiError, flowslatest.AlertTooManyKernelDrops, flowslatest.AlertTooManyDeviceDrops},
				Alerts: &[]flowslatest.FLPAlert{
					{
						Template: flowslatest.AlertTooManyKernelDrops,
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
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
}

func TestAggregateSourceDest(t *testing.T) {
	pql := aggregateSourceDest("rate(my_metric[1m])", flowslatest.GroupByNode)
	assert.Equal(t,
		`(sum(label_replace(rate(my_metric[1m]), "node", "$1", "SrcK8S_HostName", "(.*)")) by (node) + sum(label_replace(rate(my_metric[1m]), "node", "$1", "DstK8S_HostName", "(.*)")) by (node))`,
		pql,
	)

	pql = aggregateSourceDest("rate(my_metric[1m])", flowslatest.GroupByWorkload)
	assert.Equal(t,
		`(sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)"), "workload", "$1", "SrcK8S_OwnerName", "(.*)"), "kind", "$1", "SrcK8S_OwnerType", "(.*)")) by (namespace,workload,kind) + sum(label_replace(label_replace(label_replace(rate(my_metric[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)"), "workload", "$1", "DstK8S_OwnerName", "(.*)"), "kind", "$1", "DstK8S_OwnerType", "(.*)")) by (namespace,workload,kind))`,
		pql,
	)

	pql = aggregateSourceDest("rate(my_metric[1m])", "")
	assert.Equal(t, `sum(rate(my_metric[1m]))`, pql)
}

func TestPercentagePromQL(t *testing.T) {
	pql := percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "", "")
	assert.Equal(t, "100 * sum(rate(my_metric[1m])) / (sum(rate(my_total[1m]))) > 10", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "")
	assert.Equal(t, "100 * sum(rate(my_metric[1m])) / (sum(rate(my_total[1m]))) > 10 < 20", pql)

	pql = percentagePromQL("sum(rate(my_metric[1m]))", "sum(rate(my_total[1m]))", "10", "20", "2")
	assert.Equal(t, "100 * sum(rate(my_metric[1m])) / (sum(rate(my_total[1m])) > 2) > 10 < 20", pql)
}
