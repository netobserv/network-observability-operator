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

type EncodeKafka struct {
	Address      string      `yaml:"address" json:"address" doc:"address of kafka server"`
	Topic        string      `yaml:"topic" json:"topic" doc:"kafka topic to write to"`
	Balancer     string      `yaml:"balancer,omitempty" json:"balancer,omitempty" enum:"KafkaEncodeBalancerEnum" doc:"one of the following:"`
	WriteTimeout int64       `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" doc:"timeout (in seconds) for write operation performed by the Writer"`
	ReadTimeout  int64       `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" doc:"timeout (in seconds) for read operation performed by the Writer"`
	BatchBytes   int64       `yaml:"batchBytes,omitempty" json:"batchBytes,omitempty" doc:"limit the maximum size of a request in bytes before being sent to a partition"`
	BatchSize    int         `yaml:"batchSize,omitempty" json:"batchSize,omitempty" doc:"limit on how many messages will be buffered before being sent to a partition"`
	TLS          *ClientTLS  `yaml:"tls" json:"tls" doc:"TLS client configuration (optional)"`
	SASL         *SASLConfig `yaml:"sasl" json:"sasl" doc:"SASL configuration (optional)"`
}

type KafkaEncodeBalancerEnum struct {
	RoundRobin string `yaml:"roundRobin" json:"roundRobin" doc:"RoundRobin balancer"`
	LeastBytes string `yaml:"leastBytes" json:"leastBytes" doc:"LeastBytes balancer"`
	Hash       string `yaml:"hash" json:"hash" doc:"Hash balancer"`
	Crc32      string `yaml:"crc32" json:"crc32" doc:"Crc32 balancer"`
	Murmur2    string `yaml:"murmur2" json:"murmur2" doc:"Murmur2 balancer"`
}

func KafkaEncodeBalancerName(operation string) string {
	return GetEnumName(KafkaEncodeBalancerEnum{}, operation)
}
