/*
 * Copyright (C) 2021 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	 http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package ingest

import (
	"fmt"
	"sync/atomic"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	log "github.com/sirupsen/logrus"
)

type Fake struct {
	Count    int64
	params   config.Ingest
	In       chan config.GenericMap
	exitChan <-chan struct{}
}

// Ingest reads records from an input channel and writes them as-is to the output channel
func (inf *Fake) Ingest(out chan<- config.GenericMap) {
	for {
		select {
		case <-inf.exitChan:
			log.Debugf("exiting IngestFake because of signal")
			return
		case records := <-inf.In:
			out <- records
			atomic.AddInt64(&inf.Count, 1)
		}
	}
}

// NewIngestFake creates a new ingester
func NewIngestFake(params config.StageParam) (Ingester, error) {
	log.Debugf("entering NewIngestFake")
	if params.Ingest == nil {
		return nil, fmt.Errorf("ingest not specified")
	}

	return &Fake{
		params:   *params.Ingest,
		In:       make(chan config.GenericMap),
		exitChan: utils.ExitChannel(),
	}, nil
}
