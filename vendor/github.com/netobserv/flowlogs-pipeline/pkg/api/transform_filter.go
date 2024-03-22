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

type TransformFilterEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	RemoveField              TransformFilterEnum = "remove_field"                 // removes the field from the entry
	RemoveEntryIfExists      TransformFilterEnum = "remove_entry_if_exists"       // removes the entry if the field exists
	RemoveEntryIfDoesntExist TransformFilterEnum = "remove_entry_if_doesnt_exist" // removes the entry if the field does not exist
	RemoveEntryIfEqual       TransformFilterEnum = "remove_entry_if_equal"        // removes the entry if the field value equals specified value
	RemoveEntryIfNotEqual    TransformFilterEnum = "remove_entry_if_not_equal"    // removes the entry if the field value does not equal specified value
	AddField                 TransformFilterEnum = "add_field"                    // adds (input) field to the entry; overrides previous value if present (key=input, value=value)
	AddFieldIfDoesntExist    TransformFilterEnum = "add_field_if_doesnt_exist"    // adds a field to the entry if the field does not exist
	AddFieldIf               TransformFilterEnum = "add_field_if"                 // add output field set to assignee if input field satisfies criteria from parameters field
	AddRegExIf               TransformFilterEnum = "add_regex_if"                 // add output field if input field satisfies regex pattern from parameters field
	AddLabel                 TransformFilterEnum = "add_label"                    // add (input) field to list of labels with value taken from Value field (key=input, value=value)
	AddLabelIf               TransformFilterEnum = "add_label_if"                 // add output field to list of labels with value taken from assignee field if input field satisfies criteria from parameters field
)

type TransformFilterRule struct {
	Type                     TransformFilterEnum              `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	RemoveField              *TransformFilterGenericRule      `yaml:"removeField,omitempty" json:"removeField,omitempty" doc:"configuration for remove_field rule"`
	RemoveEntryIfExists      *TransformFilterGenericRule      `yaml:"removeEntryIfExists,omitempty" json:"removeEntryIfExists,omitempty" doc:"configuration for remove_entry_if_exists rule"`
	RemoveEntryIfDoesntExist *TransformFilterGenericRule      `yaml:"removeEntryIfDoesntExist,omitempty" json:"removeEntryIfDoesntExist,omitempty" doc:"configuration for remove_entry_if_doesnt_exist rule"`
	RemoveEntryIfEqual       *TransformFilterGenericRule      `yaml:"removeEntryIfEqual,omitempty" json:"removeEntryIfEqual,omitempty" doc:"configuration for remove_entry_if_equal rule"`
	RemoveEntryIfNotEqual    *TransformFilterGenericRule      `yaml:"removeEntryIfNotEqual,omitempty" json:"removeEntryIfNotEqual,omitempty" doc:"configuration for remove_entry_if_not_equal rule"`
	AddField                 *TransformFilterGenericRule      `yaml:"addField,omitempty" json:"addField,omitempty" doc:"configuration for add_field rule"`
	AddFieldIfDoesntExist    *TransformFilterGenericRule      `yaml:"addFieldIfDoesntExist,omitempty" json:"addFieldIfDoesntExist,omitempty" doc:"configuration for add_field_if_doesnt_exist rule"`
	AddFieldIf               *TransformFilterRuleWithAssignee `yaml:"addFieldIf,omitempty" json:"addFieldIf,omitempty" doc:"configuration for add_field_if rule"`
	AddRegExIf               *TransformFilterRuleWithAssignee `yaml:"addRegexIf,omitempty" json:"addRegexIf,omitempty" doc:"configuration for add_regex_if rule"`
	AddLabel                 *TransformFilterGenericRule      `yaml:"addLabel,omitempty" json:"addLabel,omitempty" doc:"configuration for add_label rule"`
	AddLabelIf               *TransformFilterRuleWithAssignee `yaml:"addLabelIf,omitempty" json:"addLabelIf,omitempty" doc:"configuration for add_label_if rule"`
}

type TransformFilterGenericRule struct {
	Input string      `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Value interface{} `yaml:"value,omitempty" json:"value,omitempty" doc:"specified value of input field:"`
}

type TransformFilterRuleWithAssignee struct {
	Input      string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Parameters string `yaml:"parameters,omitempty" json:"parameters,omitempty" doc:"parameters specific to type"`
	Assignee   string `yaml:"assignee,omitempty" json:"assignee,omitempty" doc:"value needs to assign to output field"`
}
