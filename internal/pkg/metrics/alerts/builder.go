package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type srcOrDst string

const (
	asSource srcOrDst = "Src"
	asDest   srcOrDst = "Dst"
)

type ruleBuilder struct {
	template          flowslatest.HealthRuleTemplate
	healthRule        *flowslatest.HealthRuleVariant
	mode              flowslatest.HealthRuleMode
	enabledMetrics    []string
	side              srcOrDst
	severity          string
	threshold         string
	upperThreshold    string
	upperValueRange   string
	trafficLinkFilter string
	duration          monitoringv1.Duration
}

func BuildRules(ctx context.Context, fc *flowslatest.FlowCollectorSpec) []monitoringv1.Rule {
	log := log.FromContext(ctx)
	rules := []monitoringv1.Rule{}

	if fc.HasExperimentalAlertsHealth() {
		healthRules := fc.GetFLPHealthRules()
		metrics := fc.GetIncludeList()
		for _, healthRule := range healthRules {
			if ok, _ := healthRule.IsAllowed(fc); !ok {
				continue
			}
			for _, variant := range healthRule.Variants {
				if r, err := convertToRules(healthRule.Template, healthRule.Mode, &variant, metrics); err != nil {
					log.Error(err, "unable to configure a health rule")
				} else if len(r) > 0 {
					rules = append(rules, r...)
				}
			}
		}
	}

	if !slices.Contains(fc.Processor.Metrics.DisableHealthRules, flowslatest.HealthRuleNoFlows) {
		r := alertNoFlows()
		rules = append(rules, *r)
	}
	if !slices.Contains(fc.Processor.Metrics.DisableHealthRules, flowslatest.HealthRuleLokiError) {
		r := alertLokiError()
		rules = append(rules, *r)
	}

	return rules
}

func convertToRules(template flowslatest.HealthRuleTemplate, mode flowslatest.HealthRuleMode, healthRule *flowslatest.HealthRuleVariant, enabledMetrics []string) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	var upperThreshold string
	sides := []srcOrDst{asSource, asDest}
	if healthRule.GroupBy == "" {
		// No side for global group
		sides = []srcOrDst{""}
	}

	// Default mode to alert if not specified
	if mode == "" {
		mode = flowslatest.ModeAlert
	}

	// For recording rules, create a single rule without severity levels
	if mode == flowslatest.ModeRecording {
		for _, side := range sides {
			rb := ruleBuilder{
				template:       template,
				healthRule:     healthRule,
				mode:           mode,
				enabledMetrics: enabledMetrics,
				side:           side,
				severity:       "",
				threshold:      "", // Not used for recording rules
				upperThreshold: "",
				duration:       monitoringv1.Duration("5m"),
			}

			if r, err := rb.convertToRule(); err != nil {
				return nil, err
			} else if r != nil {
				rules = append(rules, *r)
			}
		}
		return rules, nil
	}

	// For alert mode, create up to 3 rules, one per severity, with non-overlapping thresholds
	thresholds := []struct {
		s string
		t string
	}{
		{s: "critical", t: healthRule.Thresholds.Critical},
		{s: "warning", t: healthRule.Thresholds.Warning},
		{s: "info", t: healthRule.Thresholds.Info},
	}
	for _, st := range thresholds {
		if st.t != "" {
			for _, side := range sides {
				rb := ruleBuilder{
					template:       template,
					healthRule:     healthRule,
					mode:           mode,
					enabledMetrics: enabledMetrics,
					side:           side,
					severity:       st.s,
					threshold:      st.t,
					upperThreshold: upperThreshold,
					duration:       monitoringv1.Duration("5m"),
				}
				if r, err := rb.convertToRule(); err != nil {
					return nil, err
				} else if r != nil {
					rules = append(rules, *r)
				}
			}
			upperThreshold = st.t
		}
	}
	return rules, nil
}

