package v1beta2

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
	DefaultAlerts = []FLPAlert{
		{
			Template: AlertPacketDropsByKernel,
			Variants: []AlertVariant{
				{
					Thresholds: AlertThresholds{
						Info:    "10",
						Warning: "20",
					},
					LowVolumeThreshold: "5",
					GroupBy:            GroupByNamespace,
				},
				{
					Thresholds: AlertThresholds{
						Info:    "5",
						Warning: "10",
					},
					GroupBy: GroupByNode,
				},
			},
		},
		{
			Template: AlertPacketDropsByNetDev,
			Variants: []AlertVariant{
				{
					Thresholds: AlertThresholds{
						Warning: "5",
					},
					GroupBy: GroupByNode,
				},
			},
		},
	}
)
