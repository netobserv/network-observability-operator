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
	"sort"
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	log "github.com/sirupsen/logrus"
)

const (
	expiryOrder            = utils.OrderID("expiryOrder")
	nextHeartbeatTimeOrder = utils.OrderID("nextHeartbeatTimeOrder")
	activeLabel            = "active"
	terminatingLabel       = "terminating"
)

// connectionStore provides means to manage the connections such as retrieving a connection by its hash and organizing
// them in groups sorted by expiry time and next report time.
// This allows efficient retrieval and removal of connections.
type connectionStore struct {
	groups          []*groupType
	hashID2groupIdx map[uint64]int
	metrics         *metricsType
	now             func() time.Time
}

type groupType struct {
	scheduling api.ConnTrackSchedulingGroup
	// active connections
	activeMom *utils.MultiOrderedMap
	// connections that detected EndConnection from TCP FIN flag. These will not trigger updates anymore until pop
	// check expireConnection func
	terminatingMom *utils.MultiOrderedMap
	labelValue     string
}

func (cs *connectionStore) getGroupIdx(conn connection) (groupIdx int) {
	for i, group := range cs.groups {
		if conn.isMatchSelector(group.scheduling.Selector) {
			// connection belongs to scheduling group i
			return i
		}
	}
	// Shouldn't get here since the last scheduling group should have a selector that matches any connection.
	log.Errorf("BUG. connection with hash %x doesn't match any selector", conn.getHash().hashTotal)
	lastGroupIdx := len(cs.groups) - 1
	return lastGroupIdx
}

func (cs *connectionStore) addConnection(hashID uint64, conn connection) {
	groupIdx := cs.getGroupIdx(conn)
	mom := cs.groups[groupIdx].activeMom

	err := mom.AddRecord(utils.Key(hashID), conn)
	if err != nil {
		log.Errorf("BUG. connection with hash %x already exists in store. %v", hashID, conn)
	}
	cs.hashID2groupIdx[hashID] = groupIdx

	groupLabel := cs.groups[groupIdx].labelValue
	activeLen := cs.groups[groupIdx].activeMom.Len()
	cs.metrics.connStoreLength.WithLabelValues(groupLabel, activeLabel).Set(float64(activeLen))
}

func (cs *connectionStore) getConnection(hashID uint64) (connection, bool, bool) {
	groupIdx, found := cs.hashID2groupIdx[hashID]
	if !found {
		return nil, false, false
	}
	mom := cs.groups[groupIdx].activeMom

	// get connection from active map
	isRunning := true
	record, ok := mom.GetRecord(utils.Key(hashID))
	if !ok {
		// fallback on terminating map if not found
		isRunning = false
		mom := cs.groups[groupIdx].terminatingMom
		record, ok = mom.GetRecord(utils.Key(hashID))
		if !ok {
			return nil, false, false
		}
	}
	conn := record.(connection)
	return conn, true, isRunning
}

func (cs *connectionStore) setConnectionTerminating(hashID uint64) {
	conn, ok, active := cs.getConnection(hashID)
	if !ok {
		log.Panicf("BUG. connection hash %x doesn't exist", hashID)
		return
	} else if !active {
		// connection is terminating
		return
	}
	groupIdx := cs.hashID2groupIdx[hashID]
	groupLabel := cs.groups[groupIdx].labelValue
	activeMom := cs.groups[groupIdx].activeMom
	terminatingMom := cs.groups[groupIdx].terminatingMom
	timeout := cs.groups[groupIdx].scheduling.TerminatingTimeout.Duration
	newExpiryTime := cs.now().Add(timeout)
	conn.setExpiryTime(newExpiryTime)
	// Remove connection from active map
	activeMom.RemoveRecord(utils.Key(hashID))
	activeLen := cs.groups[groupIdx].activeMom.Len()
	cs.metrics.connStoreLength.WithLabelValues(groupLabel, activeLabel).Set(float64(activeLen))
	// Add connection to terminating map
	err := terminatingMom.AddRecord(utils.Key(hashID), conn)
	if err != nil {
		log.Errorf("BUG. connection with hash %x already exists in store. %v", hashID, conn)
	}
	terminatingLen := cs.groups[groupIdx].terminatingMom.Len()
	cs.metrics.connStoreLength.WithLabelValues(groupLabel, terminatingLabel).Set(float64(terminatingLen))
}

func (cs *connectionStore) updateConnectionExpiryTime(hashID uint64) {
	conn, ok, active := cs.getConnection(hashID)
	if !ok {
		log.Panicf("BUG. connection hash %x doesn't exist", hashID)
		return
	} else if !active {
		// connection is terminating. expiry time can't be updated anymore
		return
	}
	groupIdx := cs.hashID2groupIdx[hashID]
	mom := cs.groups[groupIdx].activeMom
	timeout := cs.groups[groupIdx].scheduling.EndConnectionTimeout.Duration
	newExpiryTime := cs.now().Add(timeout)
	conn.setExpiryTime(newExpiryTime)
	// Move to the back of the list
	err := mom.MoveToBack(utils.Key(hashID), expiryOrder)
	if err != nil {
		log.Panicf("BUG. Can't update connection expiry time for hash %x: %v", hashID, err)
		return
	}
}

