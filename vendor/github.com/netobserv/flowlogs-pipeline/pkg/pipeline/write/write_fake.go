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

package write

import (
	"sync"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
)

type Fake struct {
	// access is locked and copied to avoid race condition errors during tests
	mt         sync.Mutex
	allRecords []config.GenericMap
}

// Write stores in memory all records.
func (w *Fake) Write(in config.GenericMap) {
	logrus.Trace("entering writeFake Write")
	w.mt.Lock()
	w.allRecords = append(w.allRecords, in.Copy())
	w.mt.Unlock()
}

func (w *Fake) AllRecords() []config.GenericMap {
	w.mt.Lock()
	defer w.mt.Unlock()
	var copies []config.GenericMap
	for _, r := range w.allRecords {
		copies = append(copies, r.Copy())
	}
	return copies
}

// NewWriteFake creates a new write.
func NewWriteFake(_ config.StageParam) (Writer, error) {
	logrus.Debugf("entering NewWriteFake")
	w := &Fake{}
	return w, nil
}
