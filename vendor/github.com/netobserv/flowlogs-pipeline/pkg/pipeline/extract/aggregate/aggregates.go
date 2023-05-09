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

var defaultExpiryTime = 10 * time.Minute

type Aggregates struct {
	Aggregates []Aggregate
	expiryTime time.Duration
}

type Definitions []api.AggregateDefinition

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

func (aggregates *Aggregates) AddAggregate(aggregateDefinition api.AggregateDefinition) []Aggregate {
	aggregate := Aggregate{
		Definition: aggregateDefinition,
		cache:      utils.NewTimedCache(0, nil),
		mutex:      &sync.Mutex{},
		expiryTime: aggregates.expiryTime,
	}

	appendedAggregates := append(aggregates.Aggregates, aggregate)
	return appendedAggregates
}

func (aggregates *Aggregates) cleanupExpiredEntriesLoop() {

	ticker := time.NewTicker(aggregates.expiryTime)
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
		aggregate.cache.CleanupExpiredEntries(aggregate.expiryTime, aggregate.Cleanup)
		aggregate.mutex.Unlock()
	}
}

func NewAggregatesFromConfig(definitions []api.AggregateDefinition) (Aggregates, error) {
	aggregates := Aggregates{
		expiryTime: defaultExpiryTime,
	}

	for _, aggregateDefinition := range definitions {
		aggregates.Aggregates = aggregates.AddAggregate(aggregateDefinition)
	}

	aggregates.cleanupExpiredEntriesLoop()

	return aggregates, nil
}
