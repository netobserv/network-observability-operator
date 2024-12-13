/*
 * Copyright (C) 2019 IBM, Inc.
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

package pipeline

import (
	"fmt"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/ingest"
	"github.com/netobserv/gopipes/pkg/node"
	log "github.com/sirupsen/logrus"
)

// interface definitions of pipeline components
const (
	StageIngest    = "ingest"
	StageTransform = "transform"
	StageExtract   = "extract"
	StageEncode    = "encode"
	StageWrite     = "write"
)

// Pipeline manager
type Pipeline struct {
	startNodes       []*node.Start[config.GenericMap]
	terminalNodes    []*node.Terminal[config.GenericMap]
	pipelineEntryMap map[string]*pipelineEntry
	IsRunning        bool
	// TODO: this field is only used for test verification. We should rewrite the build process
	// to be able to remove it from here
	pipelineStages []*pipelineEntry
	Metrics        *operational.Metrics
	configWatcher  *pipelineConfigWatcher
}

// NewPipeline defines the pipeline elements
func NewPipeline(cfg *config.ConfigFileStruct) (*Pipeline, error) {
	return newPipelineFromIngester(cfg, nil)
}

// newPipelineFromIngester defines the pipeline elements from a preset ingester (e.g. for in-process receiver)
func newPipelineFromIngester(cfg *config.ConfigFileStruct, ing ingest.Ingester) (*Pipeline, error) {
	log.Debugf("entering newPipelineFromIngester")

	log.Debugf("stages = %v ", cfg.Pipeline)
	log.Debugf("configParams = %v ", cfg.Parameters)

	builder := newBuilder(cfg)
	if ing != nil {
		builder.presetIngester(ing)
	}
	if err := builder.readStages(); err != nil {
		return nil, err
	}
	pipeline, err := builder.build()
	if err != nil {
		return nil, err
	}
	pipeline.configWatcher, err = newPipelineConfigWatcher(cfg, pipeline.pipelineEntryMap)
	return pipeline, err
}

func (p *Pipeline) Run() {
	// starting the graph
	for _, s := range p.startNodes {
		s.Start()
	}
	p.IsRunning = true

	if p.configWatcher != nil {
		go p.configWatcher.Run()
	}

	// blocking the execution until the graph terminal stages end
	for _, t := range p.terminalNodes {
		<-t.Done()
	}
	p.IsRunning = false
}

func (p *Pipeline) IsReady() error {
	if !p.IsRunning {
		return fmt.Errorf("pipeline is not running")
	}
	return nil
}

func (p *Pipeline) IsAlive() error {
	if !p.IsRunning {
		return fmt.Errorf("pipeline is not running")
	}
	return nil
}
