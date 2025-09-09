package v1beta2

import (
	"fmt"
	"slices"
	"strings"
)

type AlertTemplate string
type AlertGroupBy string

const (
	AlertNoFlows             AlertTemplate = "NetObservNoFlows"
	AlertLokiError           AlertTemplate = "NetObservLokiError"
	AlertPacketDropsByKernel AlertTemplate = "PacketDropsByKernel"
	AlertPacketDropsByNetDev AlertTemplate = "PacketDropsByNetDev"
	GroupByNode              AlertGroupBy  = "Node"
	GroupByNamespace         AlertGroupBy  = "Namespace"
	GroupByWorkload          AlertGroupBy  = "Workload"
)

type FLPAlert struct {
	// Alert template name.
	// Possible values are: `PacketDropsByKernel`, `PacketDropsByNetDev`.
	// More information on alerts: https://github.com/netobserv/network-observability-operator/blob/main/docs/Alerts.md
	// +kubebuilder:validation:Enum:="PacketDropsByKernel";"PacketDropsByNetDev"
	// +required
	Template AlertTemplate `json:"template,omitempty"`

	// A list of variants for this template
	// +required
	Variants []AlertVariant `json:"variants,omitempty"`
}

type AlertVariant struct {
	// The low volume threshold allows to ignore metrics with a too low volume of traffic, in order to improve signal-to-noise.
	// It is provided as an absolute rate (bytes per second or packets per second, depending on the context).
	// When provided, it must be parsable as a float.
	LowVolumeThreshold string `json:"lowVolumeThreshold,omitempty"`

	// Thresholds of the alert per severity.
	// They are expressed as a percentage of errors above which the alert is triggered. They must be parsable as floats.
	// +required
	Thresholds AlertThresholds `json:"thresholds,omitempty"`

	// Optional grouping criteria, possible values are: `Node`, `Namespace`, `Workload`.
	// +kubebuilder:validation:Enum:="";"Node";"Namespace";"Workload"
	// +optional
	GroupBy AlertGroupBy `json:"groupBy,omitempty"`
}

type AlertThresholds struct {
	// Threshold for severity `info`. Leave empty to not generate an Info alert.
	Info string `json:"info,omitempty"`

	// Threshold for severity `warning`. Leave empty to not generate a Warning alert.
	Warning string `json:"warning,omitempty"`

	// Threshold for severity `critical`. Leave empty to not generate a Critical alert.
	Critical string `json:"critical,omitempty"`
}

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

func (s *FlowCollectorSpec) GetFLPAlerts() []FLPAlert {
	var ret []FLPAlert
	var templates []AlertTemplate // for reproducible ordering

	tplMap := make(map[AlertTemplate]FLPAlert)
	for _, group := range DefaultAlerts {
		if !slices.Contains(s.Processor.Metrics.DisableAlerts, group.Template) {
			tplMap[group.Template] = group
			templates = append(templates, group.Template)
		}
	}
	if s.Processor.Metrics.Alerts != nil {
		for _, group := range *s.Processor.Metrics.Alerts {
			if !slices.Contains(s.Processor.Metrics.DisableAlerts, group.Template) {
				// A group defined in FC overrides the default group
				tplMap[group.Template] = group
				if !slices.Contains(templates, group.Template) {
					templates = append(templates, group.Template)
				}
			}
		}
	}

	for _, name := range templates {
		tpl := tplMap[name]
		if ok, _ := tpl.IsAllowed(s); ok {
			ret = append(ret, tpl)
		}
	}

	return ret
}

func (g *FLPAlert) IsAllowed(spec *FlowCollectorSpec) (bool, string) {
	switch g.Template {
	case AlertPacketDropsByKernel:
		if !spec.Agent.EBPF.IsPktDropEnabled() {
			return false, fmt.Sprintf("Alert %s requires the %s agent feature to be enabled", AlertPacketDropsByKernel, PacketDrop)
		}
	case AlertNoFlows, AlertLokiError, AlertPacketDropsByNetDev:
		return true, ""
	}
	return true, ""
}
