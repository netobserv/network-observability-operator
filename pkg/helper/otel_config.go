package helper

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
)

// openTelemetryDefaultTransformRules defined the default Open Telemetry format
// See https://github.com/rhobs/observability-data-model/blob/main/network-observability.md#format-proposal
var openTelemetryDefaultTransformRules = []api.GenericTransformRule{{
	Input:  "SrcAddr",
	Output: "source.address",
}, {
	Input:  "SrcMac",
	Output: "source.mac",
}, {
	Input:  "SrcHostIP",
	Output: "source.host.address",
}, {
	Input:  "SrcK8S_HostName",
	Output: "source.k8s.node.name",
}, {
	Input:  "SrcPort",
	Output: "source.port",
}, {
	Input:  "SrcK8S_Name",
	Output: "source.k8s.name",
}, {
	Input:  "SrcK8S_Type",
	Output: "source.k8s.kind",
}, {
	Input:  "SrcK8S_OwnerName",
	Output: "source.k8s.owner.name",
}, {
	Input:  "SrcK8S_OwnerType",
	Output: "source.k8s.owner.kind",
}, {
	Input:  "SrcK8S_Namespace",
	Output: "source.k8s.namespace.name",
}, {
	Input:  "SrcK8S_HostIP",
	Output: "source.k8s.host.address",
}, {
	Input:  "SrcK8S_HostName",
	Output: "source.k8s.host.name",
}, {
	Input:  "SrcK8S_Zone",
	Output: "source.zone",
}, {
	Input:  "DstAddr",
	Output: "destination.address",
}, {
	Input:  "DstMac",
	Output: "destination.mac",
}, {
	Input:  "DstHostIP",
	Output: "destination.host.address",
}, {
	Input:  "DstK8S_HostName",
	Output: "destination.k8s.node.name",
}, {
	Input:  "DstPort",
	Output: "destination.port",
}, {
	Input:  "DstK8S_Name",
	Output: "destination.k8s.name",
}, {
	Input:  "DstK8S_Type",
	Output: "destination.k8s.kind",
}, {
	Input:  "DstK8S_OwnerName",
	Output: "destination.k8s.owner.name",
}, {
	Input:  "DstK8S_OwnerType",
	Output: "destination.k8s.owner.kind",
}, {
	Input:  "DstK8S_Namespace",
	Output: "destination.k8s.namespace.name",
}, {
	Input:  "DstK8S_HostIP",
	Output: "destination.k8s.host.address",
}, {
	Input:  "DstK8S_HostName",
	Output: "destination.k8s.host.name",
}, {
	Input:  "DstK8S_Zone",
	Output: "destination.zone",
}, {
	Input:  "Bytes",
	Output: "bytes",
}, {
	Input:  "Packets",
	Output: "packets",
}, {
	Input:  "Proto",
	Output: "protocol",
}, {
	Input:  "Flags",
	Output: "tcp.flags",
}, {
	Input:  "TimeFlowRttNs",
	Output: "tcp.rtt",
}, {
	Input:  "Interfaces",
	Output: "interface.names",
}, {
	Input:  "IfDirections",
	Output: "interface.directions",
}, {
	Input:  "FlowDirection",
	Output: "host.direction",
}, {
	Input:  "DnsErrno",
	Output: "dns.errno",
}, {
	Input:  "DnsFlags",
	Output: "dns.flags",
}, {
	Input:  "DnsFlagsResponseCode",
	Output: "dns.responsecode",
}, {
	Input:  "DnsId",
	Output: "dns.id",
}, {
	Input:  "DnsLatencyMs",
	Output: "dns.latency",
}, {
	Input:  "Dscp",
	Output: "dscp",
}, {
	Input:  "IcmpCode",
	Output: "icmp.code",
}, {
	Input:  "IcmpType",
	Output: "icmp.type",
}, {
	Input:  "K8S_ClusterName",
	Output: "k8s.cluster.name",
}, {
	Input:  "K8S_FlowLayer",
	Output: "k8s.layer",
}, {
	Input:  "PktDropBytes",
	Output: "drops.bytes",
}, {
	Input:  "PktDropPackets",
	Output: "drops.packets",
}, {
	Input:  "PktDropLatestDropCause",
	Output: "drops.latestcause",
}, {
	Input:  "PktDropLatestFlags",
	Output: "drops.latestflags",
}, {
	Input:  "PktDropLatestState",
	Output: "drops.lateststate",
}, {
	Input:  "TimeFlowEndMs",
	Output: "timeflowend",
}, {
	Input:  "TimeFlowStartMs",
	Output: "timeflowstart",
}, {
	Input:  "TimeReceived",
	Output: "timereceived",
}}

func GetOtelTransformConfig(rules *[]flowslatest.GenericTransformRule) api.TransformGeneric {
	transformConfig := api.TransformGeneric{
		Policy: "replace_keys",
		Rules:  openTelemetryDefaultTransformRules,
	}
	// set custom rules if specified
	if rules != nil {
		transformConfig.Rules = []api.GenericTransformRule{}
		for _, r := range *rules {
			transformConfig.Rules = append(transformConfig.Rules, api.GenericTransformRule{
				Input:      r.Input,
				Output:     r.Output,
				Multiplier: r.Multiplier,
			})
		}
	}

	return transformConfig
}

func GetOtelMetrics(flpMetrics []api.MetricsItem) []api.MetricsItem {
	var otelMetrics = []api.MetricsItem{}

	for i := range flpMetrics {
		m := flpMetrics[i]

		otelMetrics = append(otelMetrics, api.MetricsItem{
			Name:       convertToOtelLabel(m.Name),
			Type:       m.Type,
			Filters:    convertToOtelFilters(m.Filters),
			ValueKey:   convertToOtelLabel(m.ValueKey),
			Labels:     convertToOtelLabels(m.Labels),
			Buckets:    m.Buckets,
			ValueScale: m.ValueScale,
		})
	}

	return otelMetrics
}

func convertToOtelLabel(input string) string {
	for _, tr := range openTelemetryDefaultTransformRules {
		if tr.Input == input {
			return tr.Output
		}
	}

	return input
}

func convertToOtelFilters(filters []api.MetricsFilter) []api.MetricsFilter {
	var otelFilters = []api.MetricsFilter{}

	for _, f := range filters {
		otelFilters = append(otelFilters, api.MetricsFilter{
			Key:   convertToOtelLabel(f.Key),
			Value: f.Value,
			Type:  f.Type,
		})
	}

	return otelFilters
}

func convertToOtelLabels(labels []string) []string {
	var otelLabels = []string{}

	for _, l := range labels {
		otelLabels = append(otelLabels, convertToOtelLabel(l))
	}

	return otelLabels
}
