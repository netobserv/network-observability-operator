/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package confgen

import (
	"bytes"
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const TypeGrafana = "grafana"
const panelTargetTypeGraphPanel = "graphPanel"
const singleStatTypeGraphPanel = "singleStat"
const barGaugeTypeGraphPanel = "barGauge"
const heatmapTypeGraphPanel = "heatmap"
const panelTargetTypeLokiGraphPanel = "lokiGraphPanel"

const jsonNetHeaderTemplate = `
local grafana = import 'grafana.libsonnet';
local dashboard = grafana.dashboard;
local row = grafana.row;
local singlestat = grafana.singlestat;
local graphPanel = grafana.graphPanel;
local heatmapPanel = grafana.heatmapPanel;
local barGaugePanel = grafana.barGaugePanel;
local table = grafana.table;
local prometheus = grafana.prometheus;
local template = grafana.template;
`

const dashboardsTemplate = `
dashboard.new(
  schemaVersion={{.SchemaVersion}},
  title="{{.Title}}",
  time_from="{{.TimeFrom}}",
  tags={{.Tags}},
)`

const graphPanelTemplate = `
.addPanel(
  graphPanel.new(
    datasource='prometheus',
    title="{{.Title}}",
  )
  .addTarget(
    prometheus.target(
      expr='{{.Expr}}',
    )
  ), gridPos={
    x: 0,
    y: 0,
    w: 25,
    h: 20,
  }
)`

const singleStatTemplate = `
.addPanel(
  singlestat.new(
    datasource='prometheus',
    title="{{.Title}}",
  )
  .addTarget(
    prometheus.target(
      expr='{{.Expr}}',
    )
  ), gridPos={
    x: 0,
    y: 0,
    w: 5,
    h: 5,
  }
)`

// "{{`{{le}}`}}" is a trick to write "{{le}}" in a template without getting it to be parsed.
const barGaugeTemplate = `
.addPanel(
  barGaugePanel.new(
    datasource='prometheus',
    title="{{.Title}}",
    thresholds=[
          {
            "color": "green",
            "value": null
          }
        ],
  )
  .addTarget(
    prometheus.target(
      expr='{{.Expr}}',
      format='heatmap',
      legendFormat='` + "{{`{{le}}`}}" + `',
    )
  ), gridPos={
    x: 0,
    y: 0,
    w: 12,
    h: 8,
  }
)`

// "{{`{{le}}`}}" is a trick to write "{{le}}" in a template without getting it to be parsed.
const heatmapTemplate = `
.addPanel(
  heatmapPanel.new(
    datasource='prometheus',
    title="{{.Title}}",
    dataFormat="tsbuckets",
  )
  .addTarget(
    prometheus.target(
      expr='{{.Expr}}',
      format='heatmap',
      legendFormat='` + "{{`{{le}}`}}" + `',
    )
  ), gridPos={
    x: 0,
    y: 0,
    w: 25,
    h: 8,
  }
)`

const lokiGraphPanelTemplate = `
.addPanel(
  graphPanel.new(
    datasource='loki',
    title="{{.Title}}",
  )
  .addTarget(
    prometheus.target(
      expr='{{.Expr}}',
    )
  ), gridPos={
    x: 0,
    y: 0,
    w: 25,
    h: 20,
  }
)`

type Dashboards map[string]Dashboard

type Dashboard struct {
	Name   string
	Header []byte
	Panels []byte
}

func (cg *ConfGen) generateGrafanaJsonnet(folderName string) error {

	// generate dashboards
	dashboards, err := cg.generateGrafanaJsonnetDashboards()
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnetDashboards err: %v ", err)
		return err
	}

	// add all panels
	dashboards, err = cg.addPanelsToDashboards(dashboards)
	if err != nil {
		log.Debugf("cg.addPanelsToDashboards err: %v ", err)
		return err
	}

	// write to destination files
	for _, dashboard := range dashboards {
		output := []byte(jsonNetHeaderTemplate)
		output = append(output, dashboard.Header...)
		output = append(output, dashboard.Panels...)

		fileName := folderName + "dashboard_" + dashboard.Name + ".jsonnet"
		err = os.WriteFile(fileName, output, 0644)
		if err != nil {
			log.Debugf("os.WriteFile to file %s err: %v ", fileName, err)
			return err
		}
	}

	return nil
}

