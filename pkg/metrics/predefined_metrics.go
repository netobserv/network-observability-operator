package metrics

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/pkg/helper"
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
	latencyBuckets = []string{".005", ".01", ".02", ".03", ".04", ".05", ".075", ".1", ".25", "1"}
	mapLabels      = map[string][]string{
		tagNodes:      {"K8S_Clustername", "SrcK8S_Zone", "DstK8S_Zone", "SrcK8S_HostName", "DstK8S_HostName"},
		tagNamespaces: {"K8S_Clustername", "SrcK8S_Zone", "DstK8S_Zone", "SrcK8S_Namespace", "DstK8S_Namespace", "K8S_FlowLayer", "SrcSubnetLabel", "DstSubnetLabel"},
		tagWorkloads:  {"K8S_Clustername", "SrcK8S_Zone", "DstK8S_Zone", "SrcK8S_Namespace", "DstK8S_Namespace", "K8S_FlowLayer", "SrcSubnetLabel", "DstSubnetLabel", "SrcK8S_OwnerName", "DstK8S_OwnerName", "SrcK8S_OwnerType", "DstK8S_OwnerType", "SrcK8S_Type", "DstK8S_Type"},
	}
	mapValueFields = map[string]string{
		tagBytes:   "Bytes",
		tagPackets: "Packets",
	}
	predefinedMetrics []taggedMetricDefinition
	// Note that we set default in-code rather than in CRD, in order to keep track of value being unset or set intentionnally in FlowCollector
	DefaultIncludeList = []string{
		"node_ingress_bytes_total",
		"workload_ingress_bytes_total",
		"namespace_flows_total",
		"namespace_drop_packets_total",
		"namespace_rtt_seconds",
		"namespace_dns_latency_seconds",
	}
	// More metrics enabled when Loki is disabled, to avoid loss of information
	DefaultIncludeListLokiDisabled = []string{
		"node_ingress_bytes_total",
		"workload_ingress_bytes_total",
		"workload_flows_total",
		"workload_drop_bytes_total",
		"workload_drop_packets_total",
		"workload_rtt_seconds",
		"workload_dns_latency_seconds",
	}
	// Pre-deprecation default IgnoreTags list (1.4) - used before switching to whitelist approach,
	// to make sure there is no unintended new metrics being collected
	// Don't add anything here: this is not meant to evolve
	defaultIgnoreTags1_4 = []string{"egress", "packets", "nodes-flows", "namespaces-flows", "workloads-flows", "namespaces"}
)

type taggedMetricDefinition struct {
	metricslatest.FlowMetricSpec
	tags []string
}

