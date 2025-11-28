package alerts

import (
	"fmt"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// convertToRecordingRules converts alert configuration to recording rules
func convertToRecordingRules(template flowslatest.AlertTemplate, alert *flowslatest.HealthRuleVariant, enabledMetrics []string) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	sides := []srcOrDst{asSource, asDest}
	if alert.GroupBy == "" {
		// No side for global group
		sides = []srcOrDst{""}
	}

	// For recording rules, we create one rule per side (not per severity)
	for _, side := range sides {
		rb := ruleBuilder{
			template:       template,
			alert:          alert,
			enabledMetrics: enabledMetrics,
			side:           side,
			duration:       monitoringv1.Duration("5m"),
		}
		if r, err := rb.convertToRecordingRule(); err != nil {
			return nil, err
		} else if r != nil {
			rules = append(rules, *r)
		}
	}
	return rules, nil
}

func (rb *ruleBuilder) convertToRecordingRule() (*monitoringv1.Rule, error) {
	switch rb.template {
	case flowslatest.AlertPacketDropsByKernel:
		return rb.kernelDropsRecording()
	case flowslatest.AlertIPsecErrors:
		return rb.ipsecErrorsRecording()
	case flowslatest.AlertDNSErrors:
		return rb.dnsErrorsRecording()
	case flowslatest.AlertNetpolDenied:
		return rb.netpolDeniedRecording()
	case flowslatest.AlertLatencyHighTrend:
		return rb.latencyTrendRecording()
	case flowslatest.AlertPacketDropsByDevice:
		return rb.deviceDropsRecording()
	case flowslatest.AlertExternalEgressHighTrend, flowslatest.AlertExternalIngressHighTrend, flowslatest.AlertCrossAZ:
		// TODO: implement these
		return nil, nil
	case flowslatest.AlertLokiError, flowslatest.AlertNoFlows:
		// These are handled separately in BuildRecordingRules
		return nil, nil
	}
	return nil, fmt.Errorf("unknown recording rule template: %s", rb.template)
}

func (rb *ruleBuilder) buildRecordingRuleName() string {
	// Format: netobserv:health:<template>:<groupby>:<side>:rate5m
	// Example: netobserv:health:packet_drops_by_kernel:namespace:src:rate5m

	templateLower := strings.ToLower(string(rb.template))
	// Convert CamelCase to snake_case
	templateSnake := camelToSnake(templateLower)

	var parts []string
	parts = append(parts, "netobserv", "health", templateSnake)

	if rb.alert.GroupBy != "" {
		parts = append(parts, strings.ToLower(string(rb.alert.GroupBy)))
	}

	if rb.side != "" {
		parts = append(parts, strings.ToLower(string(rb.side)))
	}

	parts = append(parts, "rate5m")

	return strings.Join(parts, ":")
}

func (rb *ruleBuilder) buildRecordingRuleLabels() map[string]string {
	labels := map[string]string{
		"netobserv":       "health",
		"health_template": string(rb.template),
	}

	if rb.alert.GroupBy != "" {
		labels["health_groupby"] = string(rb.alert.GroupBy)
	}

	if rb.side != "" {
		labels["health_side"] = string(rb.side)
	}

	return labels
}

func (rb *ruleBuilder) kernelDropsRecording() (*monitoringv1.Rule, error) {
	metric, totalMetric := rb.getMetricsForAlert()
	metricsRate := promQLRateFromMetric(metric, "", "", "5m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "5m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")

	// Recording rule: compute the percentage without threshold comparison
	promql := fmt.Sprintf("100 * (%s) / (%s)", metricsSumBy, totalSumBy)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func (rb *ruleBuilder) deviceDropsRecording() (*monitoringv1.Rule, error) {
	// No "side" consideration on netdev metrics, so keep only 1 call from the two of them
	if rb.side == asDest {
		return nil, nil
	}

	var byLabels string
	switch rb.alert.GroupBy {
	case flowslatest.GroupByNode:
		byLabels = " by (instance)"
	case flowslatest.GroupByNamespace:
		return nil, fmt.Errorf("PacketDropsByDevice recording rule does not support grouping per namespace")
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("PacketDropsByDevice recording rule does not support grouping per workload")
	}

	promql := fmt.Sprintf(
		"100 * (sum(rate(node_network_receive_drop_total[5m]))%s + sum(rate(node_network_transmit_drop_total[5m]))%s) / (sum(rate(node_network_receive_packets_total[5m]))%s + sum(rate(node_network_transmit_packets_total[5m]))%s)",
		byLabels, byLabels, byLabels, byLabels,
	)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func (rb *ruleBuilder) ipsecErrorsRecording() (*monitoringv1.Rule, error) {
	metric, totalMetric := rb.getMetricsForAlert()
	metricsRate := promQLRateFromMetric(metric, "", "", "5m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "5m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := fmt.Sprintf("100 * (%s) / (%s)", metricsSumBy, totalSumBy)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func (rb *ruleBuilder) dnsErrorsRecording() (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if rb.side == asSource {
		return nil, nil
	}

	metric, totalMetric := rb.getMetricsForAlert()
	metricsRate := promQLRateFromMetric(metric, "_count", `{DnsFlagsResponseCode!="NoError"}`, "5m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", "", "5m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := fmt.Sprintf("100 * (%s) / (%s)", metricsSumBy, totalSumBy)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func (rb *ruleBuilder) netpolDeniedRecording() (*monitoringv1.Rule, error) {
	metric, totalMetric := rb.getMetricsForAlert()
	metricsRate := promQLRateFromMetric(metric, "", `{action="drop"}`, "5m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "5m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := fmt.Sprintf("100 * (%s) / (%s)", metricsSumBy, totalSumBy)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func (rb *ruleBuilder) latencyTrendRecording() (*monitoringv1.Rule, error) {
	offset, duration := rb.alert.GetTrendParams()

	metric, baseline := rb.getMetricsForAlert()
	metricsRate := promQLRateFromMetric(metric, "_bucket", "", "5m", "")
	baselineRate := promQLRateFromMetric(baseline, "_bucket", "", duration, " offset "+offset)
	metricQuantile := histogramQuantile(metricsRate, rb.alert.GroupBy, rb.side, "0.9")
	baselineQuantile := histogramQuantile(baselineRate, rb.alert.GroupBy, rb.side, "0.9")

	// Recording rule: compute the percentage increase without threshold comparison
	promql := fmt.Sprintf("100 * ((%s) - (%s)) / (%s)", metricQuantile, baselineQuantile, baselineQuantile)

	return &monitoringv1.Rule{
		Record: rb.buildRecordingRuleName(),
		Expr:   intstr.FromString(promql),
		Labels: rb.buildRecordingRuleLabels(),
	}, nil
}

func RecordingNoFlows() *monitoringv1.Rule {
	return &monitoringv1.Rule{
		Record: "netobserv:health:no_flows:rate1m",
		Expr:   intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m]))"),
		Labels: map[string]string{
			"netobserv":       "health",
			"health_template": "NetObservNoFlows",
		},
	}
}

func RecordingLokiError() *monitoringv1.Rule {
	return &monitoringv1.Rule{
		Record: "netobserv:health:loki_errors:rate1m",
		Expr:   intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m]))"),
		Labels: map[string]string{
			"netobserv":       "health",
			"health_template": "NetObservLokiError",
		},
	}
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
