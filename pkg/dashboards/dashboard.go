package dashboards

import (
	"fmt"
	"strings"

	"k8s.io/utils/strings/slices"
)

type rowInfo struct {
	metric    string
	group     string
	dir       string
	valueType string
}

// Queries
const (
	layerApps            = "Applications"
	layerInfra           = "Infrastructure"
	appsFilters1         = `SrcK8S_Namespace!~"|$NETOBSERV_NS|openshift.*"`
	appsFilters2         = `SrcK8S_Namespace=~"$NETOBSERV_NS|openshift.*",DstK8S_Namespace!~"|$NETOBSERV_NS|openshift.*"`
	infraFilters1        = `SrcK8S_Namespace=~"$NETOBSERV_NS|openshift.*"`
	infraFilters2        = `SrcK8S_Namespace!~"$NETOBSERV_NS|openshift.*",DstK8S_Namespace=~"$NETOBSERV_NS|openshift.*"`
	metricTagNamespaces  = "namespaces"
	metricTagNodes       = "nodes"
	metricTagWorkloads   = "workloads"
	metricTagIngress     = "ingress"
	metricTagEgress      = "egress"
	metricTagBytes       = "bytes"
	metricTagPackets     = "packets"
	metricTagDropBytes   = "drop_bytes"
	metricTagDropPackets = "drop_packets"
)

var (
	rowsInfo        []rowInfo
	mapStrTemplates = map[string]string{
		metricTagNodes: `label_replace(
			label_replace(
				topk(10,sum(
					rate($NAME[1m])
				) by (SrcK8S_HostName, DstK8S_HostName)),
				"SrcK8S_HostName", "(external)", "SrcK8S_HostName", "()"
			),
			"DstK8S_HostName", "(external)", "DstK8S_HostName", "()"
		)`,
		metricTagNamespaces: `label_replace(
			label_replace(
				topk(10,sum(
					rate($NAME{$FILTERS1}[1m]) or rate($NAME{$FILTERS2}[1m])
				) by (SrcK8S_Namespace, DstK8S_Namespace)),
				"SrcK8S_Namespace", "(not namespaced)", "SrcK8S_Namespace", "()"
			),
			"DstK8S_Namespace", "(not namespaced)", "DstK8S_Namespace", "()"
		)`,
		metricTagWorkloads: `label_replace(
			label_replace(
				topk(10,sum(
					rate($NAME{$FILTERS1}[1m]) or rate($NAME{$FILTERS2}[1m])
				) by (SrcK8S_Namespace, SrcK8S_OwnerName, DstK8S_Namespace, DstK8S_OwnerName)),
				"SrcK8S_Namespace", "non pods", "SrcK8S_Namespace", "()"
			),
			"DstK8S_Namespace", "non pods", "DstK8S_Namespace", "()"
		)`,
	}
	mapLegends = map[string]string{
		metricTagNodes:      "{{SrcK8S_HostName}} -> {{DstK8S_HostName}}",
		metricTagNamespaces: "{{SrcK8S_Namespace}} -> {{DstK8S_Namespace}}",
		metricTagWorkloads:  "{{SrcK8S_OwnerName}} ({{SrcK8S_Namespace}}) -> {{DstK8S_OwnerName}} ({{DstK8S_Namespace}})",
	}
	formatCleaner = strings.NewReplacer(
		"\"", "\\\"",
		"\t", "",
		"\n", "",
	)
	appFilterReplacer = strings.NewReplacer(
		"$FILTERS1", appsFilters1,
		"$FILTERS2", appsFilters2,
	)
	infraFilterReplacer = strings.NewReplacer(
		"$FILTERS1", infraFilters1,
		"$FILTERS2", infraFilters2,
	)
)

func init() {
	for _, group := range []string{metricTagNodes, metricTagNamespaces, metricTagWorkloads} {
		groupTrimmed := strings.TrimSuffix(group, "s")
		// byte/pkt rates
		for _, vt := range []string{metricTagBytes, metricTagPackets} {
			for _, dir := range []string{metricTagEgress, metricTagIngress} {
				rowsInfo = append(rowsInfo, rowInfo{
					metric:    fmt.Sprintf("netobserv_%s_%s_%s_total", groupTrimmed, dir, vt),
					group:     group,
					dir:       dir,
					valueType: vt,
				})
			}
		}
		// drops
		for _, vt := range []string{metricTagDropBytes, metricTagDropPackets} {
			rowsInfo = append(rowsInfo, rowInfo{
				metric:    fmt.Sprintf("netobserv_%s_%s_total", groupTrimmed, vt),
				group:     group,
				valueType: vt,
			})
		}
		// TODO: RTT dashboard (after dashboard refactoring for exposed metrics; need to handle histogram queries)
	}
}

func buildQuery(netobsNs string, rowInfo rowInfo, isApp bool) string {
	strTemplate := mapStrTemplates[rowInfo.group]
	q := strings.ReplaceAll(strTemplate, "$NAME", rowInfo.metric)
	if isApp {
		q = appFilterReplacer.Replace(q)
	} else {
		q = infraFilterReplacer.Replace(q)
	}
	q = strings.ReplaceAll(q, "$NETOBSERV_NS", netobsNs)
	// Return formatted / one line
	return formatCleaner.Replace(q)
}

