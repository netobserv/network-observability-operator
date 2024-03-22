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
	HashIDFieldName     = "_HashId"
	RecordTypeFieldName = "_RecordType"
	IsFirstFieldName    = "_IsFirst"
)

type ConnTrack struct {
	KeyDefinition         KeyDefinition                   `yaml:"keyDefinition,omitempty" json:"keyDefinition,omitempty" doc:"fields that are used to identify the connection"`
	OutputRecordTypes     []ConnTrackOutputRecordTypeEnum `yaml:"outputRecordTypes,omitempty" json:"outputRecordTypes,omitempty" doc:"(enum) output record types to emit"`
	OutputFields          []OutputField                   `yaml:"outputFields,omitempty" json:"outputFields,omitempty" doc:"list of output fields"`
	Scheduling            []ConnTrackSchedulingGroup      `yaml:"scheduling,omitempty" json:"scheduling,omitempty" doc:"list of timeouts and intervals to apply per selector"`
	MaxConnectionsTracked int                             `yaml:"maxConnectionsTracked,omitempty" json:"maxConnectionsTracked,omitempty" doc:"maximum number of connections we keep in our cache (0 means no limit)"`
	TCPFlags              ConnTrackTCPFlags               `yaml:"tcpFlags,omitempty" json:"tcpFlags,omitempty" doc:"settings for handling TCP flags"`
}

type ConnTrackOutputRecordTypeEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	ConnTrackNewConnection ConnTrackOutputRecordTypeEnum = "newConnection" // New connection
	ConnTrackEndConnection ConnTrackOutputRecordTypeEnum = "endConnection" // End connection
	ConnTrackHeartbeat     ConnTrackOutputRecordTypeEnum = "heartbeat"     // Heartbeat
	ConnTrackFlowLog       ConnTrackOutputRecordTypeEnum = "flowLog"       // Flow log
)

type KeyDefinition struct {
	FieldGroups []FieldGroup  `yaml:"fieldGroups,omitempty" json:"fieldGroups,omitempty" doc:"list of field group definitions"`
	Hash        ConnTrackHash `yaml:"hash,omitempty" json:"hash,omitempty" doc:"how to build the connection hash"`
}

type FieldGroup struct {
	Name   string   `yaml:"name,omitempty" json:"name,omitempty" doc:"field group name"`
	Fields []string `yaml:"fields" json:"fields" doc:"list of fields in the group"`
}

// ConnTrackHash determines how to compute the connection hash.
// A and B are treated as the endpoints of the connection.
// When FieldGroupARef and FieldGroupBRef are set, the hash is computed in a way
// that flow logs from A to B will have the same hash as flow logs from B to A.
// When they are not set, a different hash will be computed for A->B and B->A,
// and they are tracked as different connections.
type ConnTrackHash struct {
	FieldGroupRefs []string `yaml:"fieldGroupRefs,omitempty" json:"fieldGroupRefs,omitempty" doc:"list of field group names to build the hash"`
	FieldGroupARef string   `yaml:"fieldGroupARef,omitempty" json:"fieldGroupARef,omitempty" doc:"field group name of endpoint A"`
	FieldGroupBRef string   `yaml:"fieldGroupBRef,omitempty" json:"fieldGroupBRef,omitempty" doc:"field group name of endpoint B"`
}

type OutputField struct {
	Name          string                 `yaml:"name,omitempty" json:"name,omitempty" doc:"output field name"`
	Operation     ConnTrackOperationEnum `yaml:"operation,omitempty" json:"operation,omitempty" doc:"(enum) aggregate operation on the field value"`
	SplitAB       bool                   `yaml:"splitAB,omitempty" json:"splitAB,omitempty" doc:"When true, 2 output fields will be created. One for A->B and one for B->A flows."`
	Input         string                 `yaml:"input,omitempty" json:"input,omitempty" doc:"The input field to base the operation on. When omitted, 'name' is used"`
	ReportMissing bool                   `yaml:"reportMissing,omitempty" json:"reportMissing,omitempty" doc:"When true, missing input will produce MissingFieldError metric and error logs"`
}

type ConnTrackOperationEnum string

const (
	// For doc generation, enum definitions must match format `Constant Type = "value" // doc`
	ConnTrackSum   ConnTrackOperationEnum = "sum"   // sum
	ConnTrackCount ConnTrackOperationEnum = "count" // count
	ConnTrackMin   ConnTrackOperationEnum = "min"   // min
	ConnTrackMax   ConnTrackOperationEnum = "max"   // max
	ConnTrackFirst ConnTrackOperationEnum = "first" // first
	ConnTrackLast  ConnTrackOperationEnum = "last"  // last
)

type ConnTrackSchedulingGroup struct {
	Selector             map[string]interface{} `yaml:"selector,omitempty" json:"selector,omitempty" doc:"key-value map to match against connection fields to apply this scheduling"`
	EndConnectionTimeout Duration               `yaml:"endConnectionTimeout,omitempty" json:"endConnectionTimeout,omitempty" doc:"duration of time to wait from the last flow log to end a connection"`
	TerminatingTimeout   Duration               `yaml:"terminatingTimeout,omitempty" json:"terminatingTimeout,omitempty" doc:"duration of time to wait from detected FIN flag to end a connection"`
	HeartbeatInterval    Duration               `yaml:"heartbeatInterval,omitempty" json:"heartbeatInterval,omitempty" doc:"duration of time to wait between heartbeat reports of a connection"`
}

