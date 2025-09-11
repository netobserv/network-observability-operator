package v1beta2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/dsl"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netobserv/network-observability-operator/internal/pkg/cluster"
)

var (
	log                    = logf.Log.WithName("flowcollector-resource")
	CurrentClusterInfo     *cluster.Info
	needPrivileged         = []AgentFeature{UDNMapping, NetworkEvents}
	neededOpenShiftVersion = map[AgentFeature]string{
		PacketDrop:    "4.14.0",
		UDNMapping:    "4.18.0",
		NetworkEvents: "4.19.0",
		EbpfManager:   "4.19.0",
	}
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateCreate(ctx context.Context, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate create", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, kerr.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.Validate(ctx, fc)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	log.Info("validate update", "name", r.Name)
	fc, ok := newObj.(*FlowCollector)
	if !ok {
		return nil, kerr.NewBadRequest(fmt.Sprintf("expected a FlowCollector but got a %T", newObj))
	}
	return r.Validate(ctx, fc)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FlowCollector) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	log.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *FlowCollector) Validate(_ context.Context, fc *FlowCollector) (admission.Warnings, error) {
	v := validator{fc: &fc.Spec}
	v.validateAgent()
	v.validateFLP()
	v.warnLogLevels()
	return v.warnings, errors.Join(v.errors...)
}

type validator struct {
	fc       *FlowCollectorSpec
	warnings admission.Warnings
	errors   []error
}

func (v *validator) warnLogLevels() {
	if v.fc.Agent.EBPF.LogLevel == "debug" || v.fc.Agent.EBPF.LogLevel == "trace" {
		v.warnings = append(v.warnings, fmt.Sprintf("The log level for the eBPF agent is %s, which impacts performance and resource footprint.", v.fc.Agent.EBPF.LogLevel))
	}
	if v.fc.Processor.LogLevel == "debug" || v.fc.Processor.LogLevel == "trace" {
		v.warnings = append(v.warnings, fmt.Sprintf("The log level for the processor (flowlogs-pipeline) is %s, which impacts performance and resource footprint.", v.fc.Processor.LogLevel))
	}
}

func (v *validator) validateAgent() {
	for feat, minVersion := range neededOpenShiftVersion {
		if slices.Contains(v.fc.Agent.EBPF.Features, feat) {
			if CurrentClusterInfo != nil && CurrentClusterInfo.IsOpenShift() {
				// Make sure required version of ocp is installed
				ok, err := CurrentClusterInfo.IsOpenShiftVersionAtLeast(minVersion)
				if err != nil {
					v.warnings = append(v.warnings, fmt.Sprintf("Could not detect OpenShift cluster version: %s", err.Error()))
				} else if !ok {
					v.warnings = append(v.warnings, fmt.Sprintf("The %s feature requires OpenShift %s or above (version detected: %s)", feat, minVersion, CurrentClusterInfo.GetOpenShiftVersion()))
				}
			} else {
				v.warnings = append(v.warnings, fmt.Sprintf("Unknown environment, cannot detect if the feature %s is supported", feat))
			}
		}
	}
	if !v.fc.Agent.EBPF.Privileged {
		for _, feat := range needPrivileged {
			if slices.Contains(v.fc.Agent.EBPF.Features, feat) {
				v.warnings = append(v.warnings, fmt.Sprintf("The %s feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged", feat))
			}
		}
	}

	if slices.Contains(v.fc.Agent.EBPF.Features, PacketDrop) &&
		!v.fc.Agent.EBPF.Privileged &&
		!slices.Contains(v.fc.Agent.EBPF.Features, EbpfManager) {
		v.warnings = append(v.warnings, "The PacketDrop feature requires eBPF Agent to run in privileged mode, which is currently disabled in spec.agent.ebpf.privileged, or to use with eBPF Manager")
	}
	if v.fc.Agent.EBPF.FlowFilter != nil && v.fc.Agent.EBPF.FlowFilter.Enable != nil && *v.fc.Agent.EBPF.FlowFilter.Enable {
		m := make(map[string]bool)
		for i := range v.fc.Agent.EBPF.FlowFilter.Rules {
			rule := v.fc.Agent.EBPF.FlowFilter.Rules[i]
			key := rule.CIDR + "-" + rule.PeerCIDR
			if found := m[key]; found {
				v.errors = append(v.errors, fmt.Errorf("flow filter rule CIDR and PeerCIDR %s already exists", key))
				break
			}
			m[key] = true
			v.validateAgentFilter(&rule)
		}
		v.validateAgentFilter(&v.fc.Agent.EBPF.FlowFilter.EBPFFlowFilterRule)
	}
}

