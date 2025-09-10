package alerts

import (
	"fmt"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func alertNoFlows() *monitoringv1.Rule {
	d := monitoringv1.Duration("10m")

	// Not receiving flows
	return &monitoringv1.Rule{
		Alert: string(flowslatest.AlertNoFlows),
		Annotations: map[string]string{
			"description": "NetObserv flowlogs-pipeline is not receiving any flow, this is either a connection issue with the agent, or an agent issue",
			"summary":     "NetObserv flowlogs-pipeline is not receiving any flow",
		},
		Expr:   intstr.FromString("sum(rate(netobserv_ingest_flows_processed[1m])) == 0"),
		For:    &d,
		Labels: buildLabels("warning", false),
	}
}

func alertLokiError() *monitoringv1.Rule {
	d := monitoringv1.Duration("10m")

	return &monitoringv1.Rule{
		Alert: string(flowslatest.AlertLokiError),
		Annotations: map[string]string{
			"description": "NetObserv flowlogs-pipeline is dropping flows because of Loki errors, Loki may be down or having issues ingesting every flows. Please check Loki and flowlogs-pipeline logs.",
			"summary":     "NetObserv flowlogs-pipeline is dropping flows because of Loki errors",
		},
		Expr:   intstr.FromString("sum(rate(netobserv_loki_dropped_entries_total[1m])) > 0"),
		For:    &d,
		Labels: buildLabels("warning", false),
	}
}

func kernelDrops(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string, enabledMetrics []string) (*monitoringv1.Rule, error) {
	tpl := flowslatest.AlertPacketDropsByKernel
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of packets dropped by the kernel%s. %s",
		threshold,
		getAlertLegend(side, alert),
		addtnlDesc,
	)

	metric, totalMetric := getMetricsForAlert(tpl, alert, enabledMetrics)
	metricsRate := promQLRateFromMetric(metric, "", "", "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "2m", "")
	metricsSumBy := sumBy(metricsRate, alert.GroupBy, side, "")
	totalSumBy := sumBy(totalRate, alert.GroupBy, side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, threshold, upperThreshold, alert.LowVolumeThreshold)

	return createRule(
		tpl,
		alert,
		side,
		promql,
		"Too many packet drops by the kernel",
		description,
		severity,
		threshold,
		monitoringv1.Duration("5m"),
	)
}

func deviceDrops(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string) (*monitoringv1.Rule, error) {
	// No "side" consideration on netdev metrics, so keep only 1 call from the two of them
	if side == asDest {
		return nil, nil
	}
	const tpl = flowslatest.AlertPacketDropsByDevice
	d := monitoringv1.Duration("5m")

	var byLabels string
	var healthAnnotOverride map[string]any
	var legend string
	switch alert.GroupBy {
	case flowslatest.GroupByNode:
		byLabels = " by (instance)"
		healthAnnotOverride = map[string]any{"nodeLabels": "instance"}
		legend = " [node={{ $labels.instance }}]"
	case flowslatest.GroupByNamespace:
		return nil, fmt.Errorf("PacketDropsByDevice alert does not support grouping per namespace")
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("PacketDropsByDevice alert does not support grouping per workload")
	}

	promql := percentagePromQL(
		fmt.Sprintf("sum(rate(node_network_receive_drop_total[2m]))%s + sum(rate(node_network_transmit_drop_total[2m]))%s", byLabels, byLabels),
		fmt.Sprintf("sum(rate(node_network_receive_packets_total[2m]))%s + sum(rate(node_network_transmit_packets_total[2m]))%s", byLabels, byLabels),
		threshold,
		upperThreshold,
		alert.LowVolumeThreshold,
	)

	bAnnot, err := buildHealthAnnotation(tpl, alert, threshold, healthAnnotOverride)
	if err != nil {
		return nil, err
	}

	var gr string
	if alert.GroupBy != "" {
		gr = "Per" + string(alert.GroupBy)
	}
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", tpl, gr, strings.ToUpper(severity[:1])+severity[1:]),
		Annotations: map[string]string{
			"description":                 fmt.Sprintf("node-exporter is detecting more than %s%% of dropped packets%s. %s", threshold, legend, addtnlDesc),
			"summary":                     "Too many drops from device",
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promql),
		For:    &d,
		Labels: buildLabels(severity, true),
	}, nil
}

