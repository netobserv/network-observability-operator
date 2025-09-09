package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func BuildRules(ctx context.Context, fc *flowslatest.FlowCollectorSpec) []monitoringv1.Rule {
	log := log.FromContext(ctx)
	rules := []monitoringv1.Rule{}

	alerts := fc.GetFLPAlerts()
	for _, alert := range alerts {
		if ok, _ := alert.IsAllowed(fc); !ok {
			continue
		}
		for _, variant := range alert.Variants {
			if r, err := convertToRules(alert.Template, &variant); err != nil {
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

func convertToRules(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule
	var upperThreshold string
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
			if r, err := convertToRule(template, alert, st.s, st.t, upperThreshold); err != nil {
				return nil, err
			} else if r != nil {
				rules = append(rules, *r)
			}
			upperThreshold = st.t
		}
	}
	return rules, nil
}

func convertToRule(template flowslatest.AlertTemplate, alert *flowslatest.AlertVariant, severity, threshold, upperThreshold string) (*monitoringv1.Rule, error) {
	additionalDescription := fmt.Sprintf("You can turn off this alert by adding '%s' to spec.processor.metrics.disableAlerts in FlowCollector, or reconfigure it via spec.processor.metrics.alerts.", template)
	switch template {
	case flowslatest.AlertPacketDropsByNetDev:
		return tooManyDeviceDrops(alert, severity, threshold, upperThreshold, additionalDescription)
	case flowslatest.AlertPacketDropsByKernel:
		return tooManyKernelDrops(alert, severity, threshold, upperThreshold, additionalDescription)
	case flowslatest.AlertLokiError, flowslatest.AlertNoFlows:
		// error
	}
	return nil, fmt.Errorf("unknown alert template: %s", template)
}

type promQLRate string

func promQLRateFromElligibleMetrics(metrics []string) promQLRate {
	for i := 0; i < len(metrics); i++ {
		metrics[i] = "rate(netobserv_" + metrics[i] + "[2m])"
	}
	return promQLRate(strings.Join(metrics, " OR "))
}

func aggregateSourceDest(promQL promQLRate, groupBy flowslatest.AlertGroupBy) string {
	var nooLabels []string
	var labelsOut []string
	switch groupBy {
	case flowslatest.GroupByNode:
		nooLabels = []string{"K8S_HostName"}
		labelsOut = []string{"node"}
	case flowslatest.GroupByNamespace:
		nooLabels = []string{"K8S_Namespace"}
		labelsOut = []string{"namespace"}
	case flowslatest.GroupByWorkload:
		nooLabels = []string{"K8S_Namespace", "K8S_OwnerName", "K8S_OwnerType"}
		labelsOut = []string{"namespace", "workload", "kind"}
	}
	if len(labelsOut) > 0 {
		// promQL input is like "rate(netobserv_workload_ingress_bytes_total[1m])"
		// we need to relabel src and dst labels to the same label name in order to allow adding them
		// e.g. of desired output:
		// sum(label_replace(rate(netobserv_workload_ingress_bytes_total[1m]), "namespace", "$1", "SrcK8S_Namespace", "(.*)")) by (namespace) + sum(label_replace(rate(netobserv_workload_ingress_bytes_total[1m]), "namespace", "$1", "DstK8S_Namespace", "(.*)")) by (namespace)
		srcPromQL := string(promQL)
		dstPromQL := string(promQL)
		for i := range labelsOut {
			in := nooLabels[i]
			out := labelsOut[i]
			srcPromQL = fmt.Sprintf(`label_replace(%s, "%s", "$1", "Src%s", "(.*)")`, srcPromQL, out, in)
			dstPromQL = fmt.Sprintf(`label_replace(%s, "%s", "$1", "Dst%s", "(.*)")`, dstPromQL, out, in)
		}
		joinedLabels := strings.Join(labelsOut, ",")
		return fmt.Sprintf(
			`(sum(%s) by (%s) + sum(%s) by (%s))`, srcPromQL, joinedLabels, dstPromQL, joinedLabels,
		)
	}
	return fmt.Sprintf("sum(%s)", promQL)
}

func getAlertLegend(alert *flowslatest.AlertVariant) string {
	switch alert.GroupBy {
	case flowslatest.GroupByNode:
		return " [node={{ $labels.node }}]"
	case flowslatest.GroupByNamespace:
		return " [namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return " [workload={{ $labels.workload }} ({{ $labels.kind }})]"
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
		"100 * %s / (%s%s) > %s%s",
		promQLMetricSum,
		promQLTotalSum,
		lowVolumeThresholdPart,
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