func flowMetricsPanel(netobsNs string, rowInfo rowInfo, layer string) string {
	q := buildQuery(netobsNs, rowInfo, layer == layerApps)
	legend := mapLegends[rowInfo.group]
	return fmt.Sprintf(`
	{
		"aliasColors": {},
		"bars": false,
		"dashLength": 10,
		"dashes": false,
		"datasource": "prometheus",
		"fill": 1,
		"fillGradient": 0,
		"gridPos": {
			"h": 20,
			"w": 25,
			"x": 0,
			"y": 0
		},
		"id": 2,
		"legend": {
			"alignAsTable": false,
			"avg": false,
			"current": false,
			"max": false,
			"min": false,
			"rightSide": false,
			"show": true,
			"sideWidth": null,
			"total": false,
			"values": false
		},
		"lines": true,
		"linewidth": 1,
		"links": [],
		"nullPointMode": "null",
		"percentage": false,
		"pointradius": 5,
		"points": false,
		"renderer": "flot",
		"repeat": null,
		"seriesOverrides": [],
		"spaceLength": 10,
		"span": 6,
		"stack": false,
		"steppedLine": false,
		"targets": [
			{
				"expr": "%s",
				"format": "time_series",
				"intervalFactor": 2,
				"legendFormat": "%s",
				"refId": "A"
			}
		],
		"thresholds": [],
		"timeFrom": null,
		"timeShift": null,
		"title": "%s",
		"tooltip": {
			"shared": true,
			"sort": 0,
			"value_type": "individual"
		},
		"type": "graph",
		"xaxis": {
			"buckets": null,
			"mode": "time",
			"name": null,
			"show": true,
			"values": []
		},
		"yaxes": [
			{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
			},
			{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
			}
		]
	}
	`, q, legend, layer)
}

func flowMetricsRow(netobsNs string, rowInfo rowInfo) string {
	var verb, vt string
	switch rowInfo.dir {
	case metricTagEgress:
		verb = "sent"
	case metricTagIngress:
		verb = "received"
	}
	switch rowInfo.valueType {
	case metricTagBytes:
		vt = "byte"
	case metricTagPackets:
		vt = "packet"
	case metricTagDropBytes:
		vt = "drop bytes"
	case metricTagDropPackets:
		vt = "drop packets"
	}
	title := fmt.Sprintf("Top %s rates %s per source and destination %s", vt, verb, rowInfo.group)
	var panels string
	if rowInfo.group == metricTagNodes {
		panels = fmt.Sprintf("[%s]", flowMetricsPanel(netobsNs, rowInfo, ""))
	} else {
		panels = fmt.Sprintf("[%s, %s]", flowMetricsPanel(netobsNs, rowInfo, layerApps), flowMetricsPanel(netobsNs, rowInfo, layerInfra))
	}
	return fmt.Sprintf(`
	{
		"collapse": false,
		"editable": true,
		"height": "250px",
		"panels": %s,
		"showTitle": true,
		"title": "%s"
	}
	`, panels, title)
}

func CreateFlowMetricsDashboard(netobsNs string, metrics []string) (string, error) {
	var rows []string

	for _, ri := range rowsInfo {
		trimmed := strings.TrimPrefix(ri.metric, "netobserv_")
		if slices.Contains(metrics, trimmed) {
			rows = append(rows, flowMetricsRow(netobsNs, ri))
		} else if strings.Contains(ri.metric, "_namespace_") {
			// namespace-based panels can also be displayed using workload-based metrics
			// Try again, replacing *_namespace_* with *_workload_*
			ri.metric = strings.Replace(ri.metric, "_namespace_", "_workload_", 1)
			trimmed = strings.TrimPrefix(ri.metric, "netobserv_")
			if slices.Contains(metrics, trimmed) {
				rows = append(rows, flowMetricsRow(netobsNs, ri))
			}
		}
	}

	// return empty if dashboard doesn't contains rows
	if len(rows) == 0 {
		return "", nil
	}
	rowsStr := strings.Join(rows, ",")
	return fmt.Sprintf(`
	{
		"__inputs": [],
		"__requires": [],
		"annotations": {
			"list": []
		},
		"editable": false,
		"gnetId": null,
		"graphTooltip": 0,
		"hideControls": false,
		"id": null,
		"links": [],
		"rows": [%s],
		"refresh": "",
		"schemaVersion": 16,
		"style": "dark",
		"tags": [
			"netobserv-mixin"
		],
		"templating": {
			"list": []
		},
		"time": {
			"from": "now",
			"to": "now"
		},
		"timepicker": {
			"refresh_intervals": [
				"5s",
				"10s",
				"30s",
				"1m",
				"5m",
				"15m",
				"30m",
				"1h",
				"2h",
				"1d"
			],
			"time_options": [
				"5m",
				"15m",
				"1h",
				"6h",
				"12h",
				"24h",
				"2d",
				"7d",
				"30d"
			]
		},
		"timezone": "browser",
		"title": "NetObserv",
		"version": 0
	}
	`, rowsStr), nil
}
