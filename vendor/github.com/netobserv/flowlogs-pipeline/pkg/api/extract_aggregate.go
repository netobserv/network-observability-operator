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

type AggregateBy []string
type AggregateOperation string

type AggregateDefinition struct {
	Name      string             `yaml:"name,omitempty" json:"name,omitempty" doc:"description of aggregation result"`
	By        AggregateBy        `yaml:"by,omitempty" json:"by,omitempty" doc:"list of fields on which to aggregate"`
	Operation AggregateOperation `yaml:"operation,omitempty" json:"operation,omitempty" doc:"sum, min, max, avg or raw_values"`
	RecordKey string             `yaml:"recordKey,omitempty" json:"recordKey,omitempty" doc:"internal field on which to perform the operation"`
	TopK      int                `yaml:"topK,omitempty" json:"topK,omitempty" doc:"number of highest incidence to report (default - report all)"`
}
