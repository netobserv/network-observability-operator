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

package write

import (
	"sync"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
)

type Writer interface {
	Write(in config.GenericMap)
}
type None struct {
	// synchronized access to avoid race conditions
	mt          sync.Mutex
	prevRecords []config.GenericMap
}

// Write writes entries
func (t *None) Write(in config.GenericMap) {
	logrus.Debugf("entering Write none, in = %v", in)
	t.mt.Lock()
	t.prevRecords = append(t.prevRecords, in)
	t.mt.Unlock()
}

func (t *None) PrevRecords() []config.GenericMap {
	t.mt.Lock()
	defer t.mt.Unlock()
	var copies []config.GenericMap
	for _, rec := range t.prevRecords {
		copies = append(copies, rec.Copy())
	}
	return copies
}

// NewWriteNone create a new write
func NewWriteNone() (Writer, error) {
	return &None{}, nil
}
