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
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
)

func NewIPFIXParams(name string, ingest api.IngestIpfix) StageParam {
	return StageParam{Name: name, Ingest: &Ingest{Type: api.IpfixType, Ipfix: &ingest}}
}

func NewGRPCParams(name string, ingest api.IngestGRPCProto) StageParam {
	return StageParam{Name: name, Ingest: &Ingest{Type: api.GRPCType, GRPC: &ingest}}
}

//nolint:golint,gocritic
func NewKafkaParams(name string, ingest api.IngestKafka) StageParam {
	return StageParam{Name: name, Ingest: &Ingest{Type: api.KafkaType, Kafka: &ingest}}
}

func NewAggregateParams(name string, aggs api.Aggregates) StageParam {
	return StageParam{Name: name, Extract: &Extract{Type: api.AggregateType, Aggregates: &aggs}}
}

func NewTransformGenericParams(name string, gen api.TransformGeneric) StageParam {
	return StageParam{Name: name, Transform: &Transform{Type: api.GenericType, Generic: &gen}}
}

func NewTransformFilterParams(name string, filter api.TransformFilter) StageParam {
	return StageParam{Name: name, Transform: &Transform{Type: api.FilterType, Filter: &filter}}
}

//nolint:golint,gocritic
func NewTransformNetworkParams(name string, nw api.TransformNetwork) StageParam {
	return StageParam{Name: name, Transform: &Transform{Type: api.NetworkType, Network: &nw}}
}

//nolint:golint,gocritic
func NewConnTrackParams(name string, ct api.ConnTrack) StageParam {
	return StageParam{Name: name, Extract: &Extract{Type: api.ConnTrackType, ConnTrack: &ct}}
}

func NewTimbasedParams(name string, ct api.ExtractTimebased) StageParam {
	return StageParam{Name: name, Extract: &Extract{Type: api.TimebasedType, Timebased: &ct}}
}

func NewEncodePrometheusParams(name string, prom api.PromEncode) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.PromType, Prom: &prom}}
}

//nolint:golint,gocritic
func NewEncodeKafkaParams(name string, kafka api.EncodeKafka) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.KafkaType, Kafka: &kafka}}
}

//nolint:golint,gocritic
func NewEncodeS3Params(name string, s3 api.EncodeS3) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.S3Type, S3: &s3}}
}

//nolint:golint,gocritic
func NewEncodeOtelLogsParams(name string, otelLogs api.EncodeOtlpLogs) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.OtlpLogsType, OtlpLogs: &otelLogs}}
}

//nolint:golint,gocritic
func NewEncodeOtelMetricsParams(name string, otelMetrics api.EncodeOtlpMetrics) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.OtlpMetricsType, OtlpMetrics: &otelMetrics}}
}

//nolint:golint,gocritic
func NewEncodeOtelTracesParams(name string, otelTraces api.EncodeOtlpTraces) StageParam {
	return StageParam{Name: name, Encode: &Encode{Type: api.OtlpTracesType, OtlpTraces: &otelTraces}}
}

func NewWriteStdoutParams(name string, stdout api.WriteStdout) StageParam {
	return StageParam{Name: name, Write: &Write{Type: api.StdoutType, Stdout: &stdout}}
}

//nolint:golint,gocritic
func NewWriteLokiParams(name string, loki api.WriteLoki) StageParam {
	return StageParam{Name: name, Write: &Write{Type: api.LokiType, Loki: &loki}}
}

func NewWriteIpfixParams(name string, ipfix api.WriteIpfix) StageParam {
	return StageParam{Name: name, Write: &Write{Type: api.IpfixType, Ipfix: &ipfix}}
}
