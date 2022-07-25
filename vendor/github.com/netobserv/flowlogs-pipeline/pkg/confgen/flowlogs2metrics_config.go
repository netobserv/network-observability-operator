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
	config "github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (cg *ConfGen) GenerateFlowlogs2PipelineConfig() config.ConfigFileStruct {
	configStruct := config.ConfigFileStruct{
		LogLevel: "error",
		Pipeline: []config.Stage{
			{Name: "ingest_collector"},
			{Name: "transform_generic",
				Follows: "ingest_collector",
			},
			{Name: "transform_network",
				Follows: "transform_generic",
			},
			{Name: "extract_aggregate",
				Follows: "transform_network",
			},
			{Name: "encode_prom",
				Follows: "extract_aggregate",
			},
			{Name: "write_loki",
				Follows: "transform_network",
			},
		},
		Parameters: []config.StageParam{
			{Name: "ingest_collector",
				Ingest: &config.Ingest{
					Type: "collector",
					Collector: &api.IngestCollector{
						Port:       cg.config.Ingest.Collector.Port,
						PortLegacy: cg.config.Ingest.Collector.PortLegacy,
						HostName:   cg.config.Ingest.Collector.HostName,
					},
				},
			},
			{Name: "transform_generic",
				Transform: &config.Transform{
					Type: "generic",
					Generic: &api.TransformGeneric{
						Policy: "replace_keys",
						Rules:  cg.config.Transform.Generic.Rules,
					},
				},
			},
			{Name: "transform_network",
				Transform: &config.Transform{
					Type: "network",
					Network: &api.TransformNetwork{
						Rules: cg.transformRules,
					},
				},
			},
			{Name: "extract_aggregate",
				Extract: &config.Extract{
					Type:       "aggregates",
					Aggregates: cg.aggregateDefinitions,
				},
			},
			{Name: "encode_prom",
				Encode: &config.Encode{
					Type: "prom",
					Prom: &api.PromEncode{
						Port:    cg.config.Encode.Prom.Port,
						Prefix:  cg.config.Encode.Prom.Prefix,
						Metrics: cg.promMetrics,
					},
				},
			},
			{Name: "write_loki",
				Write: &config.Write{
					Type: cg.config.Write.Type,
					Loki: &cg.config.Write.Loki,
				},
			},
		},
	}
	return configStruct
}

func (cg *ConfGen) GenerateTruncatedConfig(stages []string) config.ConfigFileStruct {
	parameters := make([]config.StageParam, len(stages))
	for i, stage := range stages {
		switch stage {
		case "ingest":
			parameters[i] = config.StageParam{
				Name: "ingest_collector",
				Ingest: &config.Ingest{
					Type: "collector",
					Collector: &api.IngestCollector{
						Port:       cg.config.Ingest.Collector.Port,
						PortLegacy: cg.config.Ingest.Collector.PortLegacy,
						HostName:   cg.config.Ingest.Collector.HostName,
					},
				},
			}
		case "transform_generic":
			parameters[i] = config.StageParam{
				Name: "transform_generic",
				Transform: &config.Transform{
					Type: "generic",
					Generic: &api.TransformGeneric{
						Policy: "replace_keys",
						Rules:  cg.config.Transform.Generic.Rules,
					},
				},
			}
		case "transform_network":
			parameters[i] = config.StageParam{
				Name: "transform_network",
				Transform: &config.Transform{
					Type: "network",
					Network: &api.TransformNetwork{
						Rules: cg.transformRules,
					},
				},
			}
		case "extract_aggregate":
			parameters[i] = config.StageParam{
				Name: "extract_aggregate",
				Extract: &config.Extract{
					Type:       "aggregates",
					Aggregates: cg.aggregateDefinitions,
				},
			}
		case "encode_prom":
			parameters[i] = config.StageParam{
				Name: "encode_prom",
				Encode: &config.Encode{
					Type: "prom",
					Prom: &api.PromEncode{
						Port:    cg.config.Encode.Prom.Port,
						Prefix:  cg.config.Encode.Prom.Prefix,
						Metrics: cg.promMetrics,
					},
				},
			}
		case "write_loki":
			parameters[i] = config.StageParam{
				Name: "write_loki",
				Write: &config.Write{
					Type: cg.config.Write.Type,
					Loki: &cg.config.Write.Loki,
				},
			}
		}
	}
	log.Debugf("parameters = %v \n", parameters)
	configStruct := config.ConfigFileStruct{
		Parameters: parameters,
	}
	return configStruct
}

func (cg *ConfGen) writeConfigFile(fileName string, config config.ConfigFileStruct) error {
	configData, err := yaml.Marshal(&config)
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
