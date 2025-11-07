package v1beta2

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AlertTemplate string
type AlertGroupBy string
type HealthRuleMode string

const (
	AlertNoFlows                  AlertTemplate = "NetObservNoFlows"
	AlertLokiError                AlertTemplate = "NetObservLokiError"
	AlertPacketDropsByKernel      AlertTemplate = "PacketDropsByKernel"
	AlertPacketDropsByDevice      AlertTemplate = "PacketDropsByDevice"
	AlertIPsecErrors              AlertTemplate = "IPsecErrors"
	AlertNetpolDenied             AlertTemplate = "NetpolDenied"
	AlertLatencyHighTrend         AlertTemplate = "LatencyHighTrend"
	AlertDNSErrors                AlertTemplate = "DNSErrors"
	AlertExternalEgressHighTrend  AlertTemplate = "ExternalEgressHighTrend"
	AlertExternalIngressHighTrend AlertTemplate = "ExternalIngressHighTrend"
	AlertCrossAZ                  AlertTemplate = "CrossAZ"
	GroupByNode                   AlertGroupBy  = "Node"
	GroupByNamespace              AlertGroupBy  = "Namespace"
	GroupByWorkload               AlertGroupBy  = "Workload"
	HealthRuleModeAlert           HealthRuleMode = "alert"
	HealthRuleModeRecordingRule   HealthRuleMode = "recording-rule"
)

type HealthRule struct {
	// Health rule template name.
	// Possible values are: `PacketDropsByKernel`, `PacketDropsByDevice`, `IPsecErrors`, `NetpolDenied`,
	// `LatencyHighTrend`, `DNSErrors`, `ExternalEgressHighTrend`, `ExternalIngressHighTrend`, `CrossAZ`.
	// More information: https://github.com/netobserv/network-observability-operator/blob/main/docs/Alerts.md
	// +kubebuilder:validation:Enum:="PacketDropsByKernel";"PacketDropsByDevice";"IPsecErrors";"NetpolDenied";"LatencyHighTrend";"DNSErrors";"ExternalEgressHighTrend";"ExternalIngressHighTrend";"CrossAZ"
	// +required
	Template AlertTemplate `json:"template,omitempty"`

	// Mode defines whether this health rule generates an alert or a recording rule.
	// Possible values are `alert` (default) or `recording-rule`.
	// - `alert`: Generate Prometheus alerts that fire when thresholds are exceeded.
	// - `recording-rule`: Generate Prometheus recording rules that pre-compute health metrics for passive consumption.
	// Recording rules avoid alert fatigue and are useful for dashboard-based health monitoring.
	// +kubebuilder:validation:Enum:="alert";"recording-rule"
	// +kubebuilder:default:="alert"
	// +optional
	Mode HealthRuleMode `json:"mode,omitempty"`

	// A list of variants for this template
	// +required
	Variants []HealthRuleVariant `json:"variants,omitempty"`
}

type HealthRuleVariant struct {
	// The low volume threshold allows to ignore metrics with a too low volume of traffic, in order to improve signal-to-noise.
	// It is provided as an absolute rate (bytes per second or packets per second, depending on the context).
	// When provided, it must be parsable as a float.
	LowVolumeThreshold string `json:"lowVolumeThreshold,omitempty"`

	// Thresholds per severity.
	// Only used when mode is 'alert'. They are expressed as a percentage of errors above which the alert is triggered. They must be parsable as floats.
	// For recording-rule mode, this field is ignored.
	// +optional
	Thresholds AlertThresholds `json:"thresholds,omitempty"`

	// Optional grouping criteria, possible values are: `Node`, `Namespace`, `Workload`.
	// +kubebuilder:validation:Enum:="";"Node";"Namespace";"Workload"
	// +optional
	GroupBy AlertGroupBy `json:"groupBy,omitempty"`

	// For trending health rules, the time offset for baseline comparison. For example, "1d" means comparing against yesterday. Defaults to 1d.
	TrendOffset *metav1.Duration `json:"trendOffset,omitempty"`

	// For trending health rules, the duration interval for baseline comparison. For example, "2h" means comparing against a 2-hours average. Defaults to 2h.
	TrendDuration *metav1.Duration `json:"trendDuration,omitempty"`
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

func (s *FlowCollectorSpec) GetHealthRules() []HealthRule {
	var ret []HealthRule
	var templates []AlertTemplate // for reproducible ordering

	tplMap := make(map[AlertTemplate]HealthRule)
	// Load defaults, respecting DisableAlerts
	for _, rule := range DefaultHealthRules {
		if !slices.Contains(s.Processor.Metrics.DisableAlerts, rule.Template) {
			tplMap[rule.Template] = rule
			templates = append(templates, rule.Template)
		}
	}
	// Override with user-defined rules - these take precedence over DisableAlerts
	if s.Processor.Metrics.HealthRules != nil {
		for _, rule := range *s.Processor.Metrics.HealthRules {
			// A rule explicitly defined in FC overrides the default rule and DisableAlerts
			tplMap[rule.Template] = rule
			if !slices.Contains(templates, rule.Template) {
				templates = append(templates, rule.Template)
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

func (h *HealthRule) IsAllowed(spec *FlowCollectorSpec) (bool, string) {
	switch h.Template {
	case AlertPacketDropsByKernel:
		if !spec.Agent.EBPF.IsPktDropEnabled() {
			return false, fmt.Sprintf("Health rule %s requires the %s agent feature to be enabled", h.Template, PacketDrop)
		}
	case AlertIPsecErrors:
		if !spec.Agent.EBPF.IsIPSecEnabled() {
			return false, fmt.Sprintf("Health rule %s requires the %s agent feature to be enabled", h.Template, IPSec)
		}
	case AlertDNSErrors:
		if !spec.Agent.EBPF.IsDNSTrackingEnabled() {
			return false, fmt.Sprintf("Health rule %s requires the %s agent feature to be enabled", h.Template, DNSTracking)
		}
	case AlertLatencyHighTrend:
		if !spec.Agent.EBPF.IsFlowRTTEnabled() {
			return false, fmt.Sprintf("Health rule %s requires the %s agent feature to be enabled", h.Template, FlowRTT)
		}
	case AlertNetpolDenied:
		if !spec.Agent.EBPF.IsNetworkEventsEnabled() {
			return false, fmt.Sprintf("Health rule %s requires the %s agent feature to be enabled", h.Template, NetworkEvents)
		}
	case AlertNoFlows, AlertLokiError, AlertPacketDropsByDevice, AlertExternalEgressHighTrend, AlertExternalIngressHighTrend, AlertCrossAZ:
		return true, ""
	}
	return true, ""
}

func (v *HealthRuleVariant) GetTrendParams() (string, string) {
	offset := metav1.Duration{Duration: 24 * time.Hour}
	if v.TrendOffset != nil {
		offset = *v.TrendOffset
	}
	duration := metav1.Duration{Duration: 2 * time.Hour}
	if v.TrendDuration != nil {
		duration = *v.TrendDuration
	}
	return durationToStringTrimmed(&offset), durationToStringTrimmed(&duration)
}

var regTrim = regexp.MustCompile("([a-zA-Z])(0[a-zA-Z])+")

func durationToStringTrimmed(d *metav1.Duration) string {
	s := d.Duration.String()
	return regTrim.ReplaceAllString(s, "$1")
}
