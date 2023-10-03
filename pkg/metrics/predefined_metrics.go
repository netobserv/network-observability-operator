package metrics

import (
	"fmt"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
)

const (
	tagNamespaces = "namespaces"
	tagNodes      = "nodes"
	tagWorkloads  = "workloads"
	tagIngress    = "ingress"
	tagEgress     = "egress"
	tagBytes      = "bytes"
	tagPackets    = "packets"
)

var (
	mapLabels = map[string][]string{
		tagNodes:      {"SrcK8S_HostName", "DstK8S_HostName"},
		tagNamespaces: {"SrcK8S_Namespace", "DstK8S_Namespace"},
		tagWorkloads:  {"SrcK8S_Namespace", "DstK8S_Namespace", "SrcK8S_OwnerName", "DstK8S_OwnerName", "SrcK8S_OwnerType", "DstK8S_OwnerType"},
	}
	mapValueFields = map[string]string{
		tagBytes:   "Bytes",
		tagPackets: "Packets",
	}
	mapDirection = map[string]string{
		tagIngress: "0",
		tagEgress:  "1",
	}
	predefinedMetrics []taggedMetricDefinition
)

type taggedMetricDefinition struct {
	flowslatest.MetricDefinition
	tags []string
}

func init() {
	for _, group := range []string{tagNodes, tagNamespaces, tagWorkloads} {
		groupTrimmed := strings.TrimSuffix(group, "s")
		labels := mapLabels[group]
		// Bytes / packets metrics
		for _, vt := range []string{tagBytes, tagPackets} {
			valueField := mapValueFields[vt]
			for _, dir := range []string{tagEgress, tagIngress} {
				predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
					MetricDefinition: flowslatest.MetricDefinition{
						Name:       fmt.Sprintf("%s_%s_%s_total", groupTrimmed, dir, vt),
						Type:       flowslatest.CounterMetric,
						ValueField: valueField,
						Filters: []flowslatest.MetricFilter{
							{Field: "Duplicate", Value: "false"},
							{Field: "FlowDirection", Value: mapDirection[dir]},
						},
						Labels: labels,
					},
					tags: []string{group, vt, dir},
				})
			}
		}
		// Flows metrics
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			MetricDefinition: flowslatest.MetricDefinition{
				Name:   fmt.Sprintf("%s_flows_total", groupTrimmed),
				Type:   flowslatest.CounterMetric,
				Labels: labels,
			},
			tags: []string{group, group + "-flows", "flows"},
		})
	}
}

func isIgnored(def *taggedMetricDefinition, ignoreTags []string) bool {
	for _, ignoreTag := range ignoreTags {
		for _, tag := range def.tags {
			if ignoreTag == tag {
				return true
			}
		}
	}
	return false
}

func GetEnabledMetrics(ignoreTags []string) []flowslatest.MetricDefinition {
	ret := []flowslatest.MetricDefinition{}
	for i := range predefinedMetrics {
		if !isIgnored(&predefinedMetrics[i], ignoreTags) {
			ret = append(ret, predefinedMetrics[i].MetricDefinition)
		}
	}
	return ret
}
