package config

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
)

type Pipeline struct {
	stages []Stage
	config []StageParam
}

type PipelineBuilder struct {
	lastStage string
	pipeline  *Pipeline
}

func NewCollectorPipeline(name string, ingest api.IngestCollector) PipelineBuilder {
	p := Pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{{Name: name, Ingest: &Ingest{Type: api.CollectorType, Collector: &ingest}}},
	}
	return PipelineBuilder{pipeline: &p, lastStage: name}
}

func NewGRPCPipeline(name string, ingest api.IngestGRPCProto) PipelineBuilder {
	p := Pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{{Name: name, Ingest: &Ingest{Type: api.GRPCType, GRPC: &ingest}}},
	}
	return PipelineBuilder{pipeline: &p, lastStage: name}
}

func NewKafkaPipeline(name string, ingest api.IngestKafka) PipelineBuilder {
	p := Pipeline{
		stages: []Stage{{Name: name}},
		config: []StageParam{{Name: name, Ingest: &Ingest{Type: api.KafkaType, Kafka: &ingest}}},
	}
	return PipelineBuilder{pipeline: &p, lastStage: name}
}

func (b *PipelineBuilder) next(name string, param StageParam) PipelineBuilder {
	b.pipeline.stages = append(b.pipeline.stages, Stage{Name: name, Follows: b.lastStage})
	b.pipeline.config = append(b.pipeline.config, param)
	return PipelineBuilder{pipeline: b.pipeline, lastStage: name}
}

func (b *PipelineBuilder) DecodeJSON(name string) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Decode: &Decode{Type: api.JSONType}})
}

func (b *PipelineBuilder) DecodeProtobuf(name string) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Decode: &Decode{Type: api.PBType}})
}

func (b *PipelineBuilder) DecodeAWS(name string, aws api.DecodeAws) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Decode: &Decode{Type: api.AWSType, Aws: &aws}})
}

func (b *PipelineBuilder) Aggregate(name string, aggs []api.AggregateDefinition) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Extract: &Extract{Type: api.AggregateType, Aggregates: aggs}})
}

func (b *PipelineBuilder) TransformGeneric(name string, gen api.TransformGeneric) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Transform: &Transform{Type: api.GenericType, Generic: &gen}})
}

func (b *PipelineBuilder) TransformFilter(name string, filter api.TransformFilter) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Transform: &Transform{Type: api.FilterType, Filter: &filter}})
}

func (b *PipelineBuilder) TransformNetwork(name string, nw api.TransformNetwork) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Transform: &Transform{Type: api.NetworkType, Network: &nw}})
}

func (b *PipelineBuilder) EncodePrometheus(name string, prom api.PromEncode) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Encode: &Encode{Type: api.PromType, Prom: &prom}})
}

func (b *PipelineBuilder) EncodeKafka(name string, kafka api.EncodeKafka) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Encode: &Encode{Type: api.KafkaType, Kafka: &kafka}})
}

func (b *PipelineBuilder) WriteStdout(name string, stdout api.WriteStdout) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Write: &Write{Type: api.StdoutType, Stdout: &stdout}})
}

func (b *PipelineBuilder) WriteLoki(name string, loki api.WriteLoki) PipelineBuilder {
	return b.next(name, StageParam{Name: name, Write: &Write{Type: api.LokiType, Loki: &loki}})
}

func (b *PipelineBuilder) GetStages() []Stage {
	return b.pipeline.stages
}

func (b *PipelineBuilder) GetStageParams() []StageParam {
	return b.pipeline.config
}