func (rb *ruleBuilder) convertToRule() (*monitoringv1.Rule, error) {
	switch rb.template {
	case flowslatest.HealthRulePacketDropsByDevice:
		return rb.deviceDrops()
	case flowslatest.HealthRulePacketDropsByKernel:
		return rb.kernelDrops()
	case flowslatest.HealthRuleIPsecErrors:
		return rb.ipsecErrors()
	case flowslatest.HealthRuleDNSErrors:
		return rb.dnsErrors()
	case flowslatest.HealthRuleNetpolDenied:
		return rb.netpolDenied()
	case flowslatest.HealthRuleLatencyHighTrend:
		return rb.latencyTrend()
	case flowslatest.HealthRuleCrossAZ, flowslatest.HealthRuleExternalEgressHighTrend, flowslatest.HealthRuleExternalIngressHighTrend: // TODO
	case flowslatest.HealthRuleLokiError, flowslatest.HealthRuleNoFlows:
		// error
	}
	return nil, fmt.Errorf("unknown health rule template: %s", rb.template)
}

func (rb *ruleBuilder) additionalDescription() string {
	return fmt.Sprintf("You can turn off this health rule by adding '%s' to spec.processor.metrics.disableHealthRules in FlowCollector, or reconfigure it via spec.processor.metrics.healthRules.", rb.template)
}

// Known acronyms that should stay together in snake_case
var acronyms = map[string]string{
	"DNS":    "dns",
	"IPsec":  "ipsec",
	"IP":     "ip",
	"AZ":     "az",
	"TCP":    "tcp",
	"UDP":    "udp",
	"HTTP":   "http",
	"HTTPS":  "https",
}

// toSnakeCase converts a camelCase or PascalCase string to snake_case
// Handles known acronyms specially (e.g., "DNSErrors" -> "dns_errors", "CrossAZ" -> "cross_az")
func toSnakeCase(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		// Check if current position matches any acronym
		matched := false
		for acronym, replacement := range acronyms {
			if strings.HasPrefix(s[i:], acronym) {
				// Add underscore before acronym if not at start and previous char was lowercase
				if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
					result.WriteRune('_')
				}
				result.WriteString(replacement)
				i += len(acronym)
				matched = true
				break
			}
		}

		if !matched {
			// Regular camelCase handling
			r := rune(s[i])
			if i > 0 && r >= 'A' && r <= 'Z' {
				result.WriteRune('_')
			}
			result.WriteRune(r)
			i++
		}
	}

	return strings.ToLower(result.String())
}

// buildRecordingRuleName builds recording rule name following the convention:
// netobserv:health:<template>:<groupby>:<side>:rate5m
func (rb *ruleBuilder) buildRecordingRuleName() string {
	parts := []string{"netobserv", "health"}

	// Add template in snake_case
	parts = append(parts, toSnakeCase(string(rb.template)))

	// Add groupBy if present
	if rb.healthRule.GroupBy != "" {
		parts = append(parts, strings.ToLower(string(rb.healthRule.GroupBy)))
	}

	// Add side if groupBy is present (side is only relevant with groupBy)
	if rb.healthRule.GroupBy != "" && rb.side != "" {
		parts = append(parts, strings.ToLower(string(rb.side)))
	}

	// Add rate interval (default to rate5m for 2m window)
	parts = append(parts, "rate5m")

	return strings.Join(parts, ":")
}

func (rb *ruleBuilder) createRule(promQL, summary, description string) (*monitoringv1.Rule, error) {
	bAnnot, err := rb.buildHealthAnnotation(nil)
	if err != nil {
		return nil, err
	}

	var gr string
	if rb.healthRule.GroupBy != "" {
		gr = "Per" + string(rb.side) + string(rb.healthRule.GroupBy)
	}

	// Generate recording rule
	if rb.mode == flowslatest.ModeRecording {
		recordName := rb.buildRecordingRuleName()
		return &monitoringv1.Rule{
			Record: recordName,
			// Note: Recording rules cannot have annotations in Prometheus
			Expr:   intstr.FromString(promQL),
			Labels: buildLabels("", true),
		}, nil
	}

	// Generate alert rule
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", rb.template, gr, strings.ToUpper(rb.severity[:1])+rb.severity[1:]),
		Annotations: map[string]string{
			"description":                 description,
			"summary":                     summary,
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promQL),
		For:    &rb.duration,
		Labels: buildLabels(rb.severity, true),
	}, nil
}

