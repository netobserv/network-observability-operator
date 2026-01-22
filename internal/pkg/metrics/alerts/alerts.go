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
	metric, totalMetric := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", filter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", filter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.healthRule.GroupBy, rb.side, "")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.healthRule.LowVolumeThreshold, isRecording)

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend())
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) deviceDrops() (*monitoringv1.Rule, error) {
	// No "side" consideration on netdev metrics, so keep only 1 call from the two of them
	if rb.side == asDest {
		return nil, nil
	}
	var byLabels string
	var healthAnnotOverride map[string]any
	var legend string
	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		byLabels = " by (instance)"
		healthAnnotOverride = map[string]any{"nodeLabels": "instance"}
		legend = " [node={{ $labels.instance }}]"
	case flowslatest.GroupByNamespace:
		return nil, fmt.Errorf("PacketDropsByDevice health rule does not support grouping per namespace")
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("PacketDropsByDevice health rule does not support grouping per workload")
	}

	isRecording := rb.mode == flowslatest.ModeRecording
	promql := percentagePromQL(
		fmt.Sprintf("sum(rate(node_network_receive_drop_total[2m]))%s + sum(rate(node_network_transmit_drop_total[2m]))%s", byLabels, byLabels),
		fmt.Sprintf("sum(rate(node_network_receive_packets_total[2m]))%s + sum(rate(node_network_transmit_packets_total[2m]))%s", byLabels, byLabels),
		rb.threshold,
		rb.upperThreshold,
		rb.healthRule.LowVolumeThreshold,
		isRecording,
	)

	bAnnot, err := rb.buildHealthAnnotation(healthAnnotOverride)
	if err != nil {
		return nil, err
	}

	var gr string
	if rb.healthRule.GroupBy != "" {
		gr = "Per" + string(rb.healthRule.GroupBy)
	}

	// Generate recording rule
	if rb.mode == flowslatest.ModeRecording {
		recordName := rb.buildRecordingRuleName()
		return &monitoringv1.Rule{
			Record: recordName,
			// Note: Recording rules cannot have annotations in Prometheus
			Expr:   intstr.FromString(promql),
			Labels: buildRecordingRuleLabels(string(rb.template)),
		}, nil
	}

	// Generate alert rule
	description := rb.buildDescriptionFromTemplate(rb.threshold, legend)
	summary := rb.getSummaryFromTemplate()
	return &monitoringv1.Rule{
		Alert: fmt.Sprintf("%s_%s%s", rb.template, gr, strings.ToUpper(rb.severity[:1])+rb.severity[1:]),
		Annotations: map[string]string{
			"description":                 description,
			"summary":                     summary,
			"netobserv_io_network_health": string(bAnnot),
		},
		Expr:   intstr.FromString(promql),
		For:    &rb.duration,
		Labels: buildLabels(rb.severity, true),
	}, nil
}

