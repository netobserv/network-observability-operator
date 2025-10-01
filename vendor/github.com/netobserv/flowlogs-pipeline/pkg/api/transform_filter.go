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
	Rules         []TransformFilterRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of filter rules, each includes:"`
	SamplingField string                `yaml:"samplingField,omitempty" json:"samplingField,omitempty" doc:"sampling field name to be set when sampling is used; if the field already exists in flows, its value is multiplied with the new sampling"`
}

func (tf *TransformFilter) Preprocess() {
	for i := range tf.Rules {
		tf.Rules[i].preprocess()
	}
}

type TransformFilterEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	RemoveField              TransformFilterEnum = "remove_field"                 // removes the field from the entry
	RemoveEntryIfExists      TransformFilterEnum = "remove_entry_if_exists"       // removes the entry if the field exists
	RemoveEntryIfDoesntExist TransformFilterEnum = "remove_entry_if_doesnt_exist" // removes the entry if the field does not exist
	RemoveEntryIfEqual       TransformFilterEnum = "remove_entry_if_equal"        // removes the entry if the field value equals specified value
	RemoveEntryIfNotEqual    TransformFilterEnum = "remove_entry_if_not_equal"    // removes the entry if the field value does not equal specified value
	RemoveEntryAllSatisfied  TransformFilterEnum = "remove_entry_all_satisfied"   // removes the entry if all of the defined rules are satisfied
	KeepEntryQuery           TransformFilterEnum = "keep_entry_query"             // keeps the entry if it matches the query
	AddField                 TransformFilterEnum = "add_field"                    // adds (input) field to the entry; overrides previous value if present (key=input, value=value)
	AddFieldIfDoesntExist    TransformFilterEnum = "add_field_if_doesnt_exist"    // adds a field to the entry if the field does not exist
	AddFieldIf               TransformFilterEnum = "add_field_if"                 // add output field set to assignee if input field satisfies criteria from parameters field
	AddRegExIf               TransformFilterEnum = "add_regex_if"                 // add output field if input field satisfies regex pattern from parameters field
	AddLabel                 TransformFilterEnum = "add_label"                    // add (input) field to list of labels with value taken from Value field (key=input, value=value)
	AddLabelIf               TransformFilterEnum = "add_label_if"                 // add output field to list of labels with value taken from assignee field if input field satisfies criteria from parameters field
	ConditionalSampling      TransformFilterEnum = "conditional_sampling"         // define conditional sampling rules
)

type TransformFilterRemoveEntryEnum string

const (
	RemoveEntryIfExistsD      TransformFilterRemoveEntryEnum = "remove_entry_if_exists"       // removes the entry if the field exists
	RemoveEntryIfDoesntExistD TransformFilterRemoveEntryEnum = "remove_entry_if_doesnt_exist" // removes the entry if the field does not exist
	RemoveEntryIfEqualD       TransformFilterRemoveEntryEnum = "remove_entry_if_equal"        // removes the entry if the field value equals specified value
	RemoveEntryIfNotEqualD    TransformFilterRemoveEntryEnum = "remove_entry_if_not_equal"    // removes the entry if the field value does not equal specified value
)

type TransformFilterRule struct {
	Type                    TransformFilterEnum              `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	RemoveField             *TransformFilterGenericRule      `yaml:"removeField,omitempty" json:"removeField,omitempty" doc:"configuration for remove_field rule"`
	RemoveEntry             *TransformFilterGenericRule      `yaml:"removeEntry,omitempty" json:"removeEntry,omitempty" doc:"configuration for remove_entry_* rules"`
	RemoveEntryAllSatisfied []*RemoveEntryRule               `yaml:"removeEntryAllSatisfied,omitempty" json:"removeEntryAllSatisfied,omitempty" doc:"configuration for remove_entry_all_satisfied rule"`
	KeepEntryQuery          string                           `yaml:"keepEntryQuery,omitempty" json:"keepEntryQuery,omitempty" doc:"configuration for keep_entry rule"`
	KeepEntrySampling       uint16                           `yaml:"keepEntrySampling,omitempty" json:"keepEntrySampling,omitempty" doc:"sampling value for keep_entry type: 1 flow on <sampling> is kept"`
	AddField                *TransformFilterGenericRule      `yaml:"addField,omitempty" json:"addField,omitempty" doc:"configuration for add_field rule"`
	AddFieldIfDoesntExist   *TransformFilterGenericRule      `yaml:"addFieldIfDoesntExist,omitempty" json:"addFieldIfDoesntExist,omitempty" doc:"configuration for add_field_if_doesnt_exist rule"`
	AddFieldIf              *TransformFilterRuleWithAssignee `yaml:"addFieldIf,omitempty" json:"addFieldIf,omitempty" doc:"configuration for add_field_if rule"`
	AddRegExIf              *TransformFilterRuleWithAssignee `yaml:"addRegexIf,omitempty" json:"addRegexIf,omitempty" doc:"configuration for add_regex_if rule"`
	AddLabel                *TransformFilterGenericRule      `yaml:"addLabel,omitempty" json:"addLabel,omitempty" doc:"configuration for add_label rule"`
	AddLabelIf              *TransformFilterRuleWithAssignee `yaml:"addLabelIf,omitempty" json:"addLabelIf,omitempty" doc:"configuration for add_label_if rule"`
	ConditionalSampling     []*SamplingCondition             `yaml:"conditionalSampling,omitempty" json:"conditionalSampling,omitempty" doc:"sampling configuration rules"`
}

func (r *TransformFilterRule) preprocess() {
	if r.RemoveField != nil {
		r.RemoveField.preprocess()
	}
	if r.RemoveEntry != nil {
		r.RemoveEntry.preprocess()
	}
	for i := range r.RemoveEntryAllSatisfied {
		r.RemoveEntryAllSatisfied[i].RemoveEntry.preprocess()
	}
	for i := range r.ConditionalSampling {
		r.ConditionalSampling[i].preprocess()
	}
}

type TransformFilterGenericRule struct {
	Input   string      `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Value   interface{} `yaml:"value,omitempty" json:"value,omitempty" doc:"specified value of input field:"`
	CastInt bool        `yaml:"castInt,omitempty" json:"castInt,omitempty" doc:"set true to cast the value field as an int (numeric values are float64 otherwise)"`
}

func (r *TransformFilterGenericRule) preprocess() {
	if r.CastInt {
		if f, ok := r.Value.(float64); ok {
			r.Value = int(f)
		}
	}
}

type TransformFilterRuleWithAssignee struct {
	Input      string `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Output     string `yaml:"output,omitempty" json:"output,omitempty" doc:"entry output field"`
	Parameters string `yaml:"parameters,omitempty" json:"parameters,omitempty" doc:"parameters specific to type"`
	Assignee   string `yaml:"assignee,omitempty" json:"assignee,omitempty" doc:"value needs to assign to output field"`
}

type RemoveEntryRule struct {
	Type        TransformFilterRemoveEntryEnum `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	RemoveEntry *TransformFilterGenericRule    `yaml:"removeEntry,omitempty" json:"removeEntry,omitempty" doc:"configuration for remove_entry_* rules"`
}

type SamplingCondition struct {
	Value uint16             `yaml:"value,omitempty" json:"value,omitempty" doc:"sampling value: 1 flow on <sampling> is kept"`
	Rules []*RemoveEntryRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"rules to be satisfied for this sampling configuration"`
}

func (s *SamplingCondition) preprocess() {
	for i := range s.Rules {
		s.Rules[i].RemoveEntry.preprocess()
	}
}
