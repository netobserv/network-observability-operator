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

func BuildRules(ctx context.Context, fc *flowslatest.FlowCollectorSpec) []monitoringv1.Rule {
	log := log.FromContext(ctx)
	rules := []monitoringv1.Rule{}

	alerts := fc.GetFLPAlerts()
	metrics := fc.GetIncludeList()
	for _, alert := range alerts {
		if ok, _ := alert.IsAllowed(fc); !ok {
			continue
		}
		for _, variant := range alert.Variants {
			if r, err := convertToRules(alert.Template, &variant, metrics); err != nil {
				log.Error(err, "unable to configure an alert")
			} else if len(r) > 0 {
				rules = append(rules, r...)
			}
		}
	}

	if !slices.Contains(fc.Processor.Metrics.DisableAlerts, flowslatest.AlertNoFlows) {
		r := alertNoFlows()
		rules = append(rules, *r)
	}
	if !slices.Contains(fc.Processor.Metrics.DisableAlerts, flowslatest.AlertLokiError) {
		r := alertLokiError()
		rules = append(rules, *r)
	}

	return rules
}

func convertToRules(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, enabledMetrics []string) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	var upperThreshold string
	sides := []srcOrDst{asSource, asDest}
	if alert.GroupBy == "" {
		// No side for global group
		sides = []srcOrDst{""}
	}
	// Create up to 3 rules, one per severity, with non-overlapping thresholds
	thresholds := []struct {
		s string
		t string
	}{
		{s: "critical", t: alert.Thresholds.Critical},
		{s: "warning", t: alert.Thresholds.Warning},
		{s: "info", t: alert.Thresholds.Info},
	}
	for _, st := range thresholds {
		if st.t != "" {
			for _, side := range sides {
				if r, err := convertToRule(template, alert, enabledMetrics, side, st.s, st.t, upperThreshold); err != nil {
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

func convertToRule(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, enabledMetrics []string, side srcOrDst, severity, threshold, upperThreshold string) (*monitoringv1.Rule, error) {
	additionalDescription := fmt.Sprintf("You can turn off this alert by adding '%s' to spec.processor.metrics.disableAlerts in FlowCollector, or reconfigure it via spec.processor.metrics.alerts.", template)
	switch template {
	case flowslatest.AlertPacketDropsByDevice:
		return deviceDrops(alert, side, severity, threshold, upperThreshold, additionalDescription)
	case flowslatest.AlertPacketDropsByKernel:
		return kernelDrops(alert, side, severity, threshold, upperThreshold, additionalDescription, enabledMetrics)
	case flowslatest.AlertIPsecErrors:
		return ipsecErrors(alert, side, severity, threshold, upperThreshold, additionalDescription, enabledMetrics)
	case flowslatest.AlertDNSErrors:
		return dnsErrors(alert, side, severity, threshold, upperThreshold, additionalDescription, enabledMetrics)
	case flowslatest.AlertNetpolDenied:
		return netpolDenied(alert, side, severity, threshold, upperThreshold, additionalDescription, enabledMetrics)
	case flowslatest.AlertLatencyHighTrend:
		return latencyTrend(alert, side, severity, threshold, upperThreshold, additionalDescription, enabledMetrics)
	case flowslatest.AlertLokiError, flowslatest.AlertNoFlows:
		// error
	}
	return nil, fmt.Errorf("unknown alert template: %s", template)
}

func createRule(tpl flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, side srcOrDst, promQL, summary, description, severity, threshold string, d monitoringv1.Duration) (*monitoringv1.Rule, error) {
	bAnnot, err := buildHealthAnnotation(tpl, alert, threshold, nil)
	if err != nil {
		return nil, err
	}

	var gr string
	if alert.GroupBy != "" {
		gr = "Per" + string(side) + string(alert.GroupBy)
	}
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", tpl, gr, strings.ToUpper(severity[:1])+severity[1:]),
		Annotations: map[string]string{
			"description":                 description,
			"summary":                     summary,
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promQL),
		For:    &d,
		Labels: buildLabels(severity, true),
	}, nil
}

type promQLRate string

func promQLRateFromMetric(metric, suffix, filters, interval, offset string) promQLRate {
	return promQLRate(fmt.Sprintf("rate(netobserv_%s%s%s[%s]%s)", metric, suffix, filters, interval, offset))
}

func sumBy(promQL promQLRate, groupBy flowslatest.AlertGroupBy, side srcOrDst, extraLabel string) string {
	var nooLabels []string
	var labelsOut []string
	switch groupBy {
	case flowslatest.GroupByNode:
		nooLabels = []string{string(side) + "K8S_HostName"}
		labelsOut = []string{"node"}
	case flowslatest.GroupByNamespace:
		nooLabels = []string{string(side) + "K8S_Namespace"}
		labelsOut = []string{"namespace"}
	case flowslatest.GroupByWorkload:
		nooLabels = []string{string(side) + "K8S_Namespace", string(side) + "K8S_OwnerName", string(side) + "K8S_OwnerType"}
		labelsOut = []string{"namespace", "workload", "kind"}
	}
	if len(labelsOut) > 0 {
		// promQL input is like "rate(netobserv_workload_ingress_bytes_total[1m])"
		// we need to relabel src / dst labels to the same label name in order to allow adding them
		// e.g. of desired output:
		// sum(label_replace(rate(netobserv_workload_ingress_bytes_total[1m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace)
		replacedLabels := string(promQL)
		for i := range labelsOut {
			in := nooLabels[i]
			out := labelsOut[i]
			replacedLabels = fmt.Sprintf(`label_replace(%s, "%s", "$1", "%s", "(.*)")`, replacedLabels, out, in)
		}
		joinedLabels := strings.Join(labelsOut, ",")
		if extraLabel != "" {
			joinedLabels += "," + extraLabel
		}
		return fmt.Sprintf("sum(%s) by (%s)", replacedLabels, joinedLabels)
	} else if extraLabel != "" {
		return fmt.Sprintf("sum(%s) by (%s)", promQL, extraLabel)
	}
	return fmt.Sprintf("sum(%s)", promQL)
}

func histogramQuantile(promQL promQLRate, groupBy flowslatest.AlertGroupBy, side srcOrDst, quantile string) string {
	sumQL := sumBy(promQL, groupBy, side, "le")
	return fmt.Sprintf("histogram_quantile(%s, %s)", quantile, sumQL)
}

func getAlertLegend(side srcOrDst, alert *flowslatest.AlertVariant) string {
	var sideText string
	switch side {
	case asSource:
		sideText = "source "
	case asDest:
		sideText = "dest. "
	}
	switch alert.GroupBy {
	case flowslatest.GroupByNode:
		return " [" + sideText + "node={{ $labels.node }}]"
	case flowslatest.GroupByNamespace:
		return " [" + sideText + "namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return " [" + sideText + "workload={{ $labels.workload }} ({{ $labels.kind }})]"
	}
	return ""
}

func percentagePromQL(promQLMetricSum, promQLTotalSum string, threshold, upperThreshold, lowVolumeThreshold string) string {
	var lowVolumeThresholdPart, upperThresholdPart string
	if lowVolumeThreshold != "" {
		lowVolumeThresholdPart = " > " + lowVolumeThreshold
	}
	if upperThreshold != "" {
		upperThresholdPart = " < " + upperThreshold
	}

	return fmt.Sprintf(
		"100 * (%s) / (%s%s) > %s%s",
		promQLMetricSum,
		promQLTotalSum,
		lowVolumeThresholdPart,
		threshold,
		upperThresholdPart,
	)
}

func baselineIncreasePromQL(promQLMetric, promQLBaseline string, threshold, upperThreshold string) string {
	var upperThresholdPart string
	if upperThreshold != "" {
		upperThresholdPart = " < " + upperThreshold
	}

	return fmt.Sprintf(
		"100 * ((%s) - (%s)) / (%s) > %s%s",
		promQLMetric,
		promQLBaseline,
		promQLBaseline,
		threshold,
		upperThresholdPart,
	)
}

func buildHealthAnnotation(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, threshold string, override map[string]any) ([]byte, error) {
	// The health annotation contains json-encoded information used in console plugin display
	annotation := map[string]any{
		"threshold": threshold,
		"unit":      "%",
	}
	switch alert.GroupBy {
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
		return nil, fmt.Errorf("cannot encode alert annotation [template=%s]: %w", template, err)
	}
	return bAnnot, nil
}

func buildLabels(severity string, forHealth bool) map[string]string {
	m := map[string]string{
		"severity": strings.ToLower(severity),
		"app":      "netobserv", // means that the rule is created by netobserv
	}
	if forHealth {
		m["netobserv"] = "true" // means that the rule should be fetched by netobserv console plugin for health
	}
	return m
}

func getMetricsForAlert(template flowslatest.AlertTemplate, alertDef *flowslatest.AlertVariant, includeList []string) (string, string) {
	var reqMetric1, reqMetric2 string
	reqMetrics1, reqMetrics2 := flowslatest.GetElligibleMetricsForAlert(template, alertDef)
	if len(reqMetrics1) > 0 {
		reqMetric1 = flowslatest.GetFirstRequiredMetrics(reqMetrics1, includeList)
	}
	if len(reqMetrics2) > 0 {
		reqMetric2 = flowslatest.GetFirstRequiredMetrics(reqMetrics2, includeList)
	}
	return reqMetric1, reqMetric2
}
