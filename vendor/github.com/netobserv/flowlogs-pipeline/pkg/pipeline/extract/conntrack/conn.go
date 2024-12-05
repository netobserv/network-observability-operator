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
	"reflect"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
)

type connection interface {
	addAgg(fieldName string, initValue interface{})
	updateAggValue(fieldName string, newValue interface{})
	updateAggFnValue(fieldName string, newValueFn func(curr float64) float64)
	setExpiryTime(t time.Time)
	getExpiryTime() time.Time
	setNextHeartbeatTime(t time.Time)
	getNextHeartbeatTime() time.Time
	toGenericMap() config.GenericMap
	getHash() totalHashType
	// markReported marks the connection as has been reported. That is, at least one connection record has been emitted
	// for this connection (i.e. newConnection, heartbeat, endConnection).
	// It returns true on the first invocation to indicate the first report. Otherwise, it returns false.
	markReported() bool
	isMatchSelector(map[string]interface{}) bool
}

type connType struct {
	hash              totalHashType
	keys              config.GenericMap
	aggFields         map[string]interface{}
	expiryTime        time.Time
	nextHeartbeatTime time.Time
	isReported        bool
}

func (c *connType) addAgg(fieldName string, initValue interface{}) {
	c.aggFields[fieldName] = initValue
}

func (c *connType) updateAggValue(fieldName string, newValue interface{}) {
	_, ok := c.aggFields[fieldName]
	if !ok {
		log.Panicf("tried updating missing field %v", fieldName)
	}
	c.aggFields[fieldName] = newValue
}

func (c *connType) updateAggFnValue(fieldName string, newValueFn func(curr float64) float64) {
	v, ok := c.aggFields[fieldName]
	if !ok {
		log.Panicf("tried updating missing field %v", fieldName)
	}

	// existing value must be float64 for function aggregation
	switch value := v.(type) {
	case float64:
		c.aggFields[fieldName] = newValueFn(value)
	default:
		log.Panicf("tried to aggregate non float64 field %v value %v", fieldName, v)
	}
}

func (c *connType) setExpiryTime(t time.Time) {
	c.expiryTime = t
}

func (c *connType) getExpiryTime() time.Time {
	return c.expiryTime
}

func (c *connType) setNextHeartbeatTime(t time.Time) {
	c.nextHeartbeatTime = t
}

func (c *connType) getNextHeartbeatTime() time.Time {
	return c.nextHeartbeatTime
}

func (c *connType) toGenericMap() config.GenericMap {
	gm := config.GenericMap{}
	for k, v := range c.aggFields {
		if v != nil && (reflect.TypeOf(v).Kind() != reflect.Float64 || v.(float64) != 0) {
			gm[k] = v
		}
	}

	// In case of a conflict between the keys and the aggFields / cpFields, the keys should prevail.
	for k, v := range c.keys {
		gm[k] = v
	}
	return gm
}

func (c *connType) getHash() totalHashType {
	return c.hash
}

func (c *connType) markReported() bool {
	isFirst := !c.isReported
	c.isReported = true
	return isFirst
}

//nolint:cyclop
func (c *connType) isMatchSelector(selector map[string]interface{}) bool {
	for k, v := range selector {
		connValueRaw, found := c.keys[k]
		if !found {
			return false
		}
		switch connValue := connValueRaw.(type) {
		case int:
			selectorValue, err := utils.ConvertToInt(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case uint32:
			selectorValue, err := utils.ConvertToUint32(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case uint64:
			selectorValue, err := utils.ConvertToUint64(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case int64:
			selectorValue, err := utils.ConvertToInt64(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case float64:
			selectorValue, err := utils.ConvertToFloat64(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case bool:
			selectorValue, err := utils.ConvertToBool(v)
			if err != nil || connValue != selectorValue {
				return false
			}
		case string:
			selectorValue := utils.ConvertToString(v)
			if connValue != selectorValue {
				return false
			}
		default:
			connValue = utils.ConvertToString(connValue)
			selectorValue := fmt.Sprintf("%v", v)
			if connValue != selectorValue {
				return false
			}
		}
	}
	return true
}

type connBuilder struct {
	conn         *connType
	shouldSwapAB bool
	metrics      *metricsType
}

func newConnBuilder(metrics *metricsType) *connBuilder {
	return &connBuilder{
		conn: &connType{
			aggFields:  make(map[string]interface{}),
			keys:       config.GenericMap{},
			isReported: false,
		},
		metrics: metrics,
	}
}

func (cb *connBuilder) Hash(h totalHashType) *connBuilder {
	if cb.shouldSwapAB {
		h.hashA, h.hashB = h.hashB, h.hashA
	}
	cb.conn.hash = h
	return cb
}

func (cb *connBuilder) ShouldSwapAB(b bool) *connBuilder {
	cb.shouldSwapAB = b
	return cb
}

func (cb *connBuilder) keysFrom(flowLog config.GenericMap, kd *api.KeyDefinition, endpointAFields, endpointBFields []string) *connBuilder {
	for _, fg := range kd.FieldGroups {
		for _, f := range fg.Fields {
			cb.conn.keys[f] = flowLog[f]
		}
	}
	if cb.shouldSwapAB {
		for i := range endpointAFields {
			fieldA := endpointAFields[i]
			fieldB := endpointBFields[i]
			cb.conn.keys[fieldA] = flowLog[fieldB]
			cb.conn.keys[fieldB] = flowLog[fieldA]
		}
		cb.metrics.tcpFlags.WithLabelValues("swapAB").Inc()
	}
	return cb
}

func (cb *connBuilder) Aggregators(aggs []aggregator) *connBuilder {
	for _, agg := range aggs {
		agg.addField(cb.conn)
	}
	return cb
}

func (cb *connBuilder) Build() connection {
	return cb.conn
}
