// Copyright 2020 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entities

import (
	"encoding/binary"
	"fmt"

	"k8s.io/klog/v2"
)

//go:generate mockgen -copyright_file ../../license_templates/license_header.raw.txt -destination=testing/mock_record.go -package=testing github.com/vmware/go-ipfix/pkg/entities Record

// This package contains encoding of fields in the record.
// Build the record here with local buffer and write to message buffer afterwards
// Instead should we write the field directly on to message instead of having a local buffer?
// To begin with, we will have local buffer in record.
// Have an interface and expose functions to user.

type Record interface {
	PrepareRecord() error
	AddInfoElement(element InfoElementWithValue) error
	// TODO: Functions for multiple elements as well.
	GetBuffer() []byte
	GetTemplateID() uint16
	GetFieldCount() uint16
	GetOrderedElementList() []InfoElementWithValue
	GetInfoElementWithValue(name string) (InfoElementWithValue, int, bool)
	GetRecordLength() int
	GetMinDataRecordLen() uint16
	GetElementMap() map[string]interface{}
}

type baseRecord struct {
	buffer             []byte
	fieldCount         uint16
	templateID         uint16
	orderedElementList []InfoElementWithValue
	isDecoding         bool
	len                int
	Record
}

type dataRecord struct {
	baseRecord
}

func NewDataRecord(id uint16, numElements, numExtraElements int, isDecoding bool) *dataRecord {
	return &dataRecord{
		baseRecord{
			fieldCount:         0,
			templateID:         id,
			isDecoding:         isDecoding,
			orderedElementList: make([]InfoElementWithValue, numElements, numElements+numExtraElements),
		},
	}
}

func NewDataRecordFromElements(id uint16, elements []InfoElementWithValue, isDecoding bool) *dataRecord {
	length := 0
	if !isDecoding {
		for idx := range elements {
			length += elements[idx].GetLength()
		}
	}
	return &dataRecord{
		baseRecord{
			fieldCount:         uint16(len(elements)),
			templateID:         id,
			isDecoding:         isDecoding,
			len:                length,
			orderedElementList: elements,
		},
	}
}

type templateRecord struct {
	baseRecord
	// Minimum data record length required to be sent for this template.
	// Elements with variable length are considered to be one byte.
	minDataRecLength uint16
	// index is used when adding elements to orderedElementList
	index int
}

func NewTemplateRecord(id uint16, numElements int, isDecoding bool) *templateRecord {
	return &templateRecord{
		baseRecord{
			buffer:             make([]byte, 4),
			fieldCount:         uint16(numElements),
			templateID:         id,
			isDecoding:         isDecoding,
			orderedElementList: make([]InfoElementWithValue, numElements, numElements),
		},
		0,
		0,
	}
}

func NewTemplateRecordFromElements(id uint16, elements []InfoElementWithValue, isDecoding bool) *templateRecord {
	r := &templateRecord{
		baseRecord{
			buffer:             make([]byte, 4),
			fieldCount:         uint16(len(elements)),
			templateID:         id,
			isDecoding:         isDecoding,
			orderedElementList: elements,
		},
		0,
		len(elements),
	}
	for idx := range elements {
		infoElement := elements[idx].GetInfoElement()
		r.addInfoElement(infoElement)
	}
	return r
}

func (b *baseRecord) GetTemplateID() uint16 {
	return b.templateID
}

func (b *baseRecord) GetFieldCount() uint16 {
	return b.fieldCount
}

func (b *baseRecord) GetOrderedElementList() []InfoElementWithValue {
	return b.orderedElementList
}

func (b *baseRecord) GetInfoElementWithValue(name string) (InfoElementWithValue, int, bool) {
	for i, element := range b.orderedElementList {
		if element.GetName() == name {
			return element, i, true
		}
	}
	return nil, 0, false
}

