package dashboards

import (
	"fmt"
	"strings"

	"k8s.io/utils/strings/slices"
)

const (
	layerApps        = "Applications"
	layerInfra       = "Infrastructure"
	appsFilters1     = `SrcK8S_Namespace!~"|$NETOBSERV_NS|openshift.*"`
	appsFilters2     = `SrcK8S_Namespace=~"$NETOBSERV_NS|openshift.*",DstK8S_Namespace!~"|$NETOBSERV_NS|openshift.*"`
	infraFilters1    = `SrcK8S_Namespace=~"$NETOBSERV_NS|openshift.*"`
	infraFilters2    = `SrcK8S_Namespace!~"$NETOBSERV_NS|openshift.*",DstK8S_Namespace=~"$NETOBSERV_NS|openshift.*"`
	metricTagIngress = "ingress"
	metricTagEgress  = "egress"
	metricTagBytes   = "bytes"
	metricTagPackets = "packets"
)

var allRows []*Row

func init() {
	for _, scope := range []metricScope{srcDstNodeScope, srcDstNamespaceScope, srcDstWorkloadScope} {
		// byte/pkt rates
		for _, valueType := range []string{metricTagBytes, metricTagPackets} {
			valueTypeText := valueTypeToText(valueType)
			for _, dir := range []string{metricTagEgress, metricTagIngress} {
				title := fmt.Sprintf(
					"%s rate %s %s",
					valueTypeText,
					dirToVerb(dir),
					scope.titlePart,
				)
				metric := fmt.Sprintf("%s_%s_%s_total", scope.metricPart, dir, valueType)
				allRows = append(allRows, &Row{
					Metric: metric,
					Title:  title,
					Panels: topRatePanels(&scope, metric, scope.joinLabels(), scope.legendPart),
				})
			}
			// drops
			title := fmt.Sprintf(
				"%s drop rate %s",
				valueTypeText,
				scope.titlePart,
			)
			metric := fmt.Sprintf("%s_drop_%s_total", scope.metricPart, valueType)
			allRows = append(allRows, &Row{
				Metric: metric,
				Title:  title,
				Panels: topRatePanels(&scope, metric, scope.joinLabels(), scope.legendPart),
			})
		}
		// RTT
		title := fmt.Sprintf("Round-trip time %s (seconds, p99 and p50)", scope.titlePart)
		metric := fmt.Sprintf("%s_rtt_seconds", scope.metricPart)
		allRows = append(allRows, &Row{
			Metric: metric,
			Title:  title,
			Panels: histogramPanels(&scope, metric, scope.joinLabels(), scope.legendPart),
		})
		// DNS latency
		title = fmt.Sprintf("DNS latency %s (seconds, p99 and p50)", scope.titlePart)
		metric = fmt.Sprintf("%s_dns_latency_seconds", scope.metricPart)
		allRows = append(allRows, &Row{
			Metric: metric,
			Title:  title,
			Panels: histogramPanels(&scope, metric, scope.joinLabels(), scope.legendPart),
		})
		// DNS errors
		title = fmt.Sprintf("DNS request rate per code and %s", scope.titlePart)
		metric = fmt.Sprintf("%s_dns_latency_seconds", scope.metricPart)
		labels := scope.joinLabels() + ",DnsFlagsResponseCode"
		legend := scope.legendPart + ", {{DnsFlagsResponseCode}}"
		allRows = append(allRows, &Row{
			Metric: metric,
			Title:  title,
			Panels: topRatePanels(&scope, metric+"_count", labels, legend),
		})
	}
}