func (rb *ruleBuilder) ipsecErrors() (*monitoringv1.Rule, error) {
	metric, totalMetric := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", filter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", filter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.healthRule.GroupBy, rb.side, "")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.healthRule.LowVolumeThreshold, isRecording)

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend())
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) dnsErrors() (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if rb.side == asSource {
		return nil, nil
	}

	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`DnsFlagsResponseCode!~"NoError|NXDomain"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_count", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.healthRule.GroupBy, rb.side, "")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.healthRule.LowVolumeThreshold, isRecording)

	rb.trafficLink = &trafficLink{
		BackAndForth:      false,
		ExtraFilter:       `dns_flag_response_code!="NoError,NXDomain"`,
		FilterDestination: true,
	}

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend())
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) dnsNxDomainErrors() (*monitoringv1.Rule, error) {
	// DNS errors are in return traffic only
	if rb.side == asSource {
		return nil, nil
	}

	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`DnsFlagsResponseCode="NXDomain"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_count", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "_count", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.healthRule.GroupBy, rb.side, "")
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.healthRule.LowVolumeThreshold, rb.mode == flowslatest.ModeRecording)

	rb.trafficLink = &trafficLink{
		BackAndForth:      false,
		ExtraFilter:       `dns_flag_response_code="NXDomain"`,
		FilterDestination: true,
	}

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend())
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) netpolDenied() (*monitoringv1.Rule, error) {
	metric, totalMetric := rb.getMetricsForAlert()
	metricsFilter := rb.buildLabelFilter(`action="drop"`)
	totalFilter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "", metricsFilter, "2m", "")
	totalRate := promQLRateFromMetric(totalMetric, "", totalFilter, "2m", "")
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	totalSumBy := sumBy(totalRate, rb.healthRule.GroupBy, rb.side, "")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := percentagePromQL(metricsSumBy, totalSumBy, rb.threshold, rb.upperThreshold, rb.healthRule.LowVolumeThreshold, isRecording)

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend())
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) latencyTrend() (*monitoringv1.Rule, error) {
	offset, duration := rb.healthRule.GetTrendParams()

	metric, baseline := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter("")
	metricsRate := promQLRateFromMetric(metric, "_bucket", filter, "2m", "")
	baselineRate := promQLRateFromMetric(baseline, "_bucket", filter, duration, " offset "+offset)
	metricQuantile := histogramQuantile(metricsRate, rb.healthRule.GroupBy, rb.side, "0.9")
	baselineQuantile := histogramQuantile(baselineRate, rb.healthRule.GroupBy, rb.side, "0.9")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := baselineIncreasePromQL(metricQuantile, baselineQuantile, rb.threshold, rb.upperThreshold, isRecording)

	// trending comparison are on an open scale; but in the health page, we need a closed scale to compute the score
	// let's set an upper bound to max(5*threshold, 100) so score can be computed after clamping
	val, err := strconv.ParseFloat(rb.threshold, 64)
	if err != nil {
		return nil, err
	}
	rb.upperValueRange = strconv.Itoa(int(math.Max(val*5, 100)))

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend(), offset)
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) externalTrend(ingress bool) (*monitoringv1.Rule, error) {
	// Don't create ingress for asSource, or egress for asDestination
	if rb.side == asSource && ingress {
		return nil, nil
	} else if rb.side == asDest && !ingress {
		return nil, nil
	}

	filterForExternal := `DstSubnetLabel=~"|EXT:.*",DstK8S_Namespace="",DstK8S_OwnerName=""`
	trafficLinkFilter := `dst_subnet_label="",EXT:`
	if ingress {
		filterForExternal = `SrcSubnetLabel=~"|EXT:.*",SrcK8S_Namespace="",SrcK8S_OwnerName=""`
		trafficLinkFilter = `src_subnet_label="",EXT:`
	}
	offset, duration := rb.healthRule.GetTrendParams()

	metric, baseline := rb.getMetricsForAlert()
	filter := rb.buildLabelFilter(filterForExternal)
	metricsRate := promQLRateFromMetric(metric, "", filter, "2m", "")
	baselineRate := promQLRateFromMetric(baseline, "", filter, duration, " offset "+offset)
	metricsSumBy := sumBy(metricsRate, rb.healthRule.GroupBy, rb.side, "")
	baselineSumBy := sumBy(baselineRate, rb.healthRule.GroupBy, rb.side, "")
	isRecording := rb.mode == flowslatest.ModeRecording
	promql := baselineIncreasePromQL(metricsSumBy, baselineSumBy, rb.threshold, rb.upperThreshold, isRecording)

	// trending comparison are on an open scale; but in the health page, we need a closed scale to compute the score
	// let's set an upper bound to max(5*threshold, 100) so score can be computed after clamping
	val, err := strconv.ParseFloat(rb.threshold, 64)
	if err != nil {
		return nil, err
	}
	rb.upperValueRange = strconv.Itoa(int(math.Max(val*5, 100)))

	rb.trafficLink = &trafficLink{
		BackAndForth:      true,
		ExtraFilter:       trafficLinkFilter,
		FilterDestination: rb.side == asDest,
	}

	description := rb.buildDescriptionFromTemplate(rb.threshold, rb.getAlertLegend(), offset)
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) ingressErrors() (*monitoringv1.Rule, error) {
	// No "side" consideration for ingress metrics, so keep only 1 call from the two of them
	if rb.side == asDest {
		return nil, nil
	}

	var legend string
	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		return nil, fmt.Errorf("Ingress5xxErrors health rule does not support grouping per node")
	case flowslatest.GroupByNamespace:
		// Note: we'll rename exported_namespace to namespace in the PromQL using label_replace
		legend = " [namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("Ingress5xxErrors health rule does not support grouping per workload")
	}

	isRecording := rb.mode == flowslatest.ModeRecording
	// 5xx errors
	errorCode := "5xx"

	// Build PromQL with label_replace to rename exported_namespace to namespace
	var errorsQuery, totalQuery string
	if rb.healthRule.GroupBy == flowslatest.GroupByNamespace {
		// Rename exported_namespace to namespace for console plugin compatibility
		errorsQuery = fmt.Sprintf("sum(label_replace(rate(haproxy_server_http_responses_total{code=\"%s\"}[2m]), \"namespace\", \"$1\", \"exported_namespace\", \"(.*)\")) by (namespace)", errorCode)
		totalQuery = "sum(label_replace(rate(haproxy_server_http_responses_total[2m]), \"namespace\", \"$1\", \"exported_namespace\", \"(.*)\")) by (namespace)"
	} else {
		// Global (no groupBy)
		errorsQuery = fmt.Sprintf("sum(rate(haproxy_server_http_responses_total{code=\"%s\"}[2m]))", errorCode)
		totalQuery = "sum(rate(haproxy_server_http_responses_total[2m]))"
	}

	promql := percentagePromQL(
		errorsQuery,
		totalQuery,
		rb.threshold,
		rb.upperThreshold,
		"",
		isRecording,
	)

	// Set upperBound to 100 (percentage scale)
	rb.upperValueRange = "100"

	description := rb.buildDescriptionFromTemplate(rb.threshold, legend)
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}

