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

import (
	"errors"
	"regexp"
)

type TransformFilter struct {
	Rules []TransformFilterRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"list of filter rules, each includes:"`
}

func (tf *TransformFilter) Preprocess() error {
	var errs []error
	for i := range tf.Rules {
		if err := tf.Rules[i].preprocess(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
	KeepEntry                TransformFilterEnum = "keep_entry"                   // keeps the entry if the set of rules are all satisfied
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

type TransformFilterKeepEntryEnum string

const (
	KeepEntryIfExists        TransformFilterKeepEntryEnum = "keep_entry_if_exists"          // keeps the entry if the field exists
	KeepEntryIfDoesntExist   TransformFilterKeepEntryEnum = "keep_entry_if_doesnt_exist"    // keeps the entry if the field does not exist
	KeepEntryIfEqual         TransformFilterKeepEntryEnum = "keep_entry_if_equal"           // keeps the entry if the field value equals specified value
	KeepEntryIfNotEqual      TransformFilterKeepEntryEnum = "keep_entry_if_not_equal"       // keeps the entry if the field value does not equal specified value
	KeepEntryIfRegexMatch    TransformFilterKeepEntryEnum = "keep_entry_if_regex_match"     // keeps the entry if the field value matches the specified regex
	KeepEntryIfNotRegexMatch TransformFilterKeepEntryEnum = "keep_entry_if_not_regex_match" // keeps the entry if the field value does not match the specified regex
)

type TransformFilterRule struct {
	Type                    TransformFilterEnum              `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	RemoveField             *TransformFilterGenericRule      `yaml:"removeField,omitempty" json:"removeField,omitempty" doc:"configuration for remove_field rule"`
	RemoveEntry             *TransformFilterGenericRule      `yaml:"removeEntry,omitempty" json:"removeEntry,omitempty" doc:"configuration for remove_entry_* rules"`
	RemoveEntryAllSatisfied []*RemoveEntryRule               `yaml:"removeEntryAllSatisfied,omitempty" json:"removeEntryAllSatisfied,omitempty" doc:"configuration for remove_entry_all_satisfied rule"`
	KeepEntryAllSatisfied   []*KeepEntryRule                 `yaml:"keepEntryAllSatisfied,omitempty" json:"keepEntryAllSatisfied,omitempty" doc:"configuration for keep_entry rule"`
	KeepEntrySampling       uint16                           `yaml:"keepEntrySampling,omitempty" json:"keepEntrySampling,omitempty" doc:"sampling value for keep_entry type: 1 flow on <sampling> is kept"`
	AddField                *TransformFilterGenericRule      `yaml:"addField,omitempty" json:"addField,omitempty" doc:"configuration for add_field rule"`
	AddFieldIfDoesntExist   *TransformFilterGenericRule      `yaml:"addFieldIfDoesntExist,omitempty" json:"addFieldIfDoesntExist,omitempty" doc:"configuration for add_field_if_doesnt_exist rule"`
	AddFieldIf              *TransformFilterRuleWithAssignee `yaml:"addFieldIf,omitempty" json:"addFieldIf,omitempty" doc:"configuration for add_field_if rule"`
	AddRegExIf              *TransformFilterRuleWithAssignee `yaml:"addRegexIf,omitempty" json:"addRegexIf,omitempty" doc:"configuration for add_regex_if rule"`
	AddLabel                *TransformFilterGenericRule      `yaml:"addLabel,omitempty" json:"addLabel,omitempty" doc:"configuration for add_label rule"`
	AddLabelIf              *TransformFilterRuleWithAssignee `yaml:"addLabelIf,omitempty" json:"addLabelIf,omitempty" doc:"configuration for add_label_if rule"`
	ConditionalSampling     []*SamplingCondition             `yaml:"conditionalSampling,omitempty" json:"conditionalSampling,omitempty" doc:"sampling configuration rules"`
}

func (r *TransformFilterRule) preprocess() error {
	var errs []error
	if r.RemoveField != nil {
		if err := r.RemoveField.preprocess(false); err != nil {
			errs = append(errs, err)
		}
	}
	if r.RemoveEntry != nil {
		if err := r.RemoveEntry.preprocess(false); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range r.RemoveEntryAllSatisfied {
		if err := r.RemoveEntryAllSatisfied[i].RemoveEntry.preprocess(false); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range r.KeepEntryAllSatisfied {
		err := r.KeepEntryAllSatisfied[i].KeepEntry.preprocess(r.KeepEntryAllSatisfied[i].Type == KeepEntryIfRegexMatch || r.KeepEntryAllSatisfied[i].Type == KeepEntryIfNotRegexMatch)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for i := range r.ConditionalSampling {
		if err := r.ConditionalSampling[i].preprocess(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type TransformFilterGenericRule struct {
	Input   string      `yaml:"input,omitempty" json:"input,omitempty" doc:"entry input field"`
	Value   interface{} `yaml:"value,omitempty" json:"value,omitempty" doc:"specified value of input field:"`
	CastInt bool        `yaml:"castInt,omitempty" json:"castInt,omitempty" doc:"set true to cast the value field as an int (numeric values are float64 otherwise)"`
}

func (r *TransformFilterGenericRule) preprocess(isRegex bool) error {
	if isRegex {
		if s, ok := r.Value.(string); ok {
			v, err := regexp.Compile(s)
			if err != nil {
				r.Value = nil
				return err
			}
			r.Value = v
		} else {
			r.Value = nil
			return errors.New("regex filter expects string value")
		}
	} else if r.CastInt {
		if f, ok := r.Value.(float64); ok {
			r.Value = int(f)
		}
	}
	return nil
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

type KeepEntryRule struct {
	Type      TransformFilterKeepEntryEnum `yaml:"type,omitempty" json:"type,omitempty" doc:"(enum) one of the following:"`
	KeepEntry *TransformFilterGenericRule  `yaml:"keepEntry,omitempty" json:"keepEntry,omitempty" doc:"configuration for keep_entry_* rules"`
}

type SamplingCondition struct {
	Value uint16             `yaml:"value,omitempty" json:"value,omitempty" doc:"sampling value: 1 flow on <sampling> is kept"`
	Rules []*RemoveEntryRule `yaml:"rules,omitempty" json:"rules,omitempty" doc:"rules to be satisfied for this sampling configuration"`
}

func (s *SamplingCondition) preprocess() error {
	var errs []error
	for i := range s.Rules {
		if err := s.Rules[i].RemoveEntry.preprocess(false); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