func topRatePanels(scope *metricScope, metric, labels, legend string) []Panel {
	if scope.splitAppInfra {
		return []Panel{
			// App
			{
				Title: layerApps,
				Targets: []Target{{
					Expr: scope.labelReplace(
						fmt.Sprintf(
							"topk(10,sum(rate(netobserv_%s{%s}[2m]) or rate(netobserv_%s{%s}[2m])) by (%s))",
							metric,
							appsFilters1,
							metric,
							appsFilters2,
							labels,
						),
					),
					Legend: legend,
				}},
			},
			// Infra
			{
				Title: layerInfra,
				Targets: []Target{{
					Expr: scope.labelReplace(
						fmt.Sprintf(
							"topk(10,sum(rate(netobserv_%s{%s}[2m]) or rate(netobserv_%s{%s}[2m])) by (%s))",
							metric,
							infraFilters1,
							metric,
							infraFilters2,
							labels,
						),
					),
					Legend: legend,
				}},
			},
		}
	}
	// No split
	return []Panel{{
		Targets: []Target{{
			Expr: scope.labelReplace(
				fmt.Sprintf("topk(10,sum(rate(netobserv_%s[2m])) by (%s))", metric, labels),
			),
			Legend: legend,
		}},
	}}
}

func histogramPanels(scope *metricScope, metric, labels, legend string) []Panel {
	if scope.splitAppInfra {
		appRateExpr := fmt.Sprintf(
			"rate(netobserv_%s_bucket{%s}[2m]) or rate(netobserv_%s_bucket{%s}[2m])",
			metric,
			appsFilters1,
			metric,
			appsFilters2,
		)
		infraRateExpr := fmt.Sprintf(
			"rate(netobserv_%s_bucket{%s}[2m]) or rate(netobserv_%s_bucket{%s}[2m])",
			metric,
			infraFilters1,
			metric,
			infraFilters2,
		)
		return []Panel{
			// App
			{
				Title: layerApps,
				Targets: []Target{
					histogramTarget(scope, "0.99", appRateExpr, labels, legend),
					histogramTarget(scope, "0.50", appRateExpr, labels, legend),
				},
			},
			// Infra
			{
				Title: layerInfra,
				Targets: []Target{
					histogramTarget(scope, "0.99", infraRateExpr, labels, legend),
					histogramTarget(scope, "0.50", infraRateExpr, labels, legend),
				},
			},
		}
	}
	// No split
	rateExpr := fmt.Sprintf("rate(netobserv_%s[2m])", metric)
	return []Panel{{
		Targets: []Target{
			histogramTarget(scope, "0.99", rateExpr, labels, legend),
			histogramTarget(scope, "0.50", rateExpr, labels, legend),
		},
	}}
}

func histogramTarget(scope *metricScope, quantile, rateExpr, labels, legend string) Target {
	return Target{
		Expr: scope.labelReplace(
			fmt.Sprintf(
				"topk(10,histogram_quantile(%s, sum(%s) by (le,%s)) > 0)",
				quantile,
				rateExpr,
				labels,
			),
		),
		Legend: legend + ", q=" + quantile,
	}
}

func dirToVerb(dir string) string {
	switch dir {
	case metricTagEgress:
		return "sent"
	case metricTagIngress:
		return "received"
	}
	return ""
}

func valueTypeToText(t string) string {
	switch t {
	case metricTagBytes:
		return "Byte"
	case metricTagPackets:
		return "Packet"
	}
	return ""
}

func CreateFlowMetricsDashboard(netobsNs string, metrics []string) (string, error) {
	var rows []*Row
	for _, ri := range allRows {
		if slices.Contains(metrics, ri.Metric) {
			rows = append(rows, ri)
		} else if strings.Contains(ri.Metric, "namespace_") {
			// namespace-based panels can also be displayed using workload-based metrics
			// Try again, replacing *_namespace_* with *_workload_*
			equivalentMetric := strings.Replace(ri.Metric, "namespace_", "workload_", 1)
			if slices.Contains(metrics, equivalentMetric) {
				clone := ri.replaceMetric(equivalentMetric)
				rows = append(rows, clone)
			}
		}
	}
	d := Dashboard{Rows: rows}
	return d.ToGrafanaJSON(netobsNs), nil
}