func (rb *ruleBuilder) getAlertLegend() string {
	var sideText string
	switch rb.side {
	case asSource:
		sideText = "source "
	case asDest:
		sideText = "dest. "
	}
	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		return " [" + sideText + "node={{ $labels.node }}]"
	case flowslatest.GroupByNamespace:
		return " [" + sideText + "namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return " [" + sideText + "workload={{ $labels.workload }} ({{ $labels.kind }})]"
	}
	return ""
}

func (rb *ruleBuilder) buildHealthAnnotation(override map[string]any) ([]byte, error) {
	// The health annotation contains json-encoded information used in console plugin display
	annotation := map[string]any{
		"threshold": rb.threshold,
		"unit":      "%",
	}
	if rb.upperValueRange != "" {
		annotation["upperBound"] = rb.upperValueRange
	}
	if rb.trafficLinkFilter != "" {
		annotation["trafficLinkFilter"] = rb.trafficLinkFilter
	}
	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		annotation["nodeLabels"] = []string{"node"}
	case flowslatest.GroupByNamespace, flowslatest.GroupByWorkload:
		annotation["namespaceLabels"] = []string{"namespace"}
	}
	for k, v := range override {
		annotation[k] = v
	}
	bAnnot, err := json.Marshal(annotation)
	if err != nil {
		return nil, fmt.Errorf("cannot encode health rule annotation [template=%s]: %w", rb.template, err)
	}
	return bAnnot, nil
}

func buildLabels(severity string, forHealth bool) map[string]string {
	m := map[string]string{
		"app": "netobserv", // means that the rule is created by netobserv
	}
	if severity != "" {
		m["severity"] = strings.ToLower(severity)
	}
	if forHealth {
		m["netobserv"] = "true" // means that the rule should be fetched by netobserv console plugin for health
	}
	return m
}

func (rb *ruleBuilder) buildLabelFilter(additionalFilter string) string {
	var filters []string

	// Build label matchers to filter out metrics where K8s labels don't exist or are empty
	// This prevents alerts from firing with empty namespace/workload/node labels
	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		filters = append(filters, string(rb.side)+`K8S_HostName!=""`)
	case flowslatest.GroupByNamespace:
		filters = append(filters, string(rb.side)+`K8S_Namespace!=""`)
	case flowslatest.GroupByWorkload:
		filters = append(filters, string(rb.side)+`K8S_Namespace!=""`)
		filters = append(filters, string(rb.side)+`K8S_OwnerName!=""`)
		filters = append(filters, string(rb.side)+`K8S_OwnerType!=""`)
	}

	// Add additional business logic filters
	if additionalFilter != "" {
		filters = append(filters, additionalFilter)
	}

	if len(filters) == 0 {
		return ""
	}
	return "{" + strings.Join(filters, ",") + "}"
}

func (rb *ruleBuilder) getMetricsForAlert() (string, string) {
	var reqMetric1, reqMetric2 string
	reqMetrics1, reqMetrics2 := flowslatest.GetElligibleMetricsForHealthRule(rb.template, rb.healthRule)
	if len(reqMetrics1) > 0 {
		reqMetric1 = flowslatest.GetFirstRequiredMetrics(reqMetrics1, rb.enabledMetrics)
	}
	if len(reqMetrics2) > 0 {
		reqMetric2 = flowslatest.GetFirstRequiredMetrics(reqMetrics2, rb.enabledMetrics)
	}
	return reqMetric1, reqMetric2
}
