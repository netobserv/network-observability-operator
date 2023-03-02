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

type FilterOperationEnum struct {
	FilterOperationSum  string `yaml:"sum" json:"sum" doc:"set output field to sum of parameters fields in the time window"`
	FilterOperationAvg  string `yaml:"avg" json:"avg" doc:"set output field to average of parameters fields in the time window"`
	FilterOperationMin  string `yaml:"min" json:"min" doc:"set output field to minimum of parameters fields in the time window"`
	FilterOperationMax  string `yaml:"max" json:"max" doc:"set output field to maximum of parameters fields in the time window"`
	FilterOperationCnt  string `yaml:"count" json:"count" doc:"set output field to number of flows registered in the time window"`
	FilterOperationLast string `yaml:"last" json:"last" doc:"set output field to last of parameters fields in the time window"`
	FilterOperationDiff string `yaml:"diff" json:"diff" doc:"set output field to the difference of the first and last parameters fields in the time window"`
}

func FilterOperationName(operation string) string {
	return GetEnumName(FilterOperationEnum{}, operation)
}

type ExtractTimebased struct {
	Rules []TimebasedFilterRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of filter rules, each includes:"`
}

type TimebasedFilterRule struct {
	Name          string   `yaml:"name,omitempty" json:"name,omitempty" doc:"description of filter result"`
	IndexKey      string   `yaml:"indexKey,omitempty" json:"indexKey,omitempty" doc:"internal field to index TopK"`
	OperationType string   `yaml:"operationType,omitempty" json:"operationType,omitempty" enum:"FilterOperationEnum" doc:"sum, min, max, avg, count, last or diff"`
	OperationKey  string   `yaml:"operationKey,omitempty" json:"operationKey,omitempty" doc:"internal field on which to perform the operation"`
	TopK          int      `yaml:"topK,omitempty" json:"topK,omitempty" doc:"number of highest incidence to report (default - report all)"`
	Reversed      bool     `yaml:"reversed,omitempty" json:"reversed,omitempty" doc:"report lowest incidence instead of highest (default - false)"`
	TimeInterval  Duration `yaml:"timeInterval,omitempty" json:"timeInterval,omitempty" doc:"time duration of data to use to compute the metric"`
}
