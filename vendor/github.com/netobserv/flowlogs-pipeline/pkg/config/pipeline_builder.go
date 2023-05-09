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
	"errors"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
)

type pipeline struct {
	stages []Stage
	config []StageParam
}

// PipelineBuilderStage holds information about a created pipeline stage. This stage can be used to chain a following stage, or several of them (resulting in a fork).
// Example:
//
//	firstStage := NewCollectorPipeline("first stage", ...)
//	secondStage := firstStage.DecodeJSON("second stage")
//	thirdStage := secondStage.WriteLoki("third stage", ...)
//	forkedStage := secondStage.WriteStdout("fork following second stage", ...)
//
// All created stages hold a pointer to the whole pipeline, so that the resulting pipeline can be retrieve from any of the stages:
//
//	forkedStage.GetStages()
//	forkedStage.GetStageParams()
//	// is equivalent to:
//	firstStage.GetStages()
//	firstStage.GetStageParams()
type PipelineBuilderStage struct {
	lastStage string
	pipeline  *pipeline
}

// NewPipeline creates a new pipeline from an existing ingest
func NewPipeline(name string, ingest *Ingest) (PipelineBuilderStage, error) {
	if ingest.Collector != nil {
		return NewCollectorPipeline(name, *ingest.Collector), nil
	}
	if ingest.GRPC != nil {
		return NewGRPCPipeline(name, *ingest.GRPC), nil
	}
	if ingest.Kafka != nil {
		return NewKafkaPipeline(name, *ingest.Kafka), nil
	}
	return PipelineBuilderStage{}, errors.New("Missing ingest params")
}

// NewCollectorPipeline creates a new pipeline from an `IngestCollector` initial stage (listening for NetFlows / IPFIX)
func NewCollectorPipeline(name string, ingest api.IngestCollector) PipelineBuilderStage {
	p := pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{NewCollectorParams(name, ingest)},
	}
	return PipelineBuilderStage{pipeline: &p, lastStage: name}
}

// NewGRPCPipeline creates a new pipeline from an `IngestGRPCProto` initial stage (listening for NetObserv's eBPF agent protobuf)
func NewGRPCPipeline(name string, ingest api.IngestGRPCProto) PipelineBuilderStage {
	p := pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{NewGRPCParams(name, ingest)},
	}
	return PipelineBuilderStage{pipeline: &p, lastStage: name}
}

// NewKafkaPipeline creates a new pipeline from an `IngestKafka` initial stage (listening for flow events on Kafka)
func NewKafkaPipeline(name string, ingest api.IngestKafka) PipelineBuilderStage {
	p := pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{NewKafkaParams(name, ingest)},
	}
	return PipelineBuilderStage{pipeline: &p, lastStage: name}
}

func (b *PipelineBuilderStage) next(name string, param StageParam) PipelineBuilderStage {
	b.pipeline.stages = append(b.pipeline.stages, Stage{Name: name, Follows: b.lastStage})
	b.pipeline.config = append(b.pipeline.config, param)
	return PipelineBuilderStage{pipeline: b.pipeline, lastStage: name}
}

// Aggregate chains the current stage with an aggregate stage and returns that new stage
func (b *PipelineBuilderStage) Aggregate(name string, aggs []api.AggregateDefinition) PipelineBuilderStage {
	return b.next(name, NewAggregateParams(name, aggs))
}

// ExtractTimebased chains the current stage with a ExtractTimebased stage and returns that new stage
func (b *PipelineBuilderStage) ExtractTimebased(name string, tb api.ExtractTimebased) PipelineBuilderStage {
	return b.next(name, StageParam{Name: name, Extract: &Extract{Type: api.TimebasedType, Timebased: &tb}})
}

// TransformGeneric chains the current stage with a TransformGeneric stage and returns that new stage
func (b *PipelineBuilderStage) TransformGeneric(name string, gen api.TransformGeneric) PipelineBuilderStage {
	return b.next(name, NewTransformGenericParams(name, gen))
}

// TransformFilter chains the current stage with a TransformFilter stage and returns that new stage
func (b *PipelineBuilderStage) TransformFilter(name string, filter api.TransformFilter) PipelineBuilderStage {
	return b.next(name, NewTransformFilterParams(name, filter))
}

// TransformNetwork chains the current stage with a TransformNetwork stage and returns that new stage
func (b *PipelineBuilderStage) TransformNetwork(name string, nw api.TransformNetwork) PipelineBuilderStage {
	return b.next(name, NewTransformNetworkParams(name, nw))
}

// ConnTrack chains the current stage with a ConnTrack stage and returns that new stage
func (b *PipelineBuilderStage) ConnTrack(name string, ct api.ConnTrack) PipelineBuilderStage {
	return b.next(name, NewConnTrackParams(name, ct))
}

// EncodePrometheus chains the current stage with a PromEncode stage (to expose metrics in Prometheus format) and returns that new stage
func (b *PipelineBuilderStage) EncodePrometheus(name string, prom api.PromEncode) PipelineBuilderStage {
	return b.next(name, NewEncodePrometheusParams(name, prom))
}

// EncodeKafka chains the current stage with an EncodeKafka stage (writing to a Kafka topic) and returns that new stage
func (b *PipelineBuilderStage) EncodeKafka(name string, kafka api.EncodeKafka) PipelineBuilderStage {
	return b.next(name, NewEncodeKafkaParams(name, kafka))
}

// WriteStdout chains the current stage with a WriteStdout stage and returns that new stage
func (b *PipelineBuilderStage) WriteStdout(name string, stdout api.WriteStdout) PipelineBuilderStage {
	return b.next(name, NewWriteStdoutParams(name, stdout))
}

// WriteLoki chains the current stage with a WriteLoki stage and returns that new stage
func (b *PipelineBuilderStage) WriteLoki(name string, loki api.WriteLoki) PipelineBuilderStage {
	return b.next(name, NewWriteLokiParams(name, loki))
}

// WriteIpfix chains the current stage with a WriteIpfix stage and returns that new stage
func (b *PipelineBuilderStage) WriteIpfix(name string, ipfix api.WriteIpfix) PipelineBuilderStage {
	return b.next(name, NewWriteIpfixParams(name, ipfix))
}

// GetStages returns the current pipeline stages. It can be called from any of the stages, they share the same pipeline reference.
func (b *PipelineBuilderStage) GetStages() []Stage {
	return b.pipeline.stages
}

// GetStageParams returns the current pipeline stage params. It can be called from any of the stages, they share the same pipeline reference.
func (b *PipelineBuilderStage) GetStageParams() []StageParam {
	return b.pipeline.config
}
