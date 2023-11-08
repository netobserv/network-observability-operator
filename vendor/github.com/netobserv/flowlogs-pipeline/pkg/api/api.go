/*
 * Copyright (C) 2022 IBM, Inc.
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

package api

const (
	FileType                     = "file"
	FileLoopType                 = "file_loop"
	FileChunksType               = "file_chunks"
	SyntheticType                = "synthetic"
	CollectorType                = "collector"
	StdinType                    = "stdin"
	GRPCType                     = "grpc"
	FakeType                     = "fake"
	KafkaType                    = "kafka"
	S3Type                       = "s3"
	StdoutType                   = "stdout"
	LokiType                     = "loki"
	IpfixType                    = "ipfix"
	AggregateType                = "aggregates"
	TimebasedType                = "timebased"
	PromType                     = "prom"
	GenericType                  = "generic"
	NetworkType                  = "network"
	FilterType                   = "filter"
	ConnTrackType                = "conntrack"
	NoneType                     = "none"
	AddRegExIfRuleType           = "add_regex_if"
	AddIfRuleType                = "add_if"
	AddSubnetRuleType            = "add_subnet"
	AddLocationRuleType          = "add_location"
	AddServiceRuleType           = "add_service"
	AddKubernetesRuleType        = "add_kubernetes"
	ReinterpretDirectionRuleType = "reinterpret_direction"
	PromFilterExact              = "exact"
	PromFilterPresence           = "presence"
	PromFilterAbsence            = "absence"
	PromFilterRegex              = "regex"

	TagYaml = "yaml"
	TagDoc  = "doc"
	TagEnum = "enum"
)

// Note: items beginning with doc: "## title" are top level items that get divided into sections inside api.md.

type API struct {
	PromEncode         PromEncode       `yaml:"prom" doc:"## Prometheus encode API\nFollowing is the supported API format for prometheus encode:\n"`
	KafkaEncode        EncodeKafka      `yaml:"kafka" doc:"## Kafka encode API\nFollowing is the supported API format for kafka encode:\n"`
	S3Encode           EncodeS3         `yaml:"s3" doc:"## S3 encode API\nFollowing is the supported API format for S3 encode:\n"`
	IngestCollector    IngestCollector  `yaml:"collector" doc:"## Ingest collector API\nFollowing is the supported API format for the NetFlow / IPFIX collector:\n"`
	IngestKafka        IngestKafka      `yaml:"kafka" doc:"## Ingest Kafka API\nFollowing is the supported API format for the kafka ingest:\n"`
	IngestGRPCProto    IngestGRPCProto  `yaml:"grpc" doc:"## Ingest GRPC from Network Observability eBPF Agent\nFollowing is the supported API format for the Network Observability eBPF ingest:\n"`
	IngestStdin        IngestStdin      `yaml:"stdin" doc:"## Ingest Standard Input\nFollowing is the supported API format for the standard input ingest:\n"`
	TransformGeneric   TransformGeneric `yaml:"generic" doc:"## Transform Generic API\nFollowing is the supported API format for generic transformations:\n"`
	TransformFilter    TransformFilter  `yaml:"filter" doc:"## Transform Filter API\nFollowing is the supported API format for filter transformations:\n"`
	TransformNetwork   TransformNetwork `yaml:"network" doc:"## Transform Network API\nFollowing is the supported API format for network transformations:\n"`
	WriteLoki          WriteLoki        `yaml:"loki" doc:"## Write Loki API\nFollowing is the supported API format for writing to loki:\n"`
	WriteStdout        WriteStdout      `yaml:"stdout" doc:"## Write Standard Output\nFollowing is the supported API format for writing to standard output:\n"`
	ExtractAggregate   Aggregates       `yaml:"aggregates" doc:"## Aggregate metrics API\nFollowing is the supported API format for specifying metrics aggregations:\n"`
	ConnectionTracking ConnTrack        `yaml:"conntrack" doc:"## Connection tracking API\nFollowing is the supported API format for specifying connection tracking:\n"`
	ExtractTimebased   ExtractTimebased `yaml:"timebased" doc:"## Time-based Filters API\nFollowing is the supported API format for specifying metrics time-based filters:\n"`
}
