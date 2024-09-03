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

package config

import (
	"syscall"

	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
)

type GenericMap map[string]interface{}

const (
	duplicateFieldName = "Duplicate"
	protoFieldName     = "Proto"
)

// Copy will create a flat copy of GenericMap
func (m GenericMap) Copy() GenericMap {
	result := make(GenericMap, len(m))

	for k, v := range m {
		result[k] = v
	}

	return result
}

func (m GenericMap) IsDuplicate() bool {
	if duplicate, hasKey := m[duplicateFieldName]; hasKey {
		if isDuplicate, err := utils.ConvertToBool(duplicate); err == nil {
			return isDuplicate
		}
	}
	return false
}

func (m GenericMap) IsValidProtocol() bool {
	if _, ok := m[protoFieldName]; ok {
		return true
	}
	return false
}

func (m GenericMap) IsTransportProtocol() bool {
	if v, ok := m[protoFieldName]; ok {
		if proto, err := utils.ConvertToFloat64(v); err == nil {
			if proto == float64(syscall.IPPROTO_TCP) || proto == float64(syscall.IPPROTO_UDP) || proto == float64(syscall.IPPROTO_SCTP) {
				return true
			}
		}
	}
	return false
}

func (m GenericMap) LookupString(key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}
