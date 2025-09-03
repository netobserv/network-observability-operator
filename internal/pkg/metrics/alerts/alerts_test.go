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