func ipsecErrors(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string, enabledMetrics []string) (*monitoringv1.Rule, error) {
	tpl := flowslatest.AlertIPsecErrors
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of IPsec errors%s. %s",
		threshold,
		getAlertLegend(side, alert),
		addtnlDesc,
	)

	metric, totalMetric := getMetricsForAlert(tpl, alert, enabledMetrics)
	metricsRate := promQLRateFromMetric(metric, "", "", "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "2m", "")
	metricsSumBy := sumBy(metricsRate, alert.GroupBy, side, "")
	totalSumBy := sumBy(totalRate, alert.GroupBy, side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, threshold, upperThreshold, alert.LowVolumeThreshold)

	return createRule(
		tpl,
		alert,
		side,
		promql,
		"Too many IPsec errors",
		description,
		severity,
		threshold,
		monitoringv1.Duration("5m"),
	)
}

func dnsErrors(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string, enabledMetrics []string) (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if side == asSource {
		return nil, nil
	}
	tpl := flowslatest.AlertDNSErrors
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of DNS errors%s. %s",
		threshold,
		getAlertLegend(side, alert),
		addtnlDesc,
	)

	metric, totalMetric := getMetricsForAlert(tpl, alert, enabledMetrics)
	metricsRate := promQLRateFromMetric(metric, "_count", `{DnsFlagsResponseCode!="NoError"}`, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", "", "2m", "")
	metricsSumBy := sumBy(metricsRate, alert.GroupBy, side, "")
	totalSumBy := sumBy(totalRate, alert.GroupBy, side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, threshold, upperThreshold, alert.LowVolumeThreshold)

	return createRule(
		tpl,
		alert,
		side,
		promql,
		"Too many DNS errors",
		description,
		severity,
		threshold,
		monitoringv1.Duration("5m"),
	)
}

func netpolDenied(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string, enabledMetrics []string) (*monitoringv1.Rule, error) {
	tpl := flowslatest.AlertNetpolDenied
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of denied traffic due to Network Policies%s. %s",
		threshold,
		getAlertLegend(side, alert),
		addtnlDesc,
	)

	metric, totalMetric := getMetricsForAlert(tpl, alert, enabledMetrics)
	metricsRate := promQLRateFromMetric(metric, "", `{action="drop"}`, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", "", "2m", "")
	metricsSumBy := sumBy(metricsRate, alert.GroupBy, side, "")
	totalSumBy := sumBy(totalRate, alert.GroupBy, side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, threshold, upperThreshold, alert.LowVolumeThreshold)

	return createRule(
		tpl,
		alert,
		side,
		promql,
		"Traffic denied by Network Policies",
		description,
		severity,
		threshold,
		monitoringv1.Duration("5m"),
	)
}

func latencyTrend(alert *flowslatest.AlertVariant, side srcOrDst, severity, threshold, upperThreshold, addtnlDesc string, enabledMetrics []string) (*monitoringv1.Rule, error) {
	tpl := flowslatest.AlertLatencyHighTrend
	offset, duration := alert.GetTrendParams()
	description := fmt.Sprintf(
		"NetObserv is detecting TCP latency increased by more than %s%%%s, compared to baseline (offset: %s). %s",
		threshold,
		getAlertLegend(side, alert),
		offset,
		addtnlDesc,
	)

	metric, baseline := getMetricsForAlert(tpl, alert, enabledMetrics)
	metricsRate := promQLRateFromMetric(metric, "_bucket", "", "2m", "")
	baselineRate := promQLRateFromMetric(baseline, "_bucket", "", duration, " offset "+offset)
	metricQuantile := histogramQuantile(metricsRate, alert.GroupBy, side, "0.9")
	baselineQuantile := histogramQuantile(baselineRate, alert.GroupBy, side, "0.9")
	promql := baselineIncreasePromQL(metricQuantile, baselineQuantile, threshold, upperThreshold)

	return createRule(
		tpl,
		alert,
		side,
		promql,
		"TCP latency increase",
		description,
		severity,
		threshold,
		monitoringv1.Duration("5m"),
	)
}
