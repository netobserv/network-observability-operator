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

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (cg *ConfGen) GenerateFlowlogs2PipelineConfig() *config.ConfigFileStruct {
	pipeline, _ := config.NewPipeline("ingest_collector", &cg.config.Ingest)
	next := pipeline
	if cg.config.Transform.Generic != nil {
		gen := *cg.config.Transform.Generic
		if len(gen.Policy) == 0 {
			gen.Policy = "replace_keys"
		}
		next = next.TransformGeneric("transform_generic", gen)
	}
	if len(cg.transformRules) > 0 {
		next = next.TransformNetwork("transform_network", api.TransformNetwork{
			Rules: cg.transformRules,
		})
	}
	if len(cg.aggregateDefinitions) > 0 {
		agg := next.Aggregate("extract_aggregate", cg.aggregateDefinitions)
		agg.EncodePrometheus("encode_prom", api.PromEncode{
			Port:    cg.config.Encode.Prom.Port,
			Prefix:  cg.config.Encode.Prom.Prefix,
			Metrics: cg.promMetrics,
		})
	}
	if cg.config.Write.Loki != nil {
		next.WriteLoki("write_loki", *cg.config.Write.Loki)
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
	err = ioutil.WriteFile(fileName, []byte(data), 0664)
	if err != nil {
		return err
	}

	return nil
}