func init() {
	for _, group := range []string{tagNodes, tagNamespaces, tagWorkloads} {
		groupTrimmed := strings.TrimSuffix(group, "s")
		labels := mapLabels[group]
		// Bytes / packets metrics
		for _, vt := range []string{tagBytes, tagPackets} {
			valueField := mapValueFields[vt]
			for _, dir := range []metricslatest.FlowDirection{metricslatest.Egress, metricslatest.Ingress} {
				lowDir := strings.ToLower(string(dir))
				predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
					FlowMetricSpec: metricslatest.FlowMetricSpec{
						MetricName: fmt.Sprintf("%s_%s_%s_total", groupTrimmed, lowDir, vt),
						Type:       metricslatest.CounterMetric,
						ValueField: valueField,
						Direction:  dir,
						Labels:     labels,
						Charts:     trafficCharts(group, vt, lowDir),
					},
					tags: []string{group, vt, lowDir},
				})
			}
		}
		// Flows metrics
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			FlowMetricSpec: metricslatest.FlowMetricSpec{
				MetricName: fmt.Sprintf("%s_flows_total", groupTrimmed),
				Type:       "counter",
				Labels:     labels,
			},
			tags: []string{group, group + "-flows", "flows"},
		})
		// RTT metrics
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			FlowMetricSpec: metricslatest.FlowMetricSpec{
				MetricName: fmt.Sprintf("%s_rtt_seconds", groupTrimmed),
				Type:       metricslatest.HistogramMetric,
				ValueField: "TimeFlowRttNs",
				Filters: []metricslatest.MetricFilter{
					{Field: "TimeFlowRttNs", MatchType: metricslatest.MatchPresence},
				},
				Labels:  labels,
				Divider: "1000000000", // ns => s
				Buckets: latencyBuckets,
				Charts:  rttCharts(group),
			},
			tags: []string{group, "rtt"},
		})
		// Drops metrics
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			FlowMetricSpec: metricslatest.FlowMetricSpec{
				MetricName: fmt.Sprintf("%s_drop_packets_total", groupTrimmed),
				Type:       metricslatest.CounterMetric,
				ValueField: "PktDropPackets",
				Filters: []metricslatest.MetricFilter{
					{Field: "PktDropPackets", MatchType: metricslatest.MatchPresence},
				},
				Labels: labels,
				Charts: dropCharts(group, "pps"),
			},
			tags: []string{group, tagPackets, "drops"},
		})
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			FlowMetricSpec: metricslatest.FlowMetricSpec{
				MetricName: fmt.Sprintf("%s_drop_bytes_total", groupTrimmed),
				Type:       metricslatest.CounterMetric,
				ValueField: "PktDropBytes",
				Filters: []metricslatest.MetricFilter{
					{Field: "PktDropBytes", MatchType: metricslatest.MatchPresence},
				},
				Labels: labels,
				Charts: dropCharts(group, "Bps"),
			},
			tags: []string{group, tagBytes, "drop"},
		})
		// DNS metrics
		dnsLabels := labels
		dnsLabels = append(dnsLabels, "DnsFlagsResponseCode")
		predefinedMetrics = append(predefinedMetrics, taggedMetricDefinition{
			FlowMetricSpec: metricslatest.FlowMetricSpec{
				MetricName: fmt.Sprintf("%s_dns_latency_seconds", groupTrimmed),
				Type:       metricslatest.HistogramMetric,
				ValueField: "DnsLatencyMs",
				Filters: []metricslatest.MetricFilter{
					{Field: "DnsId", MatchType: metricslatest.MatchPresence},
				},
				Labels:  dnsLabels,
				Divider: "1000", // ms => s
				Buckets: latencyBuckets,
				Charts:  dnsCharts(group),
			},
			tags: []string{group, "dns"},
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

func convertIgnoreTagsToIncludeList(ignoreTags []string) []flowslatest.FLPMetric {
	ret := []flowslatest.FLPMetric{}
	for i := range predefinedMetrics {
		if !isIgnored(&predefinedMetrics[i], ignoreTags) {
			ret = append(ret, flowslatest.FLPMetric(predefinedMetrics[i].MetricName))
		}
	}
	return ret
}

func GetAsIncludeList(ignoreTags []string, includeList *[]flowslatest.FLPMetric) *[]flowslatest.FLPMetric {
	if includeList == nil && len(ignoreTags) > 0 {
		if reflect.DeepEqual(ignoreTags, defaultIgnoreTags1_4) {
			return nil
		}
		converted := convertIgnoreTagsToIncludeList(ignoreTags)
		return &converted
	}
	return includeList
}

func GetAllNames() []string {
	names := []string{}
	for i := range predefinedMetrics {
		names = append(names, predefinedMetrics[i].MetricName)
	}
	return names
}

func GetDefinitions(names []string, toRemove []string) []metricslatest.FlowMetric {
	ret := []metricslatest.FlowMetric{}
	for i := range predefinedMetrics {
		for _, name := range names {
			if predefinedMetrics[i].MetricName == name {
				spec := predefinedMetrics[i].FlowMetricSpec
				spec.Labels = removeLabels(spec.Labels, toRemove)
				ret = append(ret, metricslatest.FlowMetric{Spec: spec})
			}
		}
	}
	return ret
}

func removeLabels(initial []string, toRemove []string) []string {
	var labels []string
	for _, lbl := range initial {
		if !slices.Contains(toRemove, lbl) {
			labels = append(labels, lbl)
		}
	}
	return labels
}

func GetIncludeList(spec *flowslatest.FlowCollectorSpec) []string {
	var list []string
	if spec.Processor.Metrics.IncludeList == nil {
		if helper.UseLoki(spec) {
			list = DefaultIncludeList
		} else {
			// When loki is disabled, increase what's available through metrics by default, to minimize the loss of information
			list = DefaultIncludeListLokiDisabled
		}
	} else {
		for _, m := range *spec.Processor.Metrics.IncludeList {
			list = append(list, string(m))
		}
	}
	if !helper.IsPktDropEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_drop_")
	}
	if !helper.IsFlowRTTEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_rtt_")
	}
	if !helper.IsDNSTrackingEnabled(&spec.Agent.EBPF) {
		list = removeMetricsByPattern(list, "_dns_")
	}
	return list
}

func removeMetricsByPattern(list []string, search string) []string {
	var filtered []string
	for _, m := range list {
		if !strings.Contains(m, search) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func MergePredefined(fm []metricslatest.FlowMetric, fc *flowslatest.FlowCollectorSpec) []metricslatest.FlowMetric {
	names := GetIncludeList(fc)
	var toRemove []string
	if !helper.IsZoneEnabled(&fc.Processor) {
		toRemove = append(toRemove, "SrcK8S_Zone", "DstK8S_Zone")
	}
	if !helper.IsMultiClusterEnabled(&fc.Processor) {
		toRemove = append(toRemove, "K8S_Clustername")
	}
	predefined := GetDefinitions(names, toRemove)
	return append(predefined, fm...)
}