type ConnTrackTCPFlags struct {
	FieldName           string `yaml:"fieldName,omitempty" json:"fieldName,omitempty" doc:"name of the field containing TCP flags"`
	DetectEndConnection bool   `yaml:"detectEndConnection,omitempty" json:"detectEndConnection,omitempty" doc:"detect end connections by FIN flag"`
	SwapAB              bool   `yaml:"swapAB,omitempty" json:"swapAB,omitempty" doc:"swap source and destination when the first flowlog contains the SYN_ACK flag"`
}

//nolint:cyclop
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
		if !isOperationValid(of.Operation, of.SplitAB) {
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

	definedKeys := map[string]struct{}{}
	for _, fg := range ct.KeyDefinition.FieldGroups {
		for _, k := range fg.Fields {
			addToSet(definedKeys, k)
		}
	}
	for i, group := range ct.Scheduling {
		for k := range group.Selector {
			if _, found := definedKeys[k]; !found {
				return conntrackInvalidError{undefinedSelectorKey: true,
					msg: fmt.Errorf("selector key %q in scheduling group %v is not defined in the keys", k, i)}
			}
		}
	}

	numOfDefault := 0
	for i, group := range ct.Scheduling {
		isDefaultSelector := (len(group.Selector) == 0)
		isLastGroup := (i == len(ct.Scheduling)-1)
		if isDefaultSelector {
			numOfDefault++
		}
		if isDefaultSelector && !isLastGroup {
			return conntrackInvalidError{defaultGroupAndNotLast: true,
				msg: fmt.Errorf("scheduling group %v has a default selector but is not the last scheduling group", i)}
		}
	}

	if numOfDefault != 1 {
		return conntrackInvalidError{exactlyOneDefaultSelector: true,
			msg: fmt.Errorf("found %v default selectors. There should be exactly 1", numOfDefault)}
	}

	if len(ct.TCPFlags.FieldName) == 0 && (ct.TCPFlags.DetectEndConnection || ct.TCPFlags.SwapAB) {
		return conntrackInvalidError{emptyTCPFlagsField: true,
			msg: fmt.Errorf("TCPFlags.FieldName is empty although DetectEndConnection or SwapAB are enabled")}
	}
	if ct.TCPFlags.SwapAB && !isBidi {
		return conntrackInvalidError{swapABWithNoBidi: true,
			msg: fmt.Errorf("SwapAB is enabled although bidirection is not enabled (fieldGroupARef is empty)")}
	}

	fieldsA, fieldsB := ct.GetABFields()
	if len(fieldsA) != len(fieldsB) {
		return conntrackInvalidError{mismatchABFieldsCount: true,
			msg: fmt.Errorf("mismatch between the field count of fieldGroupARef %v and fieldGroupBRef %v", len(fieldsA), len(fieldsB))}
	}

	return nil
}

func (ct *ConnTrack) GetABFields() ([]string, []string) {
	endpointAFieldGroupName := ct.KeyDefinition.Hash.FieldGroupARef
	endpointBFieldGroupName := ct.KeyDefinition.Hash.FieldGroupBRef
	var endpointAFields []string
	var endpointBFields []string
	for _, fg := range ct.KeyDefinition.FieldGroups {
		if fg.Name == endpointAFieldGroupName {
			endpointAFields = fg.Fields
		}
		if fg.Name == endpointBFieldGroupName {
			endpointBFields = fg.Fields
		}
	}
	return endpointAFields, endpointBFields
}

// addToSet adds an item to a set and returns true if it's a new item. Otherwise, it returns false.
func addToSet(set map[string]struct{}, item string) bool {
	if _, found := set[item]; found {
		return false
	}
	set[item] = struct{}{}
	return true
}

func isOperationValid(value ConnTrackOperationEnum, splitAB bool) bool {
	valid := true
	switch value {
	case ConnTrackSum:
	case ConnTrackCount:
	case ConnTrackMin:
	case ConnTrackMax:
	case ConnTrackFirst, ConnTrackLast:
		valid = !splitAB
	default:
		valid = false
	}
	return valid
}

func isOutputRecordTypeValid(value ConnTrackOutputRecordTypeEnum) bool {
	valid := true
	switch value {
	case ConnTrackNewConnection:
	case ConnTrackEndConnection:
	case ConnTrackHeartbeat:
	case ConnTrackFlowLog:
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
	undefinedSelectorKey      bool
	defaultGroupAndNotLast    bool
	exactlyOneDefaultSelector bool
	swapABWithNoBidi          bool
	emptyTCPFlagsField        bool
	mismatchABFieldsCount     bool
}

func (err conntrackInvalidError) Error() string {
	if err.msg != nil {
		return err.msg.Error()
	}
	return ""
}

// Is makes 2 conntrackInvalidError objects equal if all their fields except for `msg` are equal.
// This is useful in the tests where we don't want to repeat the error message.
// Is() is invoked by errors.Is() which is invoked by require.ErrorIs().
func (err conntrackInvalidError) Is(target error) bool {
	err.msg = nil
	return err == target
}