func (b *baseRecord) GetElementMap() map[string]interface{} {
	elements := make(map[string]interface{})
	orderedElements := b.GetOrderedElementList()
	for _, element := range orderedElements {
		switch element.GetDataType() {
		case Unsigned8:
			elements[element.GetName()] = element.GetUnsigned8Value()
		case Unsigned16:
			elements[element.GetName()] = element.GetUnsigned16Value()
		case Unsigned32:
			elements[element.GetName()] = element.GetUnsigned32Value()
		case Unsigned64:
			elements[element.GetName()] = element.GetUnsigned64Value()
		case Signed8:
			elements[element.GetName()] = element.GetSigned8Value()
		case Signed16:
			elements[element.GetName()] = element.GetSigned16Value()
		case Signed32:
			elements[element.GetName()] = element.GetSigned32Value()
		case Signed64:
			elements[element.GetName()] = element.GetSigned64Value()
		case Float32:
			elements[element.GetName()] = element.GetFloat32Value()
		case Float64:
			elements[element.GetName()] = element.GetFloat64Value()
		case Boolean:
			elements[element.GetName()] = element.GetBooleanValue()
		case DateTimeSeconds:
			elements[element.GetName()] = element.GetUnsigned32Value()
		case DateTimeMilliseconds:
			elements[element.GetName()] = element.GetUnsigned64Value()
		case DateTimeMicroseconds, DateTimeNanoseconds:
			err := fmt.Errorf("API does not support micro and nano seconds types yet")
			elements[element.GetName()] = err
		case MacAddress:
			elements[element.GetName()] = element.GetMacAddressValue()
		case Ipv4Address, Ipv6Address:
			elements[element.GetName()] = element.GetIPAddressValue()
		case String:
			elements[element.GetName()] = element.GetStringValue()
		default:
			err := fmt.Errorf("API supports only valid information elements with datatypes given in RFC7011")
			elements[element.GetName()] = err
		}
	}
	return elements
}

func (d *dataRecord) PrepareRecord() error {
	// We do not have to do anything if it is data record
	return nil
}

func (d *dataRecord) GetBuffer() []byte {
	if len(d.buffer) == d.len || d.isDecoding {
		return d.buffer
	}
	d.buffer = make([]byte, d.len)
	index := 0
	for _, element := range d.orderedElementList {
		err := encodeInfoElementValueToBuff(element, d.buffer, index)
		if err != nil {
			klog.Error(err)
		}
		index += element.GetLength()
	}
	return d.buffer
}

func (d *dataRecord) GetRecordLength() int {
	return d.len
}

func (d *dataRecord) AddInfoElement(element InfoElementWithValue) error {
	if !d.isDecoding {
		d.len = d.len + element.GetLength()
	}
	if len(d.orderedElementList) <= int(d.fieldCount) {
		d.orderedElementList = append(d.orderedElementList, element)
	} else {
		d.orderedElementList[d.fieldCount] = element
	}
	d.fieldCount++
	return nil
}

func (t *templateRecord) PrepareRecord() error {
	// Add Template Record Header
	binary.BigEndian.PutUint16(t.buffer[0:2], t.templateID)
	binary.BigEndian.PutUint16(t.buffer[2:4], t.fieldCount)
	return nil
}

func (t *templateRecord) addInfoElement(infoElement *InfoElement) {
	initialLength := len(t.buffer)
	// Add field specifier {elementID: uint16, elementLen: uint16}
	addBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(addBytes[0:2], infoElement.ElementId)
	binary.BigEndian.PutUint16(addBytes[2:4], infoElement.Len)
	t.buffer = append(t.buffer, addBytes...)
	if infoElement.EnterpriseId != 0 {
		// Set the MSB of elementID to 1 as per RFC7011
		t.buffer[initialLength] = t.buffer[initialLength] | 0x80
		addBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(addBytes, infoElement.EnterpriseId)
		t.buffer = append(t.buffer, addBytes...)
	}
	// Keep track of minimum data record length required for sanity check
	if infoElement.Len == VariableLength {
		t.minDataRecLength = t.minDataRecLength + 1
	} else {
		t.minDataRecLength = t.minDataRecLength + infoElement.Len
	}
}

func (t *templateRecord) AddInfoElement(element InfoElementWithValue) error {
	infoElement := element.GetInfoElement()
	// val could be used to specify smaller length than default? For now assert it to be nil
	if !element.IsValueEmpty() {
		return fmt.Errorf("template record cannot take element %v with non-empty value", infoElement.Name)
	}
	t.addInfoElement(infoElement)
	t.orderedElementList[t.index] = element
	t.index++
	return nil
}

func (t *templateRecord) GetBuffer() []byte {
	return t.buffer
}

func (t *templateRecord) GetRecordLength() int {
	return len(t.buffer)
}

func (t *templateRecord) GetMinDataRecordLen() uint16 {
	return t.minDataRecLength
}
