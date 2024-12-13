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

package conntrack

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

// TODO: what's a better name for this struct?
type totalHashType struct {
	hashA     uint64
	hashB     uint64
	hashTotal uint64
}

// computeHash computes the hash of a flow log according to keyDefinition.
// Two flow logs will have the same hash if they belong to the same connection.
func computeHash(flowLog config.GenericMap, keyDefinition *api.KeyDefinition, hasher hash.Hash64, metrics *metricsType) (totalHashType, error) {
	fieldGroup2hash := make(map[string]uint64)

	// Compute the hash of each field group
	for _, fg := range keyDefinition.FieldGroups {
		h, err := computeHashFields(flowLog, fg.Fields, hasher, metrics)
		if err != nil {
			return totalHashType{}, fmt.Errorf("compute hash: %w", err)
		}
		fieldGroup2hash[fg.Name] = h
	}

	// Compute the total hash
	th := totalHashType{}
	hasher.Reset()
	for _, fgName := range keyDefinition.Hash.FieldGroupRefs {
		hasher.Write(uint64ToBytes(fieldGroup2hash[fgName]))
	}
	if keyDefinition.Hash.FieldGroupARef != "" {
		th.hashA = fieldGroup2hash[keyDefinition.Hash.FieldGroupARef]
		th.hashB = fieldGroup2hash[keyDefinition.Hash.FieldGroupBRef]
		// Determine order between A's and B's hash to get the same hash for both flow logs from A to B and from B to A.
		if th.hashA < th.hashB {
			hasher.Write(uint64ToBytes(th.hashA))
			hasher.Write(uint64ToBytes(th.hashB))
		} else {
			hasher.Write(uint64ToBytes(th.hashB))
			hasher.Write(uint64ToBytes(th.hashA))
		}
	}
	th.hashTotal = hasher.Sum64()
	return th, nil
}

func computeHashFields(flowLog config.GenericMap, fieldNames []string, hasher hash.Hash64, metrics *metricsType) (uint64, error) {
	hasher.Reset()
	for _, fn := range fieldNames {
		f, ok := flowLog[fn]
		if !ok {
			log.Warningf("Missing field %v", fn)
			if metrics != nil {
				metrics.hashErrors.WithLabelValues("MissingFieldError", fn).Inc()
			}
			continue
		}
		bytes, err := toBytes(f)
		if err != nil {
			return 0, err
		}
		hasher.Write(bytes)
	}
	return hasher.Sum64(), nil
}

func uint64ToBytes(data uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, data)
	return b
}

func toBytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	bytes := buf.Bytes()
	return bytes, nil
}
