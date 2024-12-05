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
	"fmt"
	"hash"
	"hash/fnv"
	"strconv"

	"github.com/benbjohnson/clock"
	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// direction indicates the direction of a flow log in a connection. It's used by aggregators to determine which split
// of the aggregator should be updated, xxx_AB or xxx_BA.
type direction uint8

const (
	dirNA direction = iota
	dirAB
	dirBA
)

type conntrackImpl struct {
	clock                            clock.Clock
	config                           *api.ConnTrack
	endpointAFields, endpointBFields []string
	hashProvider                     func() hash.Hash64
	connStore                        *connectionStore
	aggregators                      []aggregator
	shouldOutputFlowLogs             bool
	shouldOutputNewConnection        bool
	shouldOutputEndConnection        bool
	shouldOutputHeartbeats           bool
	metrics                          *metricsType
}

func (ct *conntrackImpl) filterFlowLog(fl config.GenericMap) bool {
	if !fl.IsValidProtocol() || !fl.IsTransportProtocol() {
		return true
	}
	return false
}

func (ct *conntrackImpl) Extract(flowLogs []config.GenericMap) []config.GenericMap {
	log.Debugf("entering Extract conntrack, in = %v", flowLogs)

	var outputRecords []config.GenericMap
	for _, fl := range flowLogs {
		if ct.filterFlowLog(fl) {
			ct.metrics.inputRecords.WithLabelValues("discarded").Inc()
			continue
		}
		computedHash, err := computeHash(fl, &ct.config.KeyDefinition, ct.hashProvider(), ct.metrics)
		if err != nil {
			log.Warningf("skipping flow log %v: %v", fl, err)
			ct.metrics.inputRecords.WithLabelValues("rejected").Inc()
			continue
		}

		if fl.IsDuplicate() {
			log.Debugf("skipping duplicated flow log %v", fl)
			ct.metrics.inputRecords.WithLabelValues("duplicate").Inc()
		} else {
			conn, exists, _ := ct.connStore.getConnection(computedHash.hashTotal)
			if !exists {
				if (ct.config.MaxConnectionsTracked > 0) && (ct.connStore.len() >= ct.config.MaxConnectionsTracked) {
					log.Warningf("too many connections; skipping flow log %v: ", fl)
					ct.metrics.inputRecords.WithLabelValues("discarded").Inc()
				} else {
					builder := newConnBuilder(ct.metrics)
					conn = builder.
						ShouldSwapAB(ct.config.TCPFlags.SwapAB && ct.containsTCPFlag(fl, SYNACKFlag)).
						Hash(computedHash).
						keysFrom(fl, &ct.config.KeyDefinition, ct.endpointAFields, ct.endpointBFields).
						Aggregators(ct.aggregators).
						Hash(computedHash).
						Build()
					ct.connStore.addConnection(computedHash.hashTotal, conn)
					ct.connStore.updateNextHeartbeatTime(computedHash.hashTotal)
					ct.updateConnection(conn, fl, computedHash, true)
					ct.metrics.inputRecords.WithLabelValues("newConnection").Inc()
					if ct.shouldOutputNewConnection {
						record := conn.toGenericMap()
						addHashField(record, computedHash.hashTotal)
						addTypeField(record, api.ConnTrackNewConnection)
						isFirst := conn.markReported()
						addIsFirstField(record, isFirst)
						outputRecords = append(outputRecords, record)
						ct.metrics.outputRecords.WithLabelValues("newConnection").Inc()
					}
				}
			} else {
				ct.updateConnection(conn, fl, computedHash, false)
				ct.metrics.inputRecords.WithLabelValues("update").Inc()
			}
		}

		if ct.shouldOutputFlowLogs {
			record := fl.Copy()
			addHashField(record, computedHash.hashTotal)
			addTypeField(record, api.ConnTrackFlowLog)
			outputRecords = append(outputRecords, record)
			ct.metrics.outputRecords.WithLabelValues("flowLog").Inc()
		}
	}

	endConnectionRecords := ct.popEndConnections()
	if ct.shouldOutputEndConnection {
		outputRecords = append(outputRecords, endConnectionRecords...)
		ct.metrics.outputRecords.WithLabelValues("endConnection").Add(float64(len(endConnectionRecords)))
	}

	if ct.shouldOutputHeartbeats {
		heartbeatRecords := ct.prepareHeartbeatRecords()
		outputRecords = append(outputRecords, heartbeatRecords...)
		ct.metrics.outputRecords.WithLabelValues("heartbeat").Add(float64(len(heartbeatRecords)))
	}

	return outputRecords
}

func (ct *conntrackImpl) popEndConnections() []config.GenericMap {
	connections := ct.connStore.popEndConnections()

	var outputRecords []config.GenericMap
	// Convert the connections to GenericMaps and add meta fields
	for _, conn := range connections {
		record := conn.toGenericMap()
		addHashField(record, conn.getHash().hashTotal)
		addTypeField(record, api.ConnTrackEndConnection)
		var isFirst bool
		if ct.shouldOutputEndConnection {
			isFirst = conn.markReported()
		}
		addIsFirstField(record, isFirst)
		outputRecords = append(outputRecords, record)
	}
	return outputRecords
}

