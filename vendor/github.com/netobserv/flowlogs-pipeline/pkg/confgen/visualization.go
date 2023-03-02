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
	jsoniter "github.com/json-iterator/go"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

type VisualizationGrafana struct {
	Expr      string `yaml:"expr"`
	Legend    string `yaml:"legendFormat"`
	Type      string `yaml:"type"`
	Title     string `yaml:"title"`
	Dashboard string `yaml:"dashboard"`
}

type Visualization struct {
	Type    string                 `yaml:"type"`
	Grafana []VisualizationGrafana `yaml:"grafana"`
}

type ConfigVisualizationGrafanaDashboard struct {
	Name          string `yaml:"name"`
	Title         string `yaml:"title"`
	TimeFrom      string `yaml:"time_from"`
	Tags          string `yaml:"tags"`
	SchemaVersion string `yaml:"schemaVersion"`
}

type ConfigVisualizationGrafanaDashboards []ConfigVisualizationGrafanaDashboard

type ConfigVisualizationGrafana struct {
	Dashboards ConfigVisualizationGrafanaDashboards `yaml:"dashboards"`
}

type Visualizations []Visualization

func (cg *ConfGen) parseVisualization(visualization *Visualization) (*Visualization, error) {
	var jsoniterJson = jsoniter.ConfigCompatibleWithStandardLibrary
	localVisualization := *visualization
	b, err := jsoniterJson.Marshal(&localVisualization)
	if err != nil {
		log.Debugf("jsoniterJson.Marshal err: %v ", err)
		return nil, err
	}

	var jsonVisualization Visualization
	err = config.JsonUnmarshalStrict(b, &jsonVisualization)
	if err != nil {
		log.Debugf("Unmarshal aggregate.Definitions err: %v ", err)
		return nil, err
	}

	cg.visualizations = append(cg.visualizations, jsonVisualization)
	return &jsonVisualization, nil
}
