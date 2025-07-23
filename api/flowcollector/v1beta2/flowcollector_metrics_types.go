package v1beta2

import (
	"fmt"
	"slices"
	"strings"
)

type FLPAlertGroupName string
type AlertSeverity string
type AlertGrouping string
type AlertGroupingDirection string

const (
	AlertNoFlows                   FLPAlertGroupName      = "NetObservNoFlows"
	AlertLokiError                 FLPAlertGroupName      = "NetObservLokiError"
	AlertTooManyDrops              FLPAlertGroupName      = "TooManyDrops"
	SeverityCritical               AlertSeverity          = "Critical"
	SeverityWarning                AlertSeverity          = "Warning"
	SeverityInfo                   AlertSeverity          = "Info"
	GroupingPerNode                AlertGrouping          = "PerNode"
	GroupingPerNamespace           AlertGrouping          = "PerNamespace"
	GroupingPerWorkload            AlertGrouping          = "PerWorkload"
	GroupingBySource               AlertGroupingDirection = "BySource"
	GroupingByDestination          AlertGroupingDirection = "ByDestination"
	GroupingBySourceAndDestination AlertGroupingDirection = "BySourceAndDestination"
)

type FLPAlertGroup struct {
	// Alert group name; TODO: more doc with the list of available alerts, similar to the metrics list.
	// Possible values are:<br>
	// - `NetObservNoFlows`, triggered when no flows are being observed for a certain period.<br>
	// - `NetObservLokiError`, triggered when flows are being dropped due to Loki errors.<br>
	// - `TooManyDrops`, triggered on high percentage of packet drops; it requires the `PacketDrop` agent feature.<br>
	// +kubebuilder:validation:Enum:="NetObservNoFlows";"NetObservLokiError";"TooManyDrops"
	// +required
	Name FLPAlertGroupName `json:"name,omitempty"`

	// A list of alert definitions for the group
	// +required
	Alerts []FLPAlert `json:"alerts,omitempty"`
}

type FLPAlert struct {
	// Alert threshold, as a percentage of errors above which the alert is triggered. It must be parsable as float.
	// +required
	Threshold string `json:"threshold,omitempty"`

	// Severity of an alert, possible values are:<br>
	// - `Critical`<br>
	// - `Warning`<br>
	// - `Info`<br>
	// +kubebuilder:validation:Enum:="Critical";"Warning";"Info"
	Severity AlertSeverity `json:"severity,omitempty"`

	// Optional grouping criteria, possible values are:<br>
	// - `PerNode`<br>
	// - `PerNamespace`<br>
	// - `PerWorkload`<br>
	// +kubebuilder:validation:Enum:="";"PerNode";"PerNamespace";"PerWorkload"
	// +optional
	Grouping AlertGrouping `json:"grouping,omitempty"`

	// Grouping direction, possible values are:<br>
	// - `BySource`<br>
	// - `ByDestination`<br>
	// - `BySourceAndDestination`<br>
	// This setting is ignored when no `grouping` is provided.
	// +kubebuilder:validation:Enum:="ByDestination";"BySource";"BySourceAndDestination"
	// +optional
	GroupingDirection AlertGroupingDirection `json:"groupingDirection,omitempty"`
}

var (
	// Note that we set default in-code rather than in CRD, in order to keep track of value being unset or set intentionnally in FlowCollector
	DefaultIncludeList = []string{
		"node_ingress_bytes_total",
		"node_egress_bytes_total",
		"workload_sampling",
		"workload_ingress_bytes_total",
		"workload_egress_bytes_total",
		"namespace_flows_total",
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
	DefaultAlerts = []FLPAlertGroup{
		{
			Name: AlertTooManyDrops,
			Alerts: []FLPAlert{
				{
					Severity:          SeverityInfo,
					Threshold:         "20",
					Grouping:          GroupingPerNamespace,
					GroupingDirection: GroupingBySourceAndDestination,
				},
				{
					Severity:          SeverityWarning,
					Threshold:         "10",
					Grouping:          GroupingPerNode,
					GroupingDirection: GroupingBySource,
				},
				{
					Severity:          SeverityWarning,
					Threshold:         "10",
					Grouping:          GroupingPerNode,
					GroupingDirection: GroupingByDestination,
				},
			},
		},
	}
)

func (s *FlowCollectorSpec) GetIncludeList() []string {
	var list []string
	if s.Processor.Metrics.IncludeList == nil {
		if s.UseLoki() {
			list = DefaultIncludeList
		} else {
			// When loki is disabled, increase what's available through metrics by default, to minimize the loss of information
			list = DefaultIncludeListLokiDisabled
		}
	} else {
		for _, m := range *s.Processor.Metrics.IncludeList {
			list = append(list, string(m))
		}
	}
	if !s.Agent.EBPF.IsPktDropEnabled() {
		list = removeMetricsByPattern(list, "_drop_")
	}
	if !s.Agent.EBPF.IsFlowRTTEnabled() {
		list = removeMetricsByPattern(list, "_rtt_")
	}
	if !s.Agent.EBPF.IsDNSTrackingEnabled() {
		list = removeMetricsByPattern(list, "_dns_")
	}
	if !s.Agent.EBPF.IsNetworkEventsEnabled() {
		list = removeMetricsByPattern(list, "_network_policy_")
	}
	if !s.HasFiltersSampling() {
		list = removeMetricsByPattern(list, "_sampling")
	}
	if !s.Agent.EBPF.IsIPSecEnabled() {
		list = removeMetricsByPattern(list, "_ipsec_")
	}
	return list
}

func removeMetricsByPattern(list []string, search string) []string {
	var filtered []string
	for _, m := range list {
		if !strings.Contains(m, search) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (s *FlowCollectorSpec) GetFLPAlerts() []FLPAlertGroup {
	var ret []FLPAlertGroup
	var names []FLPAlertGroupName // for reproducible ordering

	groups := make(map[FLPAlertGroupName]FLPAlertGroup)
	for _, group := range DefaultAlerts {
		if !slices.Contains(s.Processor.Metrics.DisableAlerts, group.Name) {
			groups[group.Name] = group
			names = append(names, group.Name)
		}
	}
	if s.Processor.Metrics.AlertGroups != nil {
		for _, group := range *s.Processor.Metrics.AlertGroups {
			// A group defined in FC overrides the default group
			groups[group.Name] = group
			if !slices.Contains(names, group.Name) {
				names = append(names, group.Name)
			}
		}
	}

	for _, name := range names {
		group := groups[name]
		if ok, _ := group.IsAllowed(s); ok {
			ret = append(ret, group)
		}
	}

	return ret
}

func (g *FLPAlertGroup) IsAllowed(spec *FlowCollectorSpec) (bool, string) {
	switch g.Name {
	case AlertTooManyDrops:
		if !spec.Agent.EBPF.IsPktDropEnabled() {
			return false, fmt.Sprintf("Alert %s requires the %s agent feature to be enabled", AlertTooManyDrops, PacketDrop)
		}
	}
	return true, ""
}
