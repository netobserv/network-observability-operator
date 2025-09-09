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

func tooManyKernelDrops(alert *flowslatest.AlertVariant, severity, threshold, upperThreshold, addtnlDesc string) (*monitoringv1.Rule, error) {
	const tpl = flowslatest.AlertPacketDropsByKernel
	d := monitoringv1.Duration("5m")

	metrics, totalMetrics := flowslatest.GetElligibleMetricsForAlert(tpl, alert)
	metricsRate := promQLRateFromElligibleMetrics(metrics)
	totalRate := promQLRateFromElligibleMetrics(totalMetrics)
	aggregatedMetricsSum := aggregateSourceDest(metricsRate, alert.GroupBy)
	aggregatedTotalSum := aggregateSourceDest(totalRate, alert.GroupBy)

	promql := percentagePromQL(aggregatedMetricsSum, aggregatedTotalSum, threshold, upperThreshold, alert.LowVolumeThreshold)

	bAnnot, err := buildHealthAnnotation(tpl, alert, threshold, nil)
	if err != nil {
		return nil, err
	}

	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of packets dropped by the kernel%s. %s",
		threshold,
		getAlertLegend(alert),
		addtnlDesc,
	)
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", tpl, strings.ToUpper(string(severity[0])), alert.GroupBy),
		Annotations: map[string]string{
			"description":                 description,
			"summary":                     "Too many drops by the kernel",
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promql),
		For:    &d,
		Labels: buildLabels(severity, true),
	}, nil

}

func tooManyDeviceDrops(alert *flowslatest.AlertVariant, severity, threshold, upperThreshold, addtnlDesc string) (*monitoringv1.Rule, error) {
	const tpl = flowslatest.AlertPacketDropsByNetDev
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
		return nil, fmt.Errorf("PacketDropsByNetDev alert does not support grouping per namespace")
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("PacketDropsByNetDev alert does not support grouping per workload")
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

	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", tpl, strings.ToUpper(string(severity[0])), alert.GroupBy),
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
