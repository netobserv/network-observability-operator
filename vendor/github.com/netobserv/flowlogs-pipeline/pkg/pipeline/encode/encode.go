/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package encode

import (
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

type encodeNone struct {
	prevRecord config.GenericMap
}

type Encoder interface {
	Encode(in config.GenericMap)
	Update(config.StageParam)
}

// Encode encodes a flow before being stored
func (t *encodeNone) Encode(in config.GenericMap) {
	t.prevRecord = in
}

func (t *encodeNone) Update(_ config.StageParam) {
	log.Warn("Encode None, update not supported")
}

// NewEncodeNone create a new encode
func NewEncodeNone() (Encoder, error) {
	log.Debugf("entering NewEncodeNone")
	return &encodeNone{}, nil
}
