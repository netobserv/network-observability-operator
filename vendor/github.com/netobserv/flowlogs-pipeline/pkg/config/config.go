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

type ConfigFileStruct struct {
	LogLevel   string       `yaml:"log-level,omitempty" json:"log-level,omitempty"`
	Pipeline   []Stage      `yaml:"pipeline,omitempty" json:"pipeline,omitempty"`
	Parameters []StageParam `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

type Stage struct {
	Name    string `yaml:"name" json:"name"`
	Follows string `yaml:"follows,omitempty" json:"follows,omitempty"`
}

type StageParam struct {
	Name      string     `yaml:"name" json:"name"`
	Ingest    *Ingest    `yaml:"ingest,omitempty" json:"ingest,omitempty"`
	Transform *Transform `yaml:"transform,omitempty" json:"transform,omitempty"`
	Extract   *Extract   `yaml:"extract,omitempty" json:"extract,omitempty"`
	Encode    *Encode    `yaml:"encode,omitempty" json:"encode,omitempty"`
	Write     *Write     `yaml:"write,omitempty" json:"write,omitempty"`
}

type Ingest struct {
	Type      string               `yaml:"type" json:"type"`
	File      *File                `yaml:"file,omitempty" json:"file,omitempty"`
	Collector *api.IngestCollector `yaml:"collector,omitempty" json:"collector,omitempty"`
	Kafka     *api.IngestKafka     `yaml:"kafka,omitempty" json:"kafka,omitempty"`
	GRPC      *api.IngestGRPCProto `yaml:"grpc,omitempty" json:"grpc,omitempty"`
}

type File struct {
	Filename string      `yaml:"filename" json:"filename"`
	Decoder  api.Decoder `yaml:"decoder" json:"decoder"`
	Loop     bool        `yaml:"loop" json:"loop"`
	Chunks   int         `yaml:"chunks" json:"chunks"`
}

type Transform struct {
	Type    string                `yaml:"type" json:"type"`
	Generic *api.TransformGeneric `yaml:"generic,omitempty" json:"generic,omitempty"`
	Filter  *api.TransformFilter  `yaml:"filter,omitempty" json:"filter,omitempty"`
	Network *api.TransformNetwork `yaml:"network,omitempty" json:"network,omitempty"`
}

type Extract struct {
	Type       string                    `yaml:"type" json:"type"`
	Aggregates []api.AggregateDefinition `yaml:"aggregates,omitempty" json:"aggregates,omitempty"`
}

type Encode struct {
	Type  string           `yaml:"type" json:"type"`
	Prom  *api.PromEncode  `yaml:"prom,omitempty" json:"prom,omitempty"`
	Kafka *api.EncodeKafka `yaml:"kafka,omitempty" json:"kafka,omitempty"`
}

type Write struct {
	Type   string           `yaml:"type" json:"type"`
	Loki   *api.WriteLoki   `yaml:"loki,omitempty" json:"loki,omitempty"`
	Stdout *api.WriteStdout `yaml:"stdout,omitempty" json:"stdout,omitempty"`
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
