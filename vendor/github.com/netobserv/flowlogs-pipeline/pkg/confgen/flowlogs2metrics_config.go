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
	"os"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (cg *ConfGen) GenerateFlowlogs2PipelineConfig() *config.ConfigFileStruct {
	pipeline, _ := config.NewPipeline("ingest_collector", &cg.config.Ingest)
	forkedNode := pipeline
	if cg.config.Transform.Generic != nil {
		gen := *cg.config.Transform.Generic
		if len(gen.Policy) == 0 {
			gen.Policy = "replace_keys"
		}
		forkedNode = forkedNode.TransformGeneric("transform_generic", gen)
	}
	if cg.config.Extract.ConnTrack != nil {
		forkedNode = forkedNode.ConnTrack("extract_conntrack", *cg.config.Extract.ConnTrack)
	}
	if len(cg.transformRules) > 0 {
		forkedNode = forkedNode.TransformNetwork("transform_network", api.TransformNetwork{
			Rules: cg.transformRules,
		})
	}
	metricsNode := forkedNode
	if len(cg.aggregateDefinitions) > 0 {
		metricsNode = metricsNode.Aggregate("extract_aggregate", cg.aggregateDefinitions)
		if len(cg.timebasedTopKs.Rules) > 0 {
			metricsNode = metricsNode.ExtractTimebased("extract_timebased", api.ExtractTimebased{
				Rules: cg.timebasedTopKs.Rules,
			})
		}
	}
	if len(cg.promMetrics) > 0 {
		metricsNode.EncodePrometheus("encode_prom", api.PromEncode{
			Address: cg.config.Encode.Prom.Address,
			Port:    cg.config.Encode.Prom.Port,
			Prefix:  cg.config.Encode.Prom.Prefix,
			Metrics: cg.promMetrics,
		})
	}
	if cg.config.Write.Loki != nil {
		forkedNode.WriteLoki("write_loki", *cg.config.Write.Loki)
	}
	return &config.ConfigFileStruct{
		LogLevel:   "error",
		Pipeline:   pipeline.GetStages(),
		Parameters: pipeline.GetStageParams(),
	}
}

func (cg *ConfGen) GenerateTruncatedConfig() []config.StageParam {
	parameters := make([]config.StageParam, len(cg.opts.GenerateStages))
	for i, stage := range cg.opts.GenerateStages {
		switch stage {
		case "ingest":
			parameters[i] = config.NewCollectorParams("ingest_collector", *cg.config.Ingest.Collector)
		case "transform_generic":
			parameters[i] = config.NewTransformGenericParams("transform_generic", *cg.config.Transform.Generic)
		case "transform_network":
			parameters[i] = config.NewTransformNetworkParams("transform_network", *cg.config.Transform.Network)
		case "extract_aggregate":
			parameters[i] = config.NewAggregateParams("extract_aggregate", cg.aggregateDefinitions)
		case "extract_timebased":
			parameters[i] = config.NewTimbasedParams("extract_timebased", cg.timebasedTopKs)
		case "encode_prom":
			parameters[i] = config.NewEncodePrometheusParams("encode_prom", api.PromEncode{
				Metrics: cg.promMetrics,
			})
		case "write_loki":
			parameters[i] = config.NewWriteLokiParams("write_loki", *cg.config.Write.Loki)
		}
	}
	log.Debugf("parameters = %v \n", parameters)
	return parameters
}

func (cg *ConfGen) writeConfigFile(fileName string, cfg interface{}) error {
	configData, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	header := "# This file was generated automatically by flowlogs-pipeline confgenerator"
	data := fmt.Sprintf("%s\n%s\n", header, configData)
	err = os.WriteFile(fileName, []byte(data), 0664)
	if err != nil {
		return err
	}

	return nil
}
