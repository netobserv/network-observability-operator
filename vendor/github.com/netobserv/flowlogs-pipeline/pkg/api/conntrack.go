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
	"fmt"
)

const (
	HashIdFieldName     = "_HashId"
	RecordTypeFieldName = "_RecordType"
)

type ConnTrack struct {
	// TODO: should by a pointer instead?
	KeyDefinition        KeyDefinition `yaml:"keyDefinition,omitempty" doc:"fields that are used to identify the connection"`
	OutputRecordTypes    []string      `yaml:"outputRecordTypes,omitempty" enum:"ConnTrackOutputRecordTypeEnum" doc:"output record types to emit"`
	OutputFields         []OutputField `yaml:"outputFields,omitempty" doc:"list of output fields"`
	EndConnectionTimeout Duration      `yaml:"endConnectionTimeout,omitempty" doc:"duration of time to wait from the last flow log to end a connection"`
}

type ConnTrackOutputRecordTypeEnum struct {
	NewConnection string `yaml:"newConnection" doc:"New connection"`
	EndConnection string `yaml:"endConnection" doc:"End connection"`
	FlowLog       string `yaml:"flowLog" doc:"Flow log"`
}

func ConnTrackOutputRecordTypeName(operation string) string {
	return GetEnumName(ConnTrackOutputRecordTypeEnum{}, operation)
}

type KeyDefinition struct {
	FieldGroups []FieldGroup  `yaml:"fieldGroups,omitempty" doc:"list of field group definitions"`
	Hash        ConnTrackHash `yaml:"hash,omitempty" doc:"how to build the connection hash"`
}

type FieldGroup struct {
	Name   string   `yaml:"name,omitempty" doc:"field group name"`
	Fields []string `yaml:"fields" doc:"list of fields in the group"`
}

// ConnTrackHash determines how to compute the connection hash.
// A and B are treated as the endpoints of the connection.
// When FieldGroupARef and FieldGroupBRef are set, the hash is computed in a way
// that flow logs from A to B will have the same hash as flow logs from B to A.
// When they are not set, a different hash will be computed for A->B and B->A,
// and they are tracked as different connections.
type ConnTrackHash struct {
	FieldGroupRefs []string `yaml:"fieldGroupRefs,omitempty" doc:"list of field group names to build the hash"`
	FieldGroupARef string   `yaml:"fieldGroupARef,omitempty" doc:"field group name of endpoint A"`
	FieldGroupBRef string   `yaml:"fieldGroupBRef,omitempty" doc:"field group name of endpoint B"`
}

type OutputField struct {
	Name      string `yaml:"name,omitempty" doc:"output field name"`
	Operation string `yaml:"operation,omitempty" enum:"ConnTrackOperationEnum" doc:"aggregate operation on the field value"`
	SplitAB   bool   `yaml:"splitAB,omitempty" doc:"When true, 2 output fields will be created. One for A->B and one for B->A flows."`
	Input     string `yaml:"input,omitempty" doc:"The input field to base the operation on. When omitted, 'name' is used"`
}

type ConnTrackOperationEnum struct {
	Sum   string `yaml:"sum" doc:"sum"`
	Count string `yaml:"count" doc:"count"`
	Min   string `yaml:"min" doc:"min"`
	Max   string `yaml:"max" doc:"max"`
}

func ConnTrackOperationName(operation string) string {
	return GetEnumName(ConnTrackOperationEnum{}, operation)
}

