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

const (
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
	opts                 *Options
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
	cg.config, err = cg.ParseConfigFile(cg.opts.SrcFolder + "/" + configFileName)
	if err != nil {
		log.Debugf("cg.ParseConfigFile err: %v ", err)
		return err
	}

	definitionFiles := getDefinitionFiles(cg.opts.SrcFolder)
	for _, definitionFile := range definitionFiles {
		b, err := ioutil.ReadFile(definitionFile)
		if err != nil {
			log.Debugf("ioutil.ReadFile err: %v ", err)
			continue
		}
		err = cg.ParseDefinition(definitionFile, b)
		if err != nil {
			log.Debugf("cg.parseDefinition err: %v ", err)
			continue
		}
	}

	cg.dedupe()

	if len(cg.opts.GenerateStages) != 0 {
		cfg := cg.GenerateTruncatedConfig()
		err = cg.writeConfigFile(cg.opts.DestConfFile, cfg)
		if err != nil {
			log.Debugf("cg.GenerateTruncatedConfig err: %v ", err)
			return err
		}
		return nil
	} else {
		config := cg.GenerateFlowlogs2PipelineConfig()
		err = cg.writeConfigFile(cg.opts.DestConfFile, config)
		if err != nil {
			log.Debugf("cg.GenerateFlowlogs2PipelineConfig err: %v ", err)
			return err
		}
	}

	err = cg.generateDoc(cg.opts.DestDocFile)
	if err != nil {
		log.Debugf("cg.generateDoc err: %v ", err)
		return err
	}

	err = cg.generateGrafanaJsonnet(cg.opts.DestGrafanaJsonnetFolder)
	if err != nil {
		log.Debugf("cg.generateGrafanaJsonnet err: %v ", err)
		return err
	}

	return nil
}

func checkHeader(bytes []byte) error {
	header := make([]byte, len(definitionHeader))
	copy(header, bytes)
	if string(header) != definitionHeader {
		return fmt.Errorf("wrong header")
	}
	return nil
}

func (cg *ConfGen) ParseDefinition(name string, bytes []byte) error {
	// check header
	err := checkHeader(bytes)
	if err != nil {
		log.Debugf("%s cg.checkHeader err: %v ", name, err)
		return err
	}

	// parse yaml
	var defFile DefFile
	err = yaml.UnmarshalStrict(bytes, &defFile)
	if err != nil {
		log.Debugf("%s yaml.UnmarshalStrict err: %v ", name, err)
		return err
	}

	//skip if their skip tag match
	for _, skipTag := range cg.opts.SkipWithTags {
		for _, tag := range defFile.Tags {
			if skipTag == tag {
				log.Infof("skipping definition %s due to skip tag %s", name, tag)
				return nil
			}
		}
	}

	// parse definition
	definition := Definition{
		FileName:    name,
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

func getDefinitionFiles(rootPath string) []string {

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

func NewConfGen(opts *Options) *ConfGen {
	return &ConfGen{
		opts:                 opts,
		transformRules:       api.NetworkTransformRules{},
		aggregateDefinitions: aggregate.Definitions{},
		definitions:          Definitions{},
		visualizations:       Visualizations{},
	}
}