func (cs *connectionStore) updateNextHeartbeatTime(hashID uint64) {
	conn, ok, active := cs.getConnection(hashID)
	if !ok {
		log.Panicf("BUG. connection hash %x doesn't exist", hashID)
		return
	} else if !active {
		// connection is terminating. heartbeat are disabled
		return
	}
	groupIdx := cs.hashID2groupIdx[hashID]
	mom := cs.groups[groupIdx].activeMom
	timeout := cs.groups[groupIdx].scheduling.HeartbeatInterval.Duration
	newNextHeartbeatTime := cs.now().Add(timeout)
	conn.setNextHeartbeatTime(newNextHeartbeatTime)
	// Move to the back of the list
	err := mom.MoveToBack(utils.Key(hashID), nextHeartbeatTimeOrder)
	if err != nil {
		log.Panicf("BUG. Can't update next heartbeat time for hash %x: %v", hashID, err)
		return
	}
}

func (cs *connectionStore) popEndConnectionOfMap(mom *utils.MultiOrderedMap, group *groupType) []connection {
	var poppedConnections []connection

	mom.IterateFrontToBack(expiryOrder, func(r utils.Record) (shouldDelete, shouldStop bool) {
		conn := r.(connection)
		expiryTime := conn.getExpiryTime()
		if cs.now().After(expiryTime) {
			// The connection has expired. We want to pop it.
			poppedConnections = append(poppedConnections, conn)
			shouldDelete, shouldStop = true, false
			delete(cs.hashID2groupIdx, conn.getHash().hashTotal)
		} else {
			// No more expired connections
			shouldDelete, shouldStop = false, true
		}
		return
	})
	groupLabel := group.labelValue
	momLen := mom.Len()
	var phaseLabel string
	switch mom {
	case group.activeMom:
		phaseLabel = activeLabel
	case group.terminatingMom:
		phaseLabel = terminatingLabel
	}
	cs.metrics.connStoreLength.WithLabelValues(groupLabel, phaseLabel).Set(float64(momLen))

	return poppedConnections
}

func (cs *connectionStore) popEndConnections() []connection {
	// Iterate over the connections by scheduling groups.
	// In each scheduling group iterate over them by their expiry time from old to new.
	var poppedConnections []connection
	for _, group := range cs.groups {
		// Pop terminating connections first
		terminatedConnections := cs.popEndConnectionOfMap(group.terminatingMom, group)
		poppedConnections = append(poppedConnections, terminatedConnections...)
		cs.metrics.endConnections.WithLabelValues(group.labelValue, "FIN_flag").Add(float64(len(terminatedConnections)))

		// Pop active connections that expired without TCP flag
		timedoutConnections := cs.popEndConnectionOfMap(group.activeMom, group)
		poppedConnections = append(poppedConnections, timedoutConnections...)
		cs.metrics.endConnections.WithLabelValues(group.labelValue, "timeout").Add(float64(len(timedoutConnections)))
	}
	return poppedConnections
}

func (cs *connectionStore) prepareHeartbeats() []connection {
	var connections []connection
	// Iterate over the connections by scheduling groups.
	// In each scheduling group iterate over them by their next heartbeat time from old to new.
	for _, group := range cs.groups {
		group.activeMom.IterateFrontToBack(nextHeartbeatTimeOrder, func(r utils.Record) (shouldDelete, shouldStop bool) {
			conn := r.(connection)
			nextHeartbeat := conn.getNextHeartbeatTime()
			needToReport := cs.now().After(nextHeartbeat)
			if needToReport {
				connections = append(connections, conn)
				cs.updateNextHeartbeatTime(conn.getHash().hashTotal)
				shouldDelete, shouldStop = false, false
			} else {
				shouldDelete, shouldStop = false, true
			}
			return
		})
	}
	return connections
}

func (cs *connectionStore) len() int {
	return len(cs.hashID2groupIdx)
}

// schedulingGroupToLabelValue returns a string representation of a scheduling group to be used as a Prometheus label
// value.
func schedulingGroupToLabelValue(groupIdx int, group api.ConnTrackSchedulingGroup) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%v: ", groupIdx))
	var keys []string
	for k := range group.Selector {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%v, ", k, group.Selector[k]))
	}
	if len(group.Selector) == 0 {
		sb.WriteString("DEFAULT")
	}
	return sb.String()
}

func newConnectionStore(scheduling []api.ConnTrackSchedulingGroup, metrics *metricsType, nowFunc func() time.Time) *connectionStore {
	groups := make([]*groupType, len(scheduling))
	for groupIdx, sg := range scheduling {
		groups[groupIdx] = &groupType{
			scheduling:     sg,
			activeMom:      utils.NewMultiOrderedMap(expiryOrder, nextHeartbeatTimeOrder),
			terminatingMom: utils.NewMultiOrderedMap(expiryOrder, nextHeartbeatTimeOrder),
			labelValue:     schedulingGroupToLabelValue(groupIdx, sg),
		}
	}

	cs := &connectionStore{
		groups:          groups,
		hashID2groupIdx: map[uint64]int{},
		metrics:         metrics,
		now:             nowFunc,
	}
	return cs
}
