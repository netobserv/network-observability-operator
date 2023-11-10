package dashboards

import (
	"fmt"
	"strings"
)

type Dashboard struct {
	Rows []*Row
}

type Row struct {
	Title  string
	Metric string
	Panels []Panel
}

type Panel struct {
	Title   string
	Targets []Target
}

type Target struct {
	Expr   string
	Legend string
}

var formatCleaner = strings.NewReplacer(
	"\"", "\\\"",
	"\t", "",
	"\n", "",
)

func (d *Dashboard) ToGrafanaJSON(netobsNs string) string {
	// return empty if dashboard doesn't contains rows
	if len(d.Rows) == 0 {
		return ""
	}

	var rows []string
	for _, ri := range d.Rows {
		rows = append(rows, ri.ToGrafanaJSON(netobsNs))
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
	`, rowsStr)
}

func (r *Row) ToGrafanaJSON(netobsNs string) string {
	var panels []string
	for _, panel := range r.Panels {
		panels = append(panels, panel.ToGrafanaJSON(netobsNs))
	}
	return fmt.Sprintf(`
	{
		"collapse": false,
		"editable": true,
		"height": "250px",
		"panels": [%s],
		"showTitle": true,
		"title": "%s"
	}
	`, strings.Join(panels, ","), r.Title)
}

func (r *Row) replaceMetric(newName string) *Row {
	clone := Row{
		Title: r.Title,
	}
	for _, p := range r.Panels {
		clone.Panels = append(clone.Panels, p.replaceMetric(r.Metric, newName))
	}
	return &clone
}

func (p *Panel) ToGrafanaJSON(netobsNs string) string {
	var targets []string
	for _, target := range p.Targets {
		targets = append(targets, target.ToGrafanaJSON(netobsNs))
	}
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
		"targets": [%s],
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
	`, strings.Join(targets, ","), p.Title)
}

func (p *Panel) replaceMetric(oldName, newName string) Panel {
	clone := Panel{
		Title: p.Title,
	}
	for _, t := range p.Targets {
		clone.Targets = append(clone.Targets, Target{
			Legend: t.Legend,
			Expr:   strings.ReplaceAll(t.Expr, oldName, newName),
		})
	}
	return clone
}

func (t *Target) ToGrafanaJSON(netobsNs string) string {
	expr := formatCleaner.Replace(strings.ReplaceAll(t.Expr, "$NETOBSERV_NS", netobsNs))
	return fmt.Sprintf(`
	{
		"expr": "%s",
		"format": "time_series",
		"intervalFactor": 2,
		"legendFormat": "%s",
		"refId": "A"
	}
	`, expr, t.Legend)
}
