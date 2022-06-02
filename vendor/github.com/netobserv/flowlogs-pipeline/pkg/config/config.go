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

package config

import (
	"encoding/json"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/sirupsen/logrus"
)

type GenericMap map[string]interface{}

var (
	Opt        = Options{}
	PipeLine   []Stage
	Parameters []StageParam
)

type Options struct {
	PipeLine   string
	Parameters string
	Health     Health
}

type Health struct {
	Port string
}

type Stage struct {
	Name    string `json:"name"`
	Follows string `json:"follows,omitempty"`
}

type StageParam struct {
	Name      string     `json:"name"`
	Ingest    *Ingest    `json:"ingest,omitempty"`
	Decode    *Decode    `json:"decode,omitempty"`
	Transform *Transform `json:"transform,omitempty"`
	Extract   *Extract   `json:"extract,omitempty"`
	Encode    *Encode    `json:"encode,omitempty"`
	Write     *Write     `json:"write,omitempty"`
}

type Ingest struct {
	Type      string               `json:"type"`
	File      *File                `json:"file,omitempty"`
	Collector *api.IngestCollector `json:"collector,omitempty"`
	Kafka     *api.IngestKafka     `json:"kafka,omitempty"`
	GRPC      *api.IngestGRPCProto `json:"grpc,omitempty"`
}

type File struct {
	Filename string `json:"filename"`
	Loop     bool   `json:"loop"`
	Chunks   int    `json:"chunks"`
}

type Decode struct {
	Type string         `json:"type"`
	Aws  *api.DecodeAws `json:"aws,omitempty"`
}

type Transform struct {
	Type    string                `json:"type"`
	Generic *api.TransformGeneric `json:"generic,omitempty"`
	Filter  *api.TransformFilter  `json:"filter,omitempty"`
	Network *api.TransformNetwork `json:"network,omitempty"`
}

type Extract struct {
	Type       string                    `json:"type"`
	Aggregates []api.AggregateDefinition `json:"aggregates,omitempty"`
}

type Encode struct {
	Type  string           `json:"type"`
	Prom  *api.PromEncode  `json:"prom,omitempty"`
	Kafka *api.EncodeKafka `json:"kafka,omitempty"`
}

type Write struct {
	Type   string           `json:"type"`
	Loki   *api.WriteLoki   `json:"loki,omitempty"`
	Stdout *api.WriteStdout `json:"stdout,omitempty"`
}

// ParseConfig creates the internal unmarshalled representation from the Pipeline and Parameters json
func ParseConfig() error {
	logrus.Debugf("config.Opt.PipeLine = %v ", Opt.PipeLine)
	err := json.Unmarshal([]byte(Opt.PipeLine), &PipeLine)
	if err != nil {
		logrus.Errorf("error when reading config file: %v", err)
		return err
	}
	logrus.Debugf("stages = %v ", PipeLine)

	err = json.Unmarshal([]byte(Opt.Parameters), &Parameters)
	if err != nil {
		logrus.Errorf("error when reading config file: %v", err)
		return err
	}
	logrus.Debugf("params = %v ", Parameters)
	return nil
}
