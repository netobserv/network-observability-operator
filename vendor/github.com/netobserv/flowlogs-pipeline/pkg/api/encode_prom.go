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

type PromTLSConf struct {
	CertPath string `yaml:"certPath,omitempty" json:"certPath,omitempty" doc:"path to the certificate file"`
	KeyPath  string `yaml:"keyPath,omitempty" json:"keyPath,omitempty" doc:"path to the key file"`
}

type PromEncode struct {
	Metrics    PromMetricsItems `yaml:"metrics,omitempty" json:"metrics,omitempty" doc:"list of prometheus metric definitions, each includes:"`
	Prefix     string           `yaml:"prefix,omitempty" json:"prefix,omitempty" doc:"prefix added to each metric name"`
	ExpiryTime Duration         `yaml:"expiryTime,omitempty" json:"expiryTime,omitempty" doc:"time duration of no-flow to wait before deleting prometheus data item"`
	MaxMetrics int              `yaml:"maxMetrics,omitempty" json:"maxMetrics,omitempty" doc:"maximum number of metrics to report (default: unlimited)"`
}

type PromEncodeOperationEnum struct {
	Gauge        string `yaml:"gauge" json:"gauge" doc:"single numerical value that can arbitrarily go up and down"`
	Counter      string `yaml:"counter" json:"counter" doc:"monotonically increasing counter whose value can only increase"`
	Histogram    string `yaml:"histogram" json:"histogram" doc:"counts samples in configurable buckets"`
	AggHistogram string `yaml:"agg_histogram" json:"agg_histogram" doc:"counts samples in configurable buckets, pre-aggregated via an Aggregate stage"`
}

func PromEncodeOperationName(operation string) string {
	return GetEnumName(PromEncodeOperationEnum{}, operation)
}

type PromMetricsItem struct {
	Name     string            `yaml:"name" json:"name" doc:"the metric name"`
	Type     string            `yaml:"type" json:"type" enum:"PromEncodeOperationEnum" doc:"one of the following:"`
	Filter   PromMetricsFilter `yaml:"filter" json:"filter" doc:"an optional criterion to filter entries by"`
	ValueKey string            `yaml:"valueKey" json:"valueKey" doc:"entry key from which to resolve metric value"`
	Labels   []string          `yaml:"labels" json:"labels" doc:"labels to be associated with the metric"`
	Buckets  []float64         `yaml:"buckets" json:"buckets" doc:"histogram buckets"`
}

type PromMetricsItems []PromMetricsItem

type PromMetricsFilter struct {
	Key   string `yaml:"key" json:"key" doc:"the key to match and filter by"`
	Value string `yaml:"value" json:"value" doc:"the value to match and filter by"`
}
