package metrics

import (
	"fmt"
	"reflect"
	"strings"

	flpapi "github.com/netobserv/flowlogs-pipeline/pkg/api"
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
		tagIngress: "0|2",
		tagEgress:  "1|2",
	}
	predefinedMetrics []taggedMetricDefinition
	// Note that we set default in-code rather than in CRD, in order to keep track of value being unset or set intentionnally in FlowCollector
	DefaultIncludeList = []string{"node_ingress_bytes_total", "workload_ingress_bytes_total", "namespace_flows_total"}
	// Pre-deprecation default IgnoreTags list (1.4) - used before switching to whitelist approach,
	// to make sure there is no unintended new metrics being collected
	// Don't add anything here: this is not meant to evolve
	defaultIgnoreTags1_4 = []string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}
)

type taggedMetricDefinition struct {
	flpapi.PromMetricsItem
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
					PromMetricsItem: flpapi.PromMetricsItem{
						Name:     fmt.Sprintf("%s_%s_%s_total", groupTrimmed, dir, vt),
						Type:     "counter",
						ValueKey: valueField,
						Filters: []flpapi.PromMetricsFilter{
							{Key: "Duplicate", Value: "false"},
							{Key: "FlowDirection", Value: mapDirection[dir], Type: "regex"},
						},
						Labels: labels,
					},
					tags: []string{group, vt, dir},
				})
			}
		}
		// Flows metrics
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			PromMetricsItem: flpapi.PromMetricsItem{
				Name:   fmt.Sprintf("%s_flows_total", groupTrimmed),
				Type:   "counter",
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

func convertIgnoreTagsToIncludeList(ignoreTags []string) []string {
	ret := []string{}
	for i := range predefinedMetrics {
		if !isIgnored(&predefinedMetrics[i], ignoreTags) {
			ret = append(ret, predefinedMetrics[i].Name)
		}
	}
	return ret
}

func GetEnabledNames(ignoreTags []string, includeList *[]string) []string {
	if includeList == nil {
		if reflect.DeepEqual(ignoreTags, defaultIgnoreTags1_4) {
			return DefaultIncludeList
		}
		return convertIgnoreTagsToIncludeList(ignoreTags)
	}
	return *includeList
}

func GetAllNames() []string {
	names := []string{}
	for i := range predefinedMetrics {
		names = append(names, predefinedMetrics[i].Name)
	}
	return names
}

func GetDefinitions(names []string) []flpapi.PromMetricsItem {
	ret := []flpapi.PromMetricsItem{}
	for i := range predefinedMetrics {
		for _, name := range names {
			if predefinedMetrics[i].Name == name {
				ret = append(ret, predefinedMetrics[i].PromMetricsItem)
			}
		}
	}
	return ret
}
