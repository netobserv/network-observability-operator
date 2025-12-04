package v1beta2

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// Note that we set default in-code rather than in CRD, in order to keep track of value being unset or set intentionnally in FlowCollector
	DefaultIncludeList = []string{
		"node_ingress_bytes_total",
		"node_egress_bytes_total",
		"node_ingress_packets_total",
		"node_drop_packets_total",
		"workload_sampling",
		"workload_ingress_bytes_total",
		"workload_egress_bytes_total",
		"namespace_flows_total",
		"namespace_ingress_packets_total",
		"namespace_drop_packets_total",
		"namespace_rtt_seconds",
		"namespace_dns_latency_seconds",
		"namespace_network_policy_events_total",
		"node_ipsec_flows_total",
		"node_to_node_ingress_flows_total",
	}
	// More metrics enabled when Loki is disabled, to avoid loss of information
	DefaultIncludeListLokiDisabled = []string{
		"node_ingress_bytes_total",
		"node_egress_bytes_total",
		"node_ingress_packets_total",
		"node_drop_packets_total",
		"workload_ingress_bytes_total",
		"workload_egress_bytes_total",
		"workload_sampling",
		"workload_ingress_packets_total",
		"workload_egress_packets_total",
		"workload_flows_total",
		"workload_drop_bytes_total",
		"workload_drop_packets_total",
		"workload_rtt_seconds",
		"workload_dns_latency_seconds",
		"namespace_network_policy_events_total",
		"node_ipsec_flows_total",
		"node_to_node_ingress_flows_total",
	}
	DefaultHealthRules = []FLPHealthRule{
		{
			Template: HealthRulePacketDropsByKernel,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Info:    "10",
						Warning: "20",
					},
					LowVolumeThreshold: "5",
					GroupBy:            GroupByNamespace,
				},
				{
					Thresholds: HealthRuleThresholds{
						Info:    "5",
						Warning: "10",
					},
					GroupBy: GroupByNode,
				},
			},
		},
		{
			Template: HealthRulePacketDropsByDevice,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Warning: "5",
					},
					GroupBy: GroupByNode,
				},
			},
		},
		{
			Template: HealthRuleIPsecErrors,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Critical: "2",
					},
				},
				{
					Thresholds: HealthRuleThresholds{
						Critical: "2",
					},
					GroupBy: GroupByNode,
				},
			},
		},
		{
			Template: HealthRuleDNSErrors,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Warning: "5",
					},
				},
				{
					Thresholds: HealthRuleThresholds{
						Info:    "5",
						Warning: "10",
					},
					GroupBy: GroupByNamespace,
				},
			},
		},
		{
			Template: HealthRuleNetpolDenied,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Info:    "5",
						Warning: "10",
					},
					GroupBy: GroupByNamespace,
				},
			},
		},
		{
			Template: HealthRuleLatencyHighTrend,
			Variants: []HealthRuleVariant{
				{
					Thresholds: HealthRuleThresholds{
						Info: "100",
					},
					GroupBy: GroupByNamespace,
					// TODO: set longer-term defaults
					TrendOffset:   &v1.Duration{Duration: 20 * time.Minute},
					TrendDuration: &v1.Duration{Duration: 20 * time.Minute},
				},
			},
		},
		// TODO: Implement ExternalEgressHighTrend, ExternalIngressHighTrend, CrossAZ templates
	}
)