func (rb *ruleBuilder) ingressLatencyTrend() (*monitoringv1.Rule, error) {
	if rb.side == asDest {
		return nil, nil
	}

	offset, _ := rb.healthRule.GetTrendParams()

	var currentMetric, baselineMetric string
	var legend string

	switch rb.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		return nil, fmt.Errorf("IngressLatencyTrend health rule does not support grouping per node")
	case flowslatest.GroupByNamespace:
		legend = " [namespace={{ $labels.namespace }}]"
		currentMetric = `avg(label_replace(haproxy_server_http_average_response_latency_milliseconds, "namespace", "$1", "exported_namespace", "(.*)")) by (namespace)`
		baselineMetric = fmt.Sprintf(`avg(label_replace(haproxy_server_http_average_response_latency_milliseconds offset %s, "namespace", "$1", "exported_namespace", "(.*)")) by (namespace)`, offset)
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("IngressLatencyTrend health rule does not support grouping per workload")
	default:
		currentMetric = `avg(haproxy_server_http_average_response_latency_milliseconds)`
		baselineMetric = fmt.Sprintf(`avg(haproxy_server_http_average_response_latency_milliseconds offset %s)`, offset)
	}

	isRecording := rb.mode == flowslatest.ModeRecording
	promql := baselineIncreasePromQL(currentMetric, baselineMetric, rb.threshold, rb.upperThreshold, isRecording)

	val, err := strconv.ParseFloat(rb.threshold, 64)
	if err != nil {
		return nil, err
	}
	rb.upperValueRange = strconv.Itoa(int(math.Max(val*5, 100)))

	description := rb.buildDescriptionFromTemplate(rb.threshold, legend, offset)
	summary := rb.getSummaryFromTemplate()
	return rb.createRule(promql, summary, description)
}
