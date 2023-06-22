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

type IngestKafka struct {
	Brokers           []string    `yaml:"brokers,omitempty" json:"brokers,omitempty" doc:"list of kafka broker addresses"`
	Topic             string      `yaml:"topic,omitempty" json:"topic,omitempty" doc:"kafka topic to listen on"`
	GroupId           string      `yaml:"groupid,omitempty" json:"groupid,omitempty" doc:"separate groupid for each consumer on specified topic"`
	GroupBalancers    []string    `yaml:"groupBalancers,omitempty" json:"groupBalancers,omitempty" doc:"list of balancing strategies (range, roundRobin, rackAffinity)"`
	StartOffset       string      `yaml:"startOffset,omitempty" json:"startOffset,omitempty" doc:"FirstOffset (least recent - default) or LastOffset (most recent) offset available for a partition"`
	BatchReadTimeout  int64       `yaml:"batchReadTimeout,omitempty" json:"batchReadTimeout,omitempty" doc:"how often (in milliseconds) to process input"`
	Decoder           Decoder     `yaml:"decoder,omitempty" json:"decoder" doc:"decoder to use (E.g. json or protobuf)"`
	BatchMaxLen       int         `yaml:"batchMaxLen,omitempty" json:"batchMaxLen,omitempty" doc:"the number of accumulated flows before being forwarded for processing"`
	PullQueueCapacity int         `yaml:"pullQueueCapacity,omitempty" json:"pullQueueCapacity,omitempty" doc:"the capacity of the queue use to store pulled flows"`
	PullMaxBytes      int         `yaml:"pullMaxBytes,omitempty" json:"pullMaxBytes,omitempty" doc:"the maximum number of bytes being pulled from kafka"`
	CommitInterval    int64       `yaml:"commitInterval,omitempty" json:"commitInterval,omitempty" doc:"the interval (in milliseconds) at which offsets are committed to the broker.  If 0, commits will be handled synchronously."`
	TLS               *ClientTLS  `yaml:"tls" json:"tls" doc:"TLS client configuration (optional)"`
	SASL              *SASLConfig `yaml:"sasl" json:"sasl" doc:"SASL configuration (optional)"`
}