func (v *validator) validateAgentFilter(f *EBPFFlowFilterRule) {
	if f.CIDR != "" {
		if _, _, err := net.ParseCIDR(f.CIDR); err != nil {
			v.errors = append(v.errors, err)
		}
	}
	hasPorts := f.Ports.IntVal > 0 || f.Ports.StrVal != ""
	if hasPorts {
		if err := validateFilterPortConfig(f.Ports); err != nil {
			v.errors = append(v.errors, err)
		}
	}
	hasSrcPorts := f.SourcePorts.IntVal > 0 || f.SourcePorts.StrVal != ""
	if hasSrcPorts {
		if err := validateFilterPortConfig(f.SourcePorts); err != nil {
			v.errors = append(v.errors, err)
		}
	}
	hasDstPorts := f.DestPorts.IntVal > 0 || f.DestPorts.StrVal != ""
	if hasDstPorts {
		if err := validateFilterPortConfig(f.DestPorts); err != nil {
			v.errors = append(v.errors, err)
		}
	}
	if hasPorts && hasSrcPorts {
		v.errors = append(v.errors, errors.New("cannot configure agent filter with ports and sourcePorts, they are mutually exclusive"))
	}
	if hasPorts && hasDstPorts {
		v.errors = append(v.errors, errors.New("cannot configure agent filter with ports and destPorts, they are mutually exclusive"))
	}
}

func validateFilterPortConfig(value intstr.IntOrString) error {
	if value.Type == intstr.Int {
		return nil
	}
	sVal := value.String()
	if strings.Contains(sVal, "-") {
		ps := strings.SplitN(sVal, "-", 2)
		if len(ps) != 2 {
			return fmt.Errorf("invalid ports range: expected two integers separated by '-' but found %s", sVal)
		}
		start, err := validatePortString(ps[0])
		if err != nil {
			return fmt.Errorf("start port in range: %w", err)
		}
		end, err := validatePortString(ps[1])
		if err != nil {
			return fmt.Errorf("end port in range: %w", err)
		}
		if start >= end {
			return fmt.Errorf("invalid port range: start is greater or equal to end")
		}
		return nil
	} else if strings.Contains(sVal, ",") {
		ps := strings.Split(sVal, ",")
		if len(ps) != 2 {
			return fmt.Errorf("invalid ports couple: expected two integers separated by ',' but found %s", sVal)
		}
		_, err := validatePortString(ps[0])
		if err != nil {
			return fmt.Errorf("first port: %w", err)
		}
		_, err = validatePortString(ps[1])
		if err != nil {
			return fmt.Errorf("second port: %w", err)
		}
		return nil
	}
	// Should be a single port then
	_, err := validatePortString(sVal)
	if err != nil {
		return err
	}
	return nil
}

func validatePortString(s string) (uint16, error) {
	p, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number %w", err)
	}
	if p == 0 {
		return 0, fmt.Errorf("invalid port 0")
	}
	return uint16(p), nil
}

func (v *validator) validateFLP() {
	v.validateFLPLogTypes()
	v.validateFLPFilters()
	v.validateFLPAlerts()
	v.validateFLPMetricsForAlerts()
}

func (v *validator) validateFLPLogTypes() {
	if v.fc.Processor.LogTypes != nil && *v.fc.Processor.LogTypes == LogTypeAll {
		v.warnings = append(v.warnings, "Enabling all log types (in spec.processor.logTypes) has a high impact on resources footprint")
	}
	if v.fc.Processor.LogTypes != nil && *v.fc.Processor.LogTypes != LogTypeFlows && v.fc.Loki.Enable != nil && !*v.fc.Loki.Enable {
		v.errors = append(v.errors, errors.New("enabling conversation tracking without Loki is not allowed, as it generates extra processing for no benefit"))
	}
}

func (v *validator) validateFLPFilters() {
	for i, filter := range v.fc.Processor.Filters {
		if _, err := dsl.Parse(filter.Query); err != nil {
			v.errors = append(v.errors, fmt.Errorf("cannot parse spec.processor.filters[%d].query: %w", i, err))
		}
	}
}

func (v *validator) validateFLPAlerts() {
	if v.fc.Processor.Metrics.Alerts != nil {
		for i, alert := range *v.fc.Processor.Metrics.Alerts {
			if _, msg := alert.IsAllowed(v.fc); len(msg) > 0 {
				v.warnings = append(v.warnings, msg)
			}
			for j, variant := range alert.Variants {
				// Check allowed groups
				if !v.isFLPAlertGroupBySupported(alert.Template, &variant) {
					v.errors = append(
						v.errors,
						fmt.Errorf(
							`%s alert template does not support grouping by %s, in spec.processor.metrics.alerts[%d].variants[%d]`,
							alert.Template, variant.GroupBy, i, j,
						),
					)
				}
				lastThreshold := float64(-1)
				thresholds := []struct {
					s string
					t string
				}{
					{s: "critical", t: variant.Thresholds.Critical},
					{s: "warning", t: variant.Thresholds.Warning},
					{s: "info", t: variant.Thresholds.Info},
				}
				for _, st := range thresholds {
					if st.t != "" {
						val, err := strconv.ParseFloat(st.t, 64)
						if err != nil {
							v.errors = append(
								v.errors,
								fmt.Errorf(`cannot parse %s threshold as float in spec.processor.metrics.alerts[%d].variants[%d]: "%s"`, st.s, i, j, st.t),
							)
						} else if val < 0 {
							v.errors = append(
								v.errors,
								fmt.Errorf(`%s threshold must be positive in spec.processor.metrics.alerts[%d].variants[%d]: "%s"`, st.s, i, j, st.t),
							)
						} else if lastThreshold > 0 && val > lastThreshold {
							v.errors = append(
								v.errors,
								fmt.Errorf(`%s threshold must be lower than %.0f, which is defined for a higher severity, in spec.processor.metrics.alerts[%d].variants[%d]: "%s"`, st.s, lastThreshold, i, j, st.t),
							)
						}
						lastThreshold = val
					}
				}
				if variant.LowVolumeThreshold != "" {
					_, err := strconv.ParseFloat(variant.LowVolumeThreshold, 64)
					if err != nil {
						v.errors = append(
							v.errors,
							fmt.Errorf(`cannot parse lowVolumeThreshold as float in spec.processor.metrics.alerts[%d].variants[%d]: "%s"`, i, j, variant.LowVolumeThreshold),
						)
					}
				}
			}
		}
	}
}

