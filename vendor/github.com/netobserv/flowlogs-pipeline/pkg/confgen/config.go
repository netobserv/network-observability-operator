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
	"os"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Options struct {
	DestConfFile             string
	DestDocFile              string
	DestGrafanaJsonnetFolder string
	SrcFolder                string
	SkipWithTags             []string
	GenerateStages           []string
	GlobalMetricsPrefix      string
}

type ConfigVisualization struct {
	Grafana ConfigVisualizationGrafana `yaml:"grafana"`
}

type Config struct {
	Description   string              `yaml:"description"`
	Ingest        config.Ingest       `yaml:"ingest"`
	Transform     config.Transform    `yaml:"transform"`
	Extract       config.Extract      `yaml:"extract"`
	Write         config.Write        `yaml:"write"`
	Encode        config.Encode       `yaml:"encode"`
	Visualization ConfigVisualization `yaml:"visualization"`
}

func (cg *ConfGen) ParseConfigFile(fileName string) (*Config, error) {
	// parse config file yaml
	// provide a minimal config for when config file is missing (as for Netobserv Openshift Operator)
	var config Config
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		if len(cg.opts.GenerateStages) == 0 {
			log.Errorf("config file %s does not exist", fileName)
			return nil, err
		}
		return &Config{}, nil
	}
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		log.Debugf("ioutil.ReadFile err: %v ", err)
		return nil, err
	}
	err = yaml.UnmarshalStrict(yamlFile, &config)
	if err != nil {
		log.Debugf("Unmarshal err: %v ", err)
		return nil, err
	}

	return &config, nil
}