func (cg *ConfGen) generateGrafanaJsonnetDashboards() (Dashboards, error) {
	dashboards := Dashboards{}
	dashboardTemplate := template.Must(template.New("dashboardTemplate").Parse(dashboardsTemplate))
	for _, dashboard := range cg.config.Visualization.Grafana.Dashboards {
		newDashboard := new(bytes.Buffer)
		err := dashboardTemplate.Execute(newDashboard, dashboard)
		if err != nil {
			log.Infof("dashboardTemplate.Execute for %s err: %v ", dashboard.Title, err)
			continue
		}

		dashboards[dashboard.Name] = Dashboard{
			Name:   dashboard.Name,
			Header: newDashboard.Bytes(),
			Panels: []byte{},
		}
	}

	return dashboards, nil
}

func (cg *ConfGen) addPanelsToDashboards(dashboards Dashboards) (Dashboards, error) {

	graphPanelTemplate := template.Must(template.New("graphPanelTemplate").Parse(graphPanelTemplate))
	singleStatTemplate := template.Must(template.New("graphPanelTemplate").Parse(singleStatTemplate))
	barGaugeTemplate := template.Must(template.New("graphPanelTemplate").Parse(barGaugeTemplate))
	heatmapTemplate := template.Must(template.New("graphPanelTemplate").Parse(heatmapTemplate))

	lokiGraphPanelTemplate := template.Must(template.New("graphPanelTemplate").Parse(lokiGraphPanelTemplate))

	for _, definition := range cg.visualizations {
		if definition.Type != TypeGrafana {
			log.Infof("skipping definition of type %s", definition.Type)
			continue
		}

		for _, panelTarget := range definition.Grafana {
			newPanel := new(bytes.Buffer)
			switch panelTarget.Type {
			case panelTargetTypeGraphPanel:
				err := graphPanelTemplate.Execute(newPanel, panelTarget)
				if err != nil {
					log.Infof("addPanelAddTargetTemplate.Execute for %s err: %v ", panelTarget.Title, err)
					continue
				}
			case singleStatTypeGraphPanel:
				err := singleStatTemplate.Execute(newPanel, panelTarget)
				if err != nil {
					log.Infof("singleStatTemplate.Execute for %s err: %v ", panelTarget.Title, err)
					continue
				}
			case barGaugeTypeGraphPanel:
				err := barGaugeTemplate.Execute(newPanel, panelTarget)
				if err != nil {
					log.Infof("barGaugeTemplate.Execute for %s err: %v ", panelTarget.Title, err)
					continue
				}
			case heatmapTypeGraphPanel:
				err := heatmapTemplate.Execute(newPanel, panelTarget)
				if err != nil {
					log.Infof("heatmapTemplate.Execute for %s err: %v ", panelTarget.Title, err)
					continue
				}
			case panelTargetTypeLokiGraphPanel:
				err := lokiGraphPanelTemplate.Execute(newPanel, panelTarget)
				if err != nil {
					log.Infof("addPanelAddTargetTemplate.Execute for %s err: %v ", panelTarget.Title, err)
					continue
				}
			default:
				log.Infof("unsuported panelTarget.Type %s ", panelTarget.Type)
				continue
			}

			dashboard, ok := dashboards[panelTarget.Dashboard]
			if ok {
				dashboards[panelTarget.Dashboard] = Dashboard{
					Name:   dashboard.Name,
					Header: dashboard.Header,
					Panels: append(dashboard.Panels, newPanel.Bytes()...),
				}
			} else {
				log.Infof("can't find dashboard %s, skipping adding panel %s", panelTarget.Dashboard, panelTarget.Title)
			}
		}
	}

	return dashboards, nil
}
