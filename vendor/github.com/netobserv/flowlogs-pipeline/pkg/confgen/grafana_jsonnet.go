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
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/google/go-jsonnet"
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
      legendFormat='{{.Legend}}',
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
      legendFormat='{{.Legend}}',
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
      legendFormat='{{.Legend}}',
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
      legendFormat='{{.Legend}}',
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
      legendFormat='{{.Legend}}',
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

func (dashboard *Dashboard) generateDashboardJsonnet() []byte {
	output := []byte(jsonNetHeaderTemplate)
	output = append(output, dashboard.Header...)
	output = append(output, dashboard.Panels...)
	return output
}

func (cg *ConfGen) generateGrafanaDashboards() (Dashboards, error) {
	// generate dashboards
	dashboards, err := cg.generateGrafanaJsonnetDashboards()
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnetDashboards err: %v ", err)
		return nil, err
	}

	// add all panels
	dashboards, err = cg.addPanelsToDashboards(dashboards)
	if err != nil {
		log.Debugf("cg.addPanelsToDashboards err: %v ", err)
		return nil, err
	}
	return dashboards, nil
}

func (cg *ConfGen) generateGrafanaJsonnetFiles(folderName string, dashboards Dashboards) error {
	err := os.MkdirAll(folderName, 0755)
	if err != nil {
		log.Debugf("os.MkdirAll err: %v ", err)
		return err
	}
	// write to destination files
	for _, dashboard := range dashboards {
		output := dashboard.generateDashboardJsonnet()

		fileName := filepath.Join(folderName, "dashboard_"+dashboard.Name+".jsonnet")
		err = os.WriteFile(fileName, output, 0644)
		if err != nil {
			log.Debugf("os.WriteFile to file %s err: %v ", fileName, err)
			return err
		}
	}

	return nil
}

func (cg *ConfGen) generateJsonFiles(folderName string, dashboards Dashboards) error {
	err := os.MkdirAll(folderName, 0755)
	if err != nil {
		log.Debugf("os.MkdirAll err: %v ", err)
		return err
	}
	// write to destination files
	for _, dashboard := range dashboards {
		jsonStr, err := cg.generateGrafanaJsonStr(dashboard)
		if err != nil {
			return err
		}
		fileName := filepath.Join(folderName, "dashboard_"+dashboard.Name+".json")
		err = os.WriteFile(fileName, []byte(jsonStr), 0644)
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
	if cg.config == nil {
		return nil, fmt.Errorf("config missing")
	}
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

const grafanaDirPath = "grafana/grafonnet-lib/grafonnet"

//go:embed grafana/grafonnet-lib/grafonnet
var grafanaDir embed.FS

type embedImporter struct {
	fsBase  embed.FS
	fsCache map[string]*fsCacheEntry
}

type fsCacheEntry struct {
	contents *[]byte
	exists   bool
}

func (cg *ConfGen) GenerateGrafanaJson() (string, error) {
	log.Debugf("grafanaDir = %v", grafanaDir)
	dashboards, err := cg.generateGrafanaDashboards()
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnetDashboards err: %v ", err)
		return "", err
	}
	panelsJson := ""
	for _, dashboard := range dashboards {
		jsonStr, err := cg.generateGrafanaJsonStr(dashboard)
		if err != nil {
			return "", err
		}
		panelsJson = panelsJson + jsonStr
	}
	return panelsJson, nil
}

func (cg *ConfGen) generateGrafanaJsonStr(dashboard Dashboard) (string, error) {
	vm := jsonnet.MakeVM()
	importer := &embedImporter{fsBase: grafanaDir}
	err := importer.initializeCache()
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnetDashboards err: %v ", err)
		return "", err
	}
	vm.Importer(importer)
	output := dashboard.generateDashboardJsonnet()
	jsonStr, err := vm.EvaluateAnonymousSnippet("/dev/null", string(output))
	if err != nil {
		log.Errorf("EvaluateFile failure, err = %v \n", err)
	}
	return jsonStr, nil
}

func (importer *embedImporter) initializeCache() error {
	importer.fsCache = make(map[string]*fsCacheEntry)
	entries, err := importer.fsBase.ReadDir(grafanaDirPath)
	if err != nil {
		return fmt.Errorf("failed to access grafana directory: %w", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		cacheEntry := &fsCacheEntry{
			exists: false,
		}
		importer.fsCache[fileName] = cacheEntry
	}
	return nil
}

// Import is the function required by the Importer interface to find source files
func (importer *embedImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// ignore the importedFrom parameter

	// search for item in cache
	entry, ok := importer.fsCache[importedPath]
	if !ok {
		contents := jsonnet.MakeContentsRaw([]byte{})
		return contents, importedPath, fmt.Errorf("grafana file not found: %s", importedPath)
	}
	if !entry.exists {
		// read in the data
		filePath := filepath.Join(grafanaDirPath, importedPath)
		fileData, err := grafanaDir.ReadFile(filePath)
		if err != nil {
			contents := jsonnet.MakeContentsRaw([]byte{})
			return contents, importedPath, fmt.Errorf("error reading grafana file: %s", importedPath)
		}
		entry.exists = true
		entry.contents = &fileData
	}
	contents := jsonnet.MakeContentsRaw(*entry.contents)
	return contents, importedPath, nil
}
