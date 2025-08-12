package metrics

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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
	assert.Equal(t, []string{
		"NetObservTooManyDropsWNamespace",
		"NetObservTooManyDropsINamespace",
		"NetObservTooManyDropsWSrcNode",
		"NetObservTooManyDropsISrcNode",
		"NetObservTooManyDropsWDstNode",
		"NetObservTooManyDropsIDstNode",
		"NetObservNoFlows",
	}, allNames(rules))
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 20% of dropped packets from [namespace={{ $labels.SrcK8S_Namespace }}] to [namespace={{ $labels.DstK8S_Namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["SrcK8S_Namespace","DstK8S_Namespace"],"threshold":"20","unit":"%"}`, rules[0].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[1].Annotations["description"], "NetObserv is detecting more than 10% of dropped packets from [namespace={{ $labels.SrcK8S_Namespace }}] to [namespace={{ $labels.DstK8S_Namespace }}]")
	assert.Equal(t, `{"namespaceLabels":["SrcK8S_Namespace","DstK8S_Namespace"],"threshold":"10","unit":"%"}`, rules[1].Annotations["netobserv_io_network_health"])
	assert.Contains(t, rules[2].Annotations["description"], "NetObserv is detecting more than 10% of dropped packets from [node={{ $labels.SrcK8S_HostName }}]")
	assert.Contains(t, rules[4].Annotations["description"], "NetObserv is detecting more than 10% of dropped packets to [node={{ $labels.DstK8S_HostName }}]")
	assert.Contains(t, rules[6].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError, flowslatest.AlertTooManyDrops},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError},
				AlertGroups: &[]flowslatest.FLPAlertGroup{
					{
						Name: flowslatest.AlertTooManyDrops,
						Alerts: []flowslatest.FLPAlert{
							{
								Thresholds: flowslatest.FLPAlertThresholds{
									Critical: "50",
								},
								Grouping:          flowslatest.GroupingPerWorkload,
								GroupingDirection: flowslatest.GroupingBySourceAndDestination,
							},
						},
					},
				},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of dropped packets from [workload={{ $labels.SrcK8S_OwnerName }} ({{ $labels.SrcK8S_OwnerType }})] to [workload={{ $labels.DstK8S_OwnerName }} ({{ $labels.DstK8S_OwnerType }})]")
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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError},
				AlertGroups: &[]flowslatest.FLPAlertGroup{
					{
						Name: flowslatest.AlertTooManyDrops,
						Alerts: []flowslatest.FLPAlert{
							{
								Thresholds: flowslatest.FLPAlertThresholds{
									Critical: "50",
								},
								GroupingDirection: flowslatest.GroupingBySourceAndDestination,
							},
						},
					},
				},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
	assert.Len(t, rules, 2)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv is detecting more than 50% of dropped packets.")
	assert.Equal(t, "100 * sum (rate(netobserv_namespace_drop_packets_total[2m]) OR rate(netobserv_workload_drop_packets_total[2m]) OR rate(netobserv_node_drop_packets_total[2m])) / (sum(rate(netobserv_namespace_ingress_packets_total[2m]) OR rate(netobserv_workload_ingress_packets_total[2m]) OR rate(netobserv_node_ingress_packets_total[2m]) OR rate(netobserv_namespace_egress_packets_total[2m]) OR rate(netobserv_workload_egress_packets_total[2m]) OR rate(netobserv_node_egress_packets_total[2m]))) > 50", rules[0].Expr.StrVal)
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
				DisableAlerts: []flowslatest.FLPAlertGroupName{flowslatest.AlertLokiError, flowslatest.AlertTooManyDrops},
				AlertGroups: &[]flowslatest.FLPAlertGroup{
					{
						Name: flowslatest.AlertTooManyDrops,
						Alerts: []flowslatest.FLPAlert{
							{
								Thresholds: flowslatest.FLPAlertThresholds{
									Critical: "50",
								},
								Grouping:          flowslatest.GroupingPerWorkload,
								GroupingDirection: flowslatest.GroupingBySourceAndDestination,
							},
						},
					},
				},
			},
		},
	}
	rules := BuildAlertRules(context.Background(), &fc)
	assert.Len(t, rules, 1)
	assert.Contains(t, rules[0].Annotations["description"], "NetObserv flowlogs-pipeline is not receiving any flow")
}
