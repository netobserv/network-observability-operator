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
	Brokers          []string `yaml:"brokers" json:"brokers" doc:"list of kafka broker addresses"`
	Topic            string   `yaml:"topic" json:"topic" doc:"kafka topic to listen on"`
	GroupId          string   `yaml:"groupid" json:"groupid" doc:"separate groupid for each consumer on specified topic"`
	GroupBalancers   []string `yaml:"groupBalancers" json:"groupBalancers" doc:"list of balancing strategies (range, roundRobin, rackAffinity)"`
	StartOffset      string   `yaml:"startOffset" json:"startOffset" doc:"FirstOffset (least recent - default) or LastOffset (most recent) offset available for a partition"`
	BatchReadTimeout int64    `yaml:"batchReadTimeout" json:"batchReadTimeout" doc:"how often (in milliseconds) to process input"`
}
