package alerts

import (
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

type ingressLatencyTrend struct {
	ctx *ruleContext
}

func newIngressLatencyTrend(ctx *ruleContext) HealthRule {
	return &ingressLatencyTrend{ctx: ctx}
}

func (r *ingressLatencyTrend) RecordingName() string {
	return buildRecordingRuleName(r.ctx, "ingress_http_latency_avg", "2m")
}

func (r *ingressLatencyTrend) GetAnnotations() (map[string]string, error) {
	var legend string

	switch r.ctx.healthRule.GroupBy {
	case flowslatest.GroupByNode:
		return nil, fmt.Errorf("IngressLatencyTrend health rule does not support grouping per node")
	case flowslatest.GroupByNamespace:
		legend = " [namespace={{ $labels.namespace }}]"
	case flowslatest.GroupByWorkload:
		return nil, fmt.Errorf("IngressLatencyTrend health rule does not support grouping per workload")
	}

	offset, _ := r.ctx.healthRule.GetTrendParams()
	healthAnnot := newHealthAnnotation(r.ctx)
	healthAnnot.CloseOpenScale(r.ctx, 5)

	return map[string]string{
		"summary": "Ingress latency increased",
		"description": fmt.Sprintf(
			"HAProxy ingress average response latency increased by more than %s%%%s, compared to baseline (offset: %s).",
			r.ctx.getLowestThreshold(),
			legend,
			offset,
		),
		"runbook_url":       buildRunbookURL(r.ctx.template),
		healthAnnotationKey: encodeHealthAnnotation(healthAnnot),
	}, nil
}

func (r *ingressLatencyTrend) Build() (*monitoringv1.Rule, error) {
	if r.ctx.side == asDest {
		return nil, nil
	}
	offset, _ := r.ctx.healthRule.GetTrendParams()

	var currentMetric, baselineMetric string
	// TODO: check p90 rather than avg?
	if r.ctx.healthRule.GroupBy == flowslatest.GroupByNamespace {
		// Rename exported_namespace to namespace for console plugin compatibility
		currentMetric = `avg(label_replace(haproxy_server_http_average_response_latency_milliseconds, "namespace", "$1", "exported_namespace", "(.*)")) by (namespace)`
		baselineMetric = fmt.Sprintf(`avg(label_replace(haproxy_server_http_average_response_latency_milliseconds offset %s, "namespace", "$1", "exported_namespace", "(.*)")) by (namespace)`, offset)
	} else {
		// Global (no groupBy)
		currentMetric = `avg(haproxy_server_http_average_response_latency_milliseconds)`
		baselineMetric = fmt.Sprintf(`avg(haproxy_server_http_average_response_latency_milliseconds offset %s)`, offset)
	}
	promql := baselineIncreasePromQL(currentMetric, baselineMetric, r.ctx.alertThreshold, r.ctx.upperThreshold)
	return createRule(r.ctx, r, promql)
}
