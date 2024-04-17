package metrics

import (
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
)

func GetFilters(fm *metricslatest.FlowMetricSpec) []metricslatest.MetricFilter {
	var filters []metricslatest.MetricFilter
	if fm.Direction == metricslatest.Egress {
		filters = append(filters, metricslatest.MetricFilter{
			Field:     "FlowDirection",
			Value:     "1|2",
			MatchType: metricslatest.MatchRegex,
		})
	} else if fm.Direction == metricslatest.Ingress {
		filters = append(filters, metricslatest.MetricFilter{
			Field:     "FlowDirection",
			Value:     "0|2",
			MatchType: metricslatest.MatchRegex,
		})
	}
	return append(fm.Filters, filters...)
}
