/*
 * Copyright (C) 2023 IBM, Inc.
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
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type ingestSynthetic struct {
	params            api.IngestSynthetic
	exitChan          <-chan struct{}
	flowLogsProcessed prometheus.Counter
}

const (
	defaultConnections    = 100
	defaultBatchLen       = 10
	defaultFlowLogsPerMin = 2000
)

var (
	flowLogsProcessed = operational.DefineMetric(
		"ingest_synthetic_flows_processed",
		"Number of flow logs processed",
		operational.TypeCounter,
		"stage",
	)
)

// Ingest generates flow logs according to provided parameters
func (ingestS *ingestSynthetic) Ingest(out chan<- config.GenericMap) {
	log.Debugf("entering IngestSynthetic Ingest, params = %v", ingestS.params)
	// get a list of flow log entries, one per desired connection
	// these flow logs will be sent again and again to simulate ongoing traffic on those connections
	flowLogs := utils.GenerateConnectionFlowEntries(ingestS.params.Connections)
	nLogs := len(flowLogs)
	next := 0

	// compute time interval between batches; divide BatchMaxLen by FlowLogsPerMin and adjust the types
	ticker := time.NewTicker(time.Duration(int(time.Minute*time.Duration(ingestS.params.BatchMaxLen)) / ingestS.params.FlowLogsPerMin))

	// loop forever
	for {
		select {
		case <-ingestS.exitChan:
			log.Debugf("exiting IngestSynthetic because of signal")
			return
		case <-ticker.C:
			log.Debugf("sending a batch of %d flow logs from index %d", ingestS.params.BatchMaxLen, next)
			for i := 0; i < ingestS.params.BatchMaxLen; i++ {
				out <- flowLogs[next]
				ingestS.flowLogsProcessed.Inc()
				next++
				if next >= nLogs {
					next = 0
				}
			}
		}
	}
}

// NewIngestSynthetic create a new ingester
func NewIngestSynthetic(opMetrics *operational.Metrics, params config.StageParam) (Ingester, error) {
	log.Debugf("entering NewIngestSynthetic")
	confIngestSynthetic := api.IngestSynthetic{}
	if params.Ingest != nil && params.Ingest.Synthetic != nil {
		confIngestSynthetic = *params.Ingest.Synthetic
	}
	if confIngestSynthetic.Connections == 0 {
		confIngestSynthetic.Connections = defaultConnections
	}
	if confIngestSynthetic.FlowLogsPerMin == 0 {
		confIngestSynthetic.FlowLogsPerMin = defaultFlowLogsPerMin
	}
	if confIngestSynthetic.BatchMaxLen == 0 {
		confIngestSynthetic.BatchMaxLen = defaultBatchLen
	}
	log.Debugf("params = %v", confIngestSynthetic)

	return &ingestSynthetic{
		params:            confIngestSynthetic,
		exitChan:          utils.ExitChannel(),
		flowLogsProcessed: opMetrics.NewCounter(&flowLogsProcessed, params.Name),
	}, nil
}
