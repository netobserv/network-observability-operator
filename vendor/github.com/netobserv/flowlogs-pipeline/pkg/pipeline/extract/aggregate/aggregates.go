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

package aggregate

import (
	"sync"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	log "github.com/sirupsen/logrus"
)

var defaultExpiryTime = 2 * time.Minute
var cleanupLoopTime = 2 * time.Minute

type Aggregates struct {
	Aggregates        []Aggregate
	cleanupLoopTime   time.Duration
	defaultExpiryTime time.Duration
}

func (aggregates *Aggregates) Evaluate(entries []config.GenericMap) error {
	for _, aggregate := range aggregates.Aggregates {
		err := aggregate.Evaluate(entries)
		if err != nil {
			log.Debugf("Evaluate error %v", err)
			continue
		}
	}

	return nil
}

func (aggregates *Aggregates) GetMetrics() []config.GenericMap {
	var metrics []config.GenericMap
	for _, aggregate := range aggregates.Aggregates {
		aggregateMetrics := aggregate.GetMetrics()
		metrics = append(metrics, aggregateMetrics...)
	}

	return metrics
}

func (aggregates *Aggregates) addAggregate(aggregateDefinition *api.AggregateDefinition) []Aggregate {
	expiryTime := aggregateDefinition.ExpiryTime
	if expiryTime.Duration == 0 {
		expiryTime.Duration = defaultExpiryTime
	}
	aggregate := Aggregate{
		definition: aggregateDefinition,
		cache:      utils.NewTimedCache(0, nil),
		mutex:      &sync.Mutex{},
		expiryTime: expiryTime.Duration,
	}

	return append(aggregates.Aggregates, aggregate)
}

func (aggregates *Aggregates) cleanupExpiredEntriesLoop() {

	ticker := time.NewTicker(aggregates.cleanupLoopTime)
	go func() {
		for {
			select {
			case <-utils.ExitChannel():
				return
			case <-ticker.C:
				aggregates.cleanupExpiredEntries()
			}
		}
	}()
}

func (aggregates *Aggregates) cleanupExpiredEntries() {
	for _, aggregate := range aggregates.Aggregates {
		aggregate.mutex.Lock()
		aggregate.cache.CleanupExpiredEntries(aggregate.expiryTime, func(_ interface{}) {})
		aggregate.mutex.Unlock()
	}
}

func NewAggregatesFromConfig(aggConfig *api.Aggregates) (Aggregates, error) {
	aggregates := Aggregates{
		cleanupLoopTime:   cleanupLoopTime,
		defaultExpiryTime: aggConfig.DefaultExpiryTime.Duration,
	}
	if aggregates.defaultExpiryTime == 0 {
		aggregates.defaultExpiryTime = defaultExpiryTime
	}

	for i := range aggConfig.Rules {
		aggregates.Aggregates = aggregates.addAggregate(&aggConfig.Rules[i])
	}

	aggregates.cleanupExpiredEntriesLoop()

	return aggregates, nil
}
