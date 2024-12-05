// Copyright 2023 VMware, Inc.
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

package exporter

import (
	"fmt"
	"time"

	"github.com/vmware/go-ipfix/pkg/entities"
)

func CreateIPFIXMsg(set entities.Set, obsDomainID uint32, seqNumber uint32, exportTime time.Time) ([]byte, error) {
	// Create a new message and use it to send the set.
	msg := entities.NewMessage(false)

	// Check if message is exceeding the limit after adding the set. Include message
	// header length too.
	msgLen := entities.MsgHeaderLength + set.GetSetLength()
	if msgLen > entities.MaxSocketMsgSize {
		// This is applicable for both TCP and UDP sockets.
		return nil, fmt.Errorf("message size exceeds max socket buffer size")
	}

	// Set the fields in the message header.
	// IPFIX version number is 10.
	// https://www.iana.org/assignments/ipfix/ipfix.xhtml#ipfix-version-numbers
	msg.SetVersion(10)
	msg.SetObsDomainID(obsDomainID)
	msg.SetMessageLen(uint16(msgLen))
	msg.SetExportTime(uint32(exportTime.Unix()))
	msg.SetSequenceNum(seqNumber)

	bytesSlice := make([]byte, msgLen)
	copy(bytesSlice[:entities.MsgHeaderLength], msg.GetMsgHeader())
	copy(bytesSlice[entities.MsgHeaderLength:entities.MsgHeaderLength+entities.SetHeaderLen], set.GetHeaderBuffer())
	index := entities.MsgHeaderLength + entities.SetHeaderLen
	for _, record := range set.GetRecords() {
		len := record.GetRecordLength()
		copy(bytesSlice[index:index+len], record.GetBuffer())
		index += len
	}

	return bytesSlice, nil
}