func (ct *conntrackImpl) prepareHeartbeatRecords() []config.GenericMap {
	connections := ct.connStore.prepareHeartbeats()

	var outputRecords []config.GenericMap
	// Convert the connections to GenericMaps and add meta fields
	for _, conn := range connections {
		record := conn.toGenericMap()
		addHashField(record, conn.getHash().hashTotal)
		addTypeField(record, api.ConnTrackHeartbeat)
		var isFirst bool
		if ct.shouldOutputHeartbeats {
			isFirst = conn.markReported()
		}
		addIsFirstField(record, isFirst)
		outputRecords = append(outputRecords, record)
	}
	return outputRecords
}

func (ct *conntrackImpl) updateConnection(conn connection, flowLog config.GenericMap, flowLogHash totalHashType, isNew bool) {
	d := ct.getFlowLogDirection(conn, flowLogHash)
	for _, agg := range ct.aggregators {
		agg.update(conn, flowLog, d, isNew)
	}

	if ct.config.TCPFlags.DetectEndConnection && ct.containsTCPFlag(flowLog, FINFlag) {
		ct.metrics.tcpFlags.WithLabelValues("detectEndConnection").Inc()
		ct.connStore.setConnectionTerminating(flowLogHash.hashTotal)
	} else {
		ct.connStore.updateConnectionExpiryTime(flowLogHash.hashTotal)
	}
}

func (ct *conntrackImpl) containsTCPFlag(flowLog config.GenericMap, queryFlag uint32) bool {
	tcpFlagsRaw, ok := flowLog[ct.config.TCPFlags.FieldName]
	if ok {
		tcpFlags, err := utils.ConvertToUint32(tcpFlagsRaw)
		if err != nil {
			log.Warningf("cannot convert TCP flag %q to uint32: %v", tcpFlagsRaw, err)
			return false
		}
		containsFlag := (tcpFlags & queryFlag) == queryFlag
		if containsFlag {
			return true
		}
	}

	return false
}

func (ct *conntrackImpl) getFlowLogDirection(conn connection, flowLogHash totalHashType) direction {
	d := dirNA
	if ct.config.KeyDefinition.Hash.FieldGroupARef != "" {
		if conn.getHash().hashA == flowLogHash.hashA {
			// A -> B
			d = dirAB
		} else {
			// B -> A
			d = dirBA
		}
	}
	return d
}

// NewConnectionTrack creates a new connection track instance
func NewConnectionTrack(opMetrics *operational.Metrics, params config.StageParam, clock clock.Clock) (extract.Extractor, error) {
	cfg := params.Extract.ConnTrack
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("ConnectionTrack config is invalid: %w", err)
	}

	metrics := newMetrics(opMetrics)

	var aggregators []aggregator
	for _, of := range cfg.OutputFields {
		agg, err := newAggregator(of, metrics)
		if err != nil {
			return nil, fmt.Errorf("error creating aggregator: %w", err)
		}
		aggregators = append(aggregators, agg)
	}
	shouldOutputFlowLogs := false
	shouldOutputNewConnection := false
	shouldOutputEndConnection := false
	shouldOutputHeartbeats := false
	for _, option := range cfg.OutputRecordTypes {
		switch option {
		case api.ConnTrackFlowLog:
			shouldOutputFlowLogs = true
		case api.ConnTrackNewConnection:
			shouldOutputNewConnection = true
		case api.ConnTrackEndConnection:
			shouldOutputEndConnection = true
		case api.ConnTrackHeartbeat:
			shouldOutputHeartbeats = true
		default:
			return nil, fmt.Errorf("unknown OutputRecordTypes: %v", option)
		}
	}

	endpointAFields, endpointBFields := cfg.GetABFields()
	conntrack := &conntrackImpl{
		clock:                     clock,
		connStore:                 newConnectionStore(cfg.Scheduling, metrics, clock.Now),
		config:                    cfg,
		endpointAFields:           endpointAFields,
		endpointBFields:           endpointBFields,
		hashProvider:              fnv.New64a,
		aggregators:               aggregators,
		shouldOutputFlowLogs:      shouldOutputFlowLogs,
		shouldOutputNewConnection: shouldOutputNewConnection,
		shouldOutputEndConnection: shouldOutputEndConnection,
		shouldOutputHeartbeats:    shouldOutputHeartbeats,
		metrics:                   metrics,
	}
	return conntrack, nil
}

func addHashField(record config.GenericMap, hashID uint64) {
	record[api.HashIDFieldName] = strconv.FormatUint(hashID, 16)
}

func addTypeField(record config.GenericMap, recordType api.ConnTrackOutputRecordTypeEnum) {
	record[api.RecordTypeFieldName] = recordType
}

func addIsFirstField(record config.GenericMap, isFirst bool) {
	record[api.IsFirstFieldName] = isFirst
}
