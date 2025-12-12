package alerts

import (
	"fmt"
	"math"
	"strconv"
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

func (rb *ruleBuilder) kernelDrops() (*monitoringv1.Rule, error) {
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of packets dropped by the kernel%s. %s",
		rb.threshold,
		rb.getAlertLegend(),
		rb.additionalDescription(),
	)

	metric, totalMetric := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", filter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", filter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.alert.LowVolumeThreshold)

	return rb.createRule(promql, "Too many packet drops by the kernel", description)
}

func (rb *ruleBuilder) deviceDrops() (*monitoringv1.Rule, error) {
	// No "side" consideration on netdev metrics, so keep only 1 call from the two of them
	if rb.side == asDest {
		return nil, nil
	}
	var byLabels string
	var healthAnnotOverride map[string]any
	var legend string
	switch rb.alert.GroupBy {
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
		rb.threshold,
		rb.upperThreshold,
		rb.alert.LowVolumeThreshold,
	)

	bAnnot, err := rb.buildHealthAnnotation(healthAnnotOverride)
	if err != nil {
		return nil, err
	}

	var gr string
	if rb.alert.GroupBy != "" {
		gr = "Per" + string(rb.alert.GroupBy)
	}
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", rb.template, gr, strings.ToUpper(rb.severity[:1])+rb.severity[1:]),
		Annotations: map[string]string{
			"description":                 fmt.Sprintf("node-exporter is detecting more than %s%% of dropped packets%s. %s", rb.threshold, legend, rb.additionalDescription()),
			"summary":                     "Too many drops from device",
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promql),
		For:    &rb.duration,
		Labels: buildLabels(rb.severity, true),
	}, nil
}

func (rb *ruleBuilder) ipsecErrors() (*monitoringv1.Rule, error) {
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of IPsec errors%s. %s",
		rb.threshold,
		rb.getAlertLegend(),
		rb.additionalDescription(),
	)

	metric, totalMetric := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", filter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", filter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.alert.LowVolumeThreshold)

	return rb.createRule(promql, "Too many IPsec errors", description)
}

func (rb *ruleBuilder) dnsErrors() (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if rb.side == asSource {
		return nil, nil
	}
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of DNS errors%s (other than NX_DOMAIN). %s",
		rb.threshold,
		rb.getAlertLegend(),
		rb.additionalDescription(),
	)

	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`DnsFlagsResponseCode!~"NoError|NXDomain"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_count", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.alert.LowVolumeThreshold)

	rb.trafficLink = &trafficLink{
		BackAndForth:      false,
		ExtraFilter:       `dns_flag_response_code!="NoError,NXDomain"`,
		FilterDestination: true,
	}

	return rb.createRule(promql, "Too many DNS errors", description)
}

func (rb *ruleBuilder) dnsNxDomainErrors() (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if rb.side == asSource {
		return nil, nil
	}
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of DNS NX_DOMAIN errors%s. In Kubernetes, this is a common error due to the resolution using several search suffixes. It can be optimized by using trailing dots in domain names. %s",
		rb.threshold,
		rb.getAlertLegend(),
		rb.additionalDescription(),
	)

	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`DnsFlagsResponseCode="NXDomain"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_count", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.alert.LowVolumeThreshold)

	rb.trafficLink = &trafficLink{
		BackAndForth:      false,
		ExtraFilter:       `dns_flag_response_code="NXDomain"`,
		FilterDestination: true,
	}

	return rb.createRule(promql, "Too many DNS NX_DOMAIN errors", description)
}

func (rb *ruleBuilder) netpolDenied() (*monitoringv1.Rule, error) {
	description := fmt.Sprintf(
		"NetObserv is detecting more than %s%% of denied traffic due to Network Policies%s. %s",
		rb.threshold,
		rb.getAlertLegend(),
		rb.additionalDescription(),
	)

	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`action="drop"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.alert.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.alert.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.alert.LowVolumeThreshold)

	return rb.createRule(promql, "Traffic denied by Network Policies", description)
}

func (rb *ruleBuilder) latencyTrend() (*monitoringv1.Rule, error) {
	offset, duration := rb.alert.GetTrendParams()
	description := fmt.Sprintf(
		"NetObserv is detecting TCP latency increased by more than %s%%%s, compared to baseline (offset: %s). %s",
		rb.threshold,
		rb.getAlertLegend(),
		offset,
		rb.additionalDescription(),
	)

	metric, baseline := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_bucket", filter, "2m", "")
	baselineRate := promQLRateFromMetric(baseline, "_bucket", filter, duration, " offset "+offset)
	metricQuantile := histogramQuantile(metricsRate, rb.alert.GroupBy, rb.side, "0.9")
	baselineQuantile := histogramQuantile(baselineRate, rb.alert.GroupBy, rb.side, "0.9")
	promql := baselineIncreasePromQL(metricQuantile, baselineQuantile, rb.threshold, rb.upperThreshold)

	// trending comparison are on an open scale; but in the health page, we need a closed scale to compute the score
	// let's set an upper bound to max(5*threshold, 100) so score can be computed after clamping
	val, err := strconv.ParseFloat(rb.threshold, 64)
	if err != nil {
		return nil, err
	}
	rb.upperValueRange = strconv.Itoa(int(math.Max(val*5, 100)))

	return rb.createRule(promql, "TCP latency increase", description)
}
