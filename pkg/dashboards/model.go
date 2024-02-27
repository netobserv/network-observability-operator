package dashboards

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Dashboard struct {
	Rows  []*Row
	Title string
}

type Row struct {
	Title    string
	Collapse bool
	Height   int
	Metric   string // TODO: remove
	Panels   []Panel
}

func NewRow(title string, collaspe bool, height int, panels []Panel) *Row {
	return &Row{
		Title:    title,
		Collapse: collaspe,
		Height:   height,
		Panels:   panels,
	}
}

type PanelType string
type PanelUnit string

const (
	PanelTypeSingleStat PanelType = "singlestat"
	PanelTypeGraph      PanelType = "graph"
	PanelUnitBytes      PanelUnit = "bytes"
	PanelUnitShort      PanelUnit = "short"
	PanelUnitSeconds    PanelUnit = "seconds"
	PanelUnitBPS        PanelUnit = "Bps"
	PanelUnitPPS        PanelUnit = "pps"
)

type Panel struct {
	Title   string
	Type    PanelType
	Targets []Target
	Span    int
	Stacked bool
	Unit    PanelUnit
}

func NewPanel(title string, t PanelType, unit PanelUnit, span int, stacked bool, targets []Target) Panel {
	return Panel{
		Title:   title,
		Type:    t,
		Unit:    unit,
		Span:    span,
		Stacked: stacked,
		Targets: targets,
	}
}

func NewGraphPanel(title string, unit PanelUnit, span int, stacked bool, targets []Target) Panel {
	return NewPanel(title, PanelTypeGraph, unit, span, stacked, targets)
}

func NewSingleStatPanel(title string, unit PanelUnit, span int, target Target) Panel {
	return NewPanel(title, PanelTypeSingleStat, unit, span, false, []Target{target})
}

type Target struct {
	Expr   string
	Legend string
}

func NewTarget(expr, legend string) Target {
	return Target{
		Expr:   expr,
		Legend: legend,
	}
}

var formatCleaner = strings.NewReplacer(
	"\"", "\\\"",
	"\t", "",
	"\n", "",
)

func FromBytes(b []byte) (*Dashboard, error) {
	var d Dashboard
	err := json.Unmarshal(b, &d)
	return &d, err
}

func (d *Dashboard) Titles() []string {
	var titles []string
	for _, r := range d.Rows {
		titles = append(titles, r.Title)
	}
	return titles
}

func (d *Dashboard) FindRow(titleSubstr string) *Row {
	for _, r := range d.Rows {
		if strings.Contains(r.Title, titleSubstr) {
			return r
		}
	}
	return nil
}

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
			"networking-mixin"
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
		"title": "%s",
		"version": 0
	}
	`, rowsStr, d.Title)
}

func (r *Row) ToGrafanaJSON(netobsNs string) string {
	var panels []string
	for _, panel := range r.Panels {
		panels = append(panels, panel.ToGrafanaJSON(netobsNs))
	}
	showTitle := true
	if r.Title == "" {
		showTitle = false
	}
	return fmt.Sprintf(`
	{
		"collapse": %t,
		"editable": true,
		"height": "%dpx",
		"panels": [%s],
		"showTitle": %t,
		"title": "%s"
	}
	`, r.Collapse, r.Height, strings.Join(panels, ","), showTitle, r.Title)
}

func (r *Row) replaceMetric(newName string) *Row {
	clone := NewRow(r.Title, r.Collapse, r.Height, nil)
	clone.Metric = r.Metric
	for _, p := range r.Panels {
		clone.Panels = append(clone.Panels, p.replaceMetric(r.Metric, newName))
	}
	return clone
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
		"gridPos": {},
		"id": 1,
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
		"span": %d,
		"stack": %t,
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
		"type": "%s",
		"xaxis": {
			"buckets": null,
			"mode": "time",
			"name": null,
			"show": true,
			"values": []
		},
		"yaxes": [
			{
				"format": "%s",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
			}
		]
	}
	`, p.Span, p.Stacked, strings.Join(targets, ","), p.Title, string(p.Type), string(p.Unit))
}

func (p *Panel) replaceMetric(oldName, newName string) Panel {
	clone := NewPanel(p.Title, p.Type, p.Unit, p.Span, p.Stacked, nil)
	for _, t := range p.Targets {
		clone.Targets = append(
			clone.Targets,
			NewTarget(strings.ReplaceAll(t.Expr, oldName, newName), t.Legend),
		)
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
