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
)

//go:generate mockgen -copyright_file ../../license_templates/license_header.raw.txt -destination=testing/mock_set.go -package=testing github.com/vmware/go-ipfix/pkg/entities Set

const (
	// TemplateRefreshTimeOut is the template refresh time out for exporting process
	TemplateRefreshTimeOut uint32 = 1800
	// TemplateTTL is the template time to live for collecting process
	TemplateTTL = TemplateRefreshTimeOut * 3
	// TemplateSetID is the setID for template record
	TemplateSetID uint16 = 2
	SetHeaderLen  int    = 4
)

type ContentType uint8

const (
	Template ContentType = iota
	Data
	// Add OptionsTemplate too when it is supported
	Undefined = 255
)

type Set interface {
	PrepareSet(setType ContentType, templateID uint16) error
	ResetSet()
	GetHeaderBuffer() []byte
	GetSetLength() int
	GetSetType() ContentType
	UpdateLenInHeader()
	AddRecord(elements []InfoElementWithValue, templateID uint16) error
	AddRecordWithExtraElements(elements []InfoElementWithValue, numExtraElements int, templateID uint16) error
	// Unlike AddRecord, AddRecordV2 uses the elements slice directly, instead of creating a new
	// one. This can result in fewer memory allocations. The caller should not modify the
	// contents of the slice after calling AddRecordV2.
	AddRecordV2(elements []InfoElementWithValue, templateID uint16) error
	GetRecords() []Record
	GetNumberOfRecords() uint32
}

type set struct {
	headerBuffer []byte
	setType      ContentType
	records      []Record
	isDecoding   bool
	length       int
}

func NewSet(isDecoding bool) Set {
	if isDecoding {
		return &set{
			records:    make([]Record, 0),
			isDecoding: isDecoding,
		}
	} else {
		return &set{
			headerBuffer: make([]byte, SetHeaderLen),
			records:      make([]Record, 0),
			isDecoding:   isDecoding,
			length:       SetHeaderLen,
		}
	}
}

func (s *set) PrepareSet(setType ContentType, templateID uint16) error {
	if setType == Undefined {
		return fmt.Errorf("set type is not properly defined")
	} else {
		s.setType = setType
	}
	if !s.isDecoding {
		// Create the set header and append it when encoding
		s.createHeader(s.setType, templateID)
	}
	return nil
}

func (s *set) ResetSet() {
	if !s.isDecoding {
		s.headerBuffer = nil
		s.headerBuffer = make([]byte, SetHeaderLen)
		s.length = SetHeaderLen
	}
	s.setType = Undefined
	s.records = nil
	s.records = make([]Record, 0)
}

func (s *set) GetHeaderBuffer() []byte {
	return s.headerBuffer
}

func (s *set) GetSetLength() int {
	return s.length
}

func (s *set) GetSetType() ContentType {
	return s.setType
}

func (s *set) UpdateLenInHeader() {
	// TODO:Add padding to the length when multiple sets are sent in IPFIX message
	if !s.isDecoding {
		// Add length to the set header
		binary.BigEndian.PutUint16(s.headerBuffer[2:4], uint16(s.length))
	}
}

func (s *set) AddRecord(elements []InfoElementWithValue, templateID uint16) error {
	return s.AddRecordWithExtraElements(elements, 0, templateID)
}

func (s *set) AddRecordWithExtraElements(elements []InfoElementWithValue, numExtraElements int, templateID uint16) error {
	var record Record
	if s.setType == Data {
		record = NewDataRecord(templateID, len(elements), numExtraElements, s.isDecoding)
	} else if s.setType == Template {
		record = NewTemplateRecord(templateID, len(elements), s.isDecoding)
		err := record.PrepareRecord()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("set type is not supported")
	}
	for i := range elements {
		err := record.AddInfoElement(elements[i])
		if err != nil {
			return err
		}
	}
	s.records = append(s.records, record)
	s.length += record.GetRecordLength()
	return nil
}

func (s *set) AddRecordV2(elements []InfoElementWithValue, templateID uint16) error {
	var record Record
	if s.setType == Data {
		record = NewDataRecordFromElements(templateID, elements, s.isDecoding)
	} else if s.setType == Template {
		record = NewTemplateRecordFromElements(templateID, elements, s.isDecoding)
		err := record.PrepareRecord()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("set type is not supported")
	}
	s.records = append(s.records, record)
	s.length += record.GetRecordLength()
	return nil
}

func (s *set) GetRecords() []Record {
	return s.records
}

func (s *set) GetNumberOfRecords() uint32 {
	return uint32(len(s.records))
}

func (s *set) createHeader(setType ContentType, templateID uint16) {
	if setType == Template {
		binary.BigEndian.PutUint16(s.headerBuffer[0:2], TemplateSetID)
	} else if setType == Data {
		binary.BigEndian.PutUint16(s.headerBuffer[0:2], templateID)
	}
}
