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
	Rules []TransformFilterRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of filter rules, each includes:"`
}

type TransformFilterOperationEnum struct {
	RemoveField              string `yaml:"remove_field" json:"remove_field" doc:"removes the field from the entry"`
	RemoveEntryIfExists      string `yaml:"remove_entry_if_exists" json:"remove_entry_if_exists" doc:"removes the entry if the field exists"`
	RemoveEntryIfDoesntExist string `yaml:"remove_entry_if_doesnt_exist" json:"remove_entry_if_doesnt_exist" doc:"removes the entry if the field does not exist"`
	RemoveEntryIfEqual       string `yaml:"remove_entry_if_equal" json:"remove_entry_if_equal" doc:"removes the entry if the field value equals specified value"`
	RemoveEntryIfNotEqual    string `yaml:"remove_entry_if_not_equal" json:"remove_entry_if_not_equal" doc:"removes the entry if the field value does not equal specified value"`
	AddFieldIfDoesntExist    string `yaml:"add_field_if_doesnt_exist" json:"add_field_if_doesnt_exist" doc:"adds a field to the entry if the field does not exist"`
	AddFieldIf               string `yaml:"add_field_if" json:"add_field_if" doc:"add output field set to assignee if input field satisfies criteria from parameters field"`
	AddRegExIf               string `yaml:"add_regex_if" json:"add_regex_if" doc:"add output field if input field satisfies regex pattern from parameters field"`
}

func TransformFilterOperationName(operation string) string {
	return GetEnumName(TransformFilterOperationEnum{}, operation)
}

type TransformFilterRule struct {
	Input      string      `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string      `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Type       string      `yaml:"type,omitempty" json:"type,omitempty" enum:"TransformFilterOperationEnum" doc:"one of the following:"`
	Value      interface{} `yaml:"value,omitempty" json:"value,omitempty" doc:"specified value of input field:"`
	Parameters string      `yaml:"parameters,omitempty" json:"parameters,omitempty" doc:"parameters specific to type"`
	Assignee   string      `yaml:"assignee,omitempty" json:"assignee,omitempty" doc:"value needs to assign to output field"`
}