func (ct *ConnTrack) Validate() error {
	isGroupAEmpty := ct.KeyDefinition.Hash.FieldGroupARef == ""
	isGroupBEmpty := ct.KeyDefinition.Hash.FieldGroupBRef == ""
	if isGroupAEmpty != isGroupBEmpty { // XOR
		return conntrackInvalidError{fieldGroupABOnlyOneIsSet: true,
			msg: fmt.Errorf("only one of 'fieldGroupARef' and 'fieldGroupBRef' is set. They should both be set or both unset")}
	}

	isBidi := !isGroupAEmpty
	for _, of := range ct.OutputFields {
		if of.SplitAB && !isBidi {
			return conntrackInvalidError{splitABWithNoBidi: true,
				msg: fmt.Errorf("output field %q has splitAB=true although bidirection is not enabled (fieldGroupARef is empty)", of.Name)}
		}
		if !isOperationValid(of.Operation) {
			return conntrackInvalidError{unknownOperation: true,
				msg: fmt.Errorf("unknown operation %q in output field %q", of.Operation, of.Name)}
		}
	}

	outputFieldNames := map[string]struct{}{}
	for _, of := range ct.OutputFields {
		if of.SplitAB {
			name := of.Name + "_AB"
			if unique := addToSet(outputFieldNames, name); !unique {
				return conntrackInvalidError{duplicateOutputFieldNames: true,
					msg: fmt.Errorf("duplicate outputField %q", name)}
			}
			name = of.Name + "_BA"
			if unique := addToSet(outputFieldNames, name); !unique {
				return conntrackInvalidError{duplicateOutputFieldNames: true,
					msg: fmt.Errorf("duplicate outputField %q", name)}
			}
		} else {
			name := of.Name
			if unique := addToSet(outputFieldNames, name); !unique {
				return conntrackInvalidError{duplicateOutputFieldNames: true,
					msg: fmt.Errorf("duplicate outputField %q", name)}
			}
		}
	}

	fieldGroups := map[string]struct{}{}
	for _, fg := range ct.KeyDefinition.FieldGroups {
		name := fg.Name
		if unique := addToSet(fieldGroups, name); !unique {
			return conntrackInvalidError{duplicateFieldGroup: true,
				msg: fmt.Errorf("duplicate fieldGroup %q", name)}
		}
	}

	if _, found := fieldGroups[ct.KeyDefinition.Hash.FieldGroupARef]; !isGroupAEmpty && !found {
		return conntrackInvalidError{undefinedFieldGroupARef: true,
			msg: fmt.Errorf("undefined fieldGroupARef %q", ct.KeyDefinition.Hash.FieldGroupARef)}
	}

	if _, found := fieldGroups[ct.KeyDefinition.Hash.FieldGroupBRef]; !isGroupBEmpty && !found {
		return conntrackInvalidError{undefinedFieldGroupBRef: true,
			msg: fmt.Errorf("undefined fieldGroupBRef %q", ct.KeyDefinition.Hash.FieldGroupBRef)}
	}

	for _, fieldGroupRef := range ct.KeyDefinition.Hash.FieldGroupRefs {
		if _, found := fieldGroups[fieldGroupRef]; !found {
			return conntrackInvalidError{undefinedFieldGroupRef: true,
				msg: fmt.Errorf("undefined fieldGroup %q", fieldGroupRef)}
		}
	}

	for _, ort := range ct.OutputRecordTypes {
		if !isOutputRecordTypeValid(ort) {
			return conntrackInvalidError{unknownOutputRecord: true,
				msg: fmt.Errorf("undefined output record type %q", ort)}
		}
	}
	return nil
}

// addToSet adds an item to a set and returns true if it's a new item. Otherwise, it returns false.
func addToSet(set map[string]struct{}, item string) bool {
	if _, found := set[item]; found {
		return false
	}
	set[item] = struct{}{}
	return true
}

func isOperationValid(value string) bool {
	valid := true
	switch value {
	case ConnTrackOperationName("Sum"):
	case ConnTrackOperationName("Count"):
	case ConnTrackOperationName("Min"):
	case ConnTrackOperationName("Max"):
	default:
		valid = false
	}
	return valid
}

func isOutputRecordTypeValid(value string) bool {
	valid := true
	switch value {
	case ConnTrackOutputRecordTypeName("NewConnection"):
	case ConnTrackOutputRecordTypeName("EndConnection"):
	case ConnTrackOutputRecordTypeName("FlowLog"):
	default:
		valid = false
	}
	return valid
}

type conntrackInvalidError struct {
	msg                       error
	fieldGroupABOnlyOneIsSet  bool
	splitABWithNoBidi         bool
	unknownOperation          bool
	duplicateFieldGroup       bool
	duplicateOutputFieldNames bool
	undefinedFieldGroupARef   bool
	undefinedFieldGroupBRef   bool
	undefinedFieldGroupRef    bool
	unknownOutputRecord       bool
}

func (err conntrackInvalidError) Error() string {
	if err.msg != nil {
		return err.msg.Error()
	}
	return ""
}

// Is makes 2 conntrackInvalidError objects equal if all their fields except for `msg` are the equal.
// This is useful in the tests where we don't want to repeat the error message.
// Is() is invoked by errors.Is() which is invoked by require.ErrorIs().
func (err conntrackInvalidError) Is(target error) bool {
	err.msg = nil
	return err == target
}
