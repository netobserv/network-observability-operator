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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract/aggregate"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	definitionExt    = ".yaml"
	definitionHeader = "#flp_confgen"
	configFileName   = "config.yaml"
)

type Definition struct {
	FileName             string
	Description          string
	Details              string
	Usage                string
	Tags                 []string
	TransformNetwork     *api.TransformNetwork
	AggregateDefinitions *aggregate.Definitions
	PromEncode           *api.PromEncode
	Visualization        *Visualization
}

type Definitions []Definition

type ConfGen struct {
	config               *Config
	transformRules       api.NetworkTransformRules
	aggregateDefinitions aggregate.Definitions
	promMetrics          api.PromMetricsItems
	visualizations       Visualizations
	definitions          Definitions
}

type DefFile struct {
	Description   string                 `yaml:"description"`
	Details       string                 `yaml:"details"`
	Usage         string                 `yaml:"usage"`
	Tags          []string               `yaml:"tags"`
	Transform     map[string]interface{} `yaml:"transform"`
	Extract       map[string]interface{} `yaml:"extract"`
	Encode        map[string]interface{} `yaml:"encode"`
	Visualization Visualization          `yaml:"visualization"`
}

func (cg *ConfGen) Run() error {
	var err error
	cg.config, err = cg.ParseConfigFile(Opt.SrcFolder + "/" + configFileName)
	if err != nil {
		log.Debugf("cg.ParseConfigFile err: %v ", err)
		return err
	}

	definitionFiles := cg.GetDefinitionFiles(Opt.SrcFolder)
	for _, definitionFile := range definitionFiles {
		err := cg.parseFile(definitionFile)
		if err != nil {
			log.Debugf("cg.parseFile err: %v ", err)
			continue
		}
	}

	cg.Dedupe()

	if len(Opt.GenerateStages) != 0 {
		config := cg.GenerateTruncatedConfig(Opt.GenerateStages)
		err = cg.writeConfigFile(Opt.DestConfFile, config)
		if err != nil {
			log.Debugf("cg.GenerateTruncatedConfig err: %v ", err)
			return err
		}
		return nil
	} else {
		config := cg.GenerateFlowlogs2PipelineConfig()
		err = cg.writeConfigFile(Opt.DestConfFile, config)
		if err != nil {
			log.Debugf("cg.GenerateFlowlogs2PipelineConfig err: %v ", err)
			return err
		}
	}

	err = cg.generateDoc(Opt.DestDocFile)
	if err != nil {
		log.Debugf("cg.generateDoc err: %v ", err)
		return err
	}

	err = cg.generateGrafanaJsonnet(Opt.DestGrafanaJsonnetFolder)
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnet err: %v ", err)
		return err
	}

	return nil
}

func (cg *ConfGen) checkHeader(fileName string) error {
	// check header
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Debugf("os.OpenFile error: %v ", err)
		return err
	}
	header := make([]byte, len(definitionHeader))
	_, err = f.Read(header)
	if err != nil || string(header) != definitionHeader {
		log.Debugf("Wrong header file: %s ", fileName)
		return fmt.Errorf("wrong header")
	}
	err = f.Close()
	if err != nil {
		log.Debugf("f.Close err: %v ", err)
		return err
	}

	return nil
}

func (cg *ConfGen) parseFile(fileName string) error {

	// check header
	err := cg.checkHeader(fileName)
	if err != nil {
		log.Debugf("cg.checkHeader err: %v ", err)
		return err
	}

	// parse yaml
	var defFile DefFile
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Debugf("ioutil.ReadFile err: %v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &defFile)
	if err != nil {
		log.Debugf("yaml.Unmarshal err: %v ", err)
		return err
	}

	//skip if their skip tag match
	for _, skipTag := range Opt.SkipWithTags {
		for _, tag := range defFile.Tags {
			if skipTag == tag {
				return fmt.Errorf("skipping definition %s due to skip tag %s", fileName, tag)
			}
		}
	}

	// parse definition
	definition := Definition{
		FileName:    fileName,
		Description: defFile.Description,
		Details:     defFile.Details,
		Usage:       defFile.Usage,
		Tags:        defFile.Tags,
	}

	// parse transport
	definition.TransformNetwork, err = cg.parseTransport(&defFile.Transform)
	if err != nil {
		log.Debugf("parseTransport err: %v ", err)
		return err
	}

	// parse extract
	definition.AggregateDefinitions, err = cg.parseExtract(&defFile.Extract)
	if err != nil {
		log.Debugf("parseExtract err: %v ", err)
		return err
	}

	// parse encode
	definition.PromEncode, err = cg.parseEncode(&defFile.Encode)
	if err != nil {
		log.Debugf("parseEncode err: %v ", err)
		return err
	}

	// parse visualization
	definition.Visualization, err = cg.parseVisualization(&defFile.Visualization)
	if err != nil {
		log.Debugf("cg.parseVisualization err: %v ", err)
		return err
	}

	cg.definitions = append(cg.definitions, definition)

	return nil
}

func (*ConfGen) GetDefinitionFiles(rootPath string) []string {

	var files []string

	_ = filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			log.Debugf("filepath.Walk err: %v ", err)
			return nil
		}
		fMode := f.Mode()
		if fMode.IsRegular() && filepath.Ext(f.Name()) == definitionExt && filepath.Base(f.Name()) != configFileName {
			files = append(files, path)
		}

		return nil
	})

	return files
}

func NewConfGen() (*ConfGen, error) {
	return &ConfGen{
		transformRules:       api.NetworkTransformRules{},
		aggregateDefinitions: aggregate.Definitions{},
		definitions:          Definitions{},
		visualizations:       Visualizations{},
	}, nil
}
