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

type TransformFilter struct {
	Rules []TransformFilterRule `yaml:"rules" json:"rules" doc:"list of filter rules, each includes:"`
}

type TransformFilterOperationEnum struct {
	RemoveField              string `yaml:"remove_field" json:"remove_field field from the entry"`
	RemoveEntryIfExists      string `yaml:"remove_entry_if_exists" json:"remove_entry_if_exists" doc:"removes the entry if the field exists"`
	RemoveEntryIfDoesntExist string `yaml:"remove_entry_if_doesnt_exist" json:"remove_entry_if_doesnt_exist" doc:"removes the entry if the field doesnt exist"`
}

func TransformFilterOperationName(operation string) string {
	return GetEnumName(TransformFilterOperationEnum{}, operation)
}

type TransformFilterRule struct {
	Input string `yaml:"input" json:"input" doc:"entry input field"`
	Type  string `yaml:"type" json:"type" enum:"TransformFilterOperationEnum" doc:"one of the following:"`
}
