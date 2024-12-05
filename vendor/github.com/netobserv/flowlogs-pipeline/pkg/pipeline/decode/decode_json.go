/*
 * Copyright (C) 2021 IBM, Inc.
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

package decode

import (
	"encoding/json"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

//nolint:revive
type DecodeJSON struct {
}

// Decode decodes input strings to a list of flow entries
func (c *DecodeJSON) Decode(line []byte) (config.GenericMap, error) {

	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debugf("decodeJSON: line = %v", string(line))
	}
	var decodedLine map[string]interface{}
	if err := json.Unmarshal(line, &decodedLine); err != nil {
		return nil, err
	}
	decodedLine2 := make(config.GenericMap, len(decodedLine))
	// flows directly ingested by flp-transformer won't have this field, so we need to add it
	// here. If the received line already contains the field, it will be overridden later
	decodedLine2["TimeReceived"] = time.Now().Unix()
	for k, v := range decodedLine {
		if v == nil {
			continue
		}
		decodedLine2[k] = v
	}
	return decodedLine2, nil
}

// NewDecodeJSON create a new decode
func NewDecodeJSON() (Decoder, error) {
	log.Debugf("entering NewDecodeJSON")
	return &DecodeJSON{}, nil
}