func (v *validator) isFLPAlertGroupBySupported(template AlertTemplate, variant *AlertVariant) bool {
	switch template {
	case AlertPacketDropsByDevice:
		return variant.GroupBy != GroupByWorkload
	case AlertIPsecErrors:
		return variant.GroupBy != GroupByWorkload && variant.GroupBy != GroupByNamespace
	case AlertPacketDropsByKernel, AlertDNSErrors, AlertExternalEgressHighTrend, AlertExternalIngressHighTrend, AlertLatencyHighTrend, AlertNetpolDenied:
		return true
	}
	return true
}

func (v *validator) validateFLPMetricsForAlerts() {
	metrics := v.fc.GetIncludeList()
	alerts := v.fc.GetFLPAlerts()
	for _, g := range alerts {
		for _, a := range g.Variants {
			reqMetrics1, reqMetrics2 := GetElligibleMetricsForAlert(g.Template, &a)
			// At least one metric from reqMetrics1 should be present, same for reqMetrics2
			if len(reqMetrics1) > 0 {
				if GetFirstRequiredMetrics(reqMetrics1, metrics) == "" {
					v.warnings = append(
						v.warnings,
						fmt.Sprintf("Alert %s/%s requires enabling at least one metric from this list: %s", g.Template, a.GroupBy, strings.Join(reqMetrics1, ", ")),
					)
				}
			}
			if len(reqMetrics2) > 0 {
				if GetFirstRequiredMetrics(reqMetrics2, metrics) == "" {
					v.warnings = append(
						v.warnings,
						fmt.Sprintf("Alert %s/%s requires enabling at least one metric from this list: %s", g.Template, a.GroupBy, strings.Join(reqMetrics2, ", ")),
					)
				}
			}
		}
	}
}

func GetFirstRequiredMetrics(anyRequired, actual []string) string {
	for _, m := range anyRequired {
		if slices.Contains(actual, m) {
			return m
		}
	}
	return ""
}

func GetElligibleMetricsForAlert(template AlertTemplate, alertDef *AlertVariant) ([]string, []string) {
	var metricPatterns, totalMetricPatterns []string
	switch template {
	case AlertPacketDropsByKernel:
		metricPatterns = []string{"%s_drop_packets_total"}
		totalMetricPatterns = []string{"%s_ingress_packets_total", "%s_egress_packets_total"}
	case AlertIPsecErrors:
		return []string{"node_ipsec_flows_total"}, []string{"node_to_node_ingress_flows_total"}
	case AlertDNSErrors:
		metricPatterns = []string{`%s_dns_latency_seconds`}
		totalMetricPatterns = []string{"%s_dns_latency_seconds"}
	case AlertExternalEgressHighTrend:
	case AlertExternalIngressHighTrend:
	case AlertLatencyHighTrend:
		metricPatterns = []string{`%s_rtt_seconds`}
		totalMetricPatterns = []string{`%s_rtt_seconds`}
	case AlertNetpolDenied:
		metricPatterns = []string{`%s_network_policy_events_total`}
		totalMetricPatterns = []string{"%s_flows_total"}
	case AlertNoFlows, AlertLokiError, AlertPacketDropsByDevice:
		// nothing
		return nil, nil
	}
	var gr []string
	switch alertDef.GroupBy {
	case GroupByNode:
		gr = []string{"node"}
	case GroupByNamespace:
		gr = []string{"namespace", "workload"}
	case GroupByWorkload:
		gr = []string{"workload"}
	default: // global => any of the metric can work
		gr = []string{"namespace", "workload", "node"}
	}
	var metrics, totalMetrics []string
	for _, p := range metricPatterns {
		for _, g := range gr {
			metrics = append(metrics, fmt.Sprintf(p, g))
		}
	}
	for _, p := range totalMetricPatterns {
		for _, g := range gr {
			totalMetrics = append(totalMetrics, fmt.Sprintf(p, g))
		}
	}
	return metrics, totalMetrics
}
