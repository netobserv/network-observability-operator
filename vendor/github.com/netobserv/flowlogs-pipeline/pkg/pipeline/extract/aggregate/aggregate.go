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
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
	util "github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	OperationSum       = "sum"
	OperationAvg       = "avg"
	OperationMax       = "max"
	OperationMin       = "min"
	OperationCount     = "count"
	OperationRawValues = "raw_values"
)

type Labels map[string]string
type NormalizedValues string

type Aggregate struct {
	definition *api.AggregateDefinition
	cache      *utils.TimedCache
	mutex      *sync.Mutex
	expiryTime time.Duration
}

type GroupState struct {
	normalizedValues NormalizedValues
	labels           Labels
	recentRawValues  []float64
	recentOpValue    float64
	recentCount      int
	totalValue       float64
	totalCount       int
}

func (aggregate *Aggregate) LabelsFromEntry(entry config.GenericMap) (Labels, bool) {
	allLabelsFound := true
	labels := Labels{}

	for _, key := range aggregate.definition.GroupByKeys {
		value, ok := entry[key]
		if !ok {
			allLabelsFound = false
		}
		labels[key] = util.ConvertToString(value)
	}

	return labels, allLabelsFound
}

func (labels Labels) getNormalizedValues() NormalizedValues {
	var normalizedAsString string

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		normalizedAsString += labels[k] + ","
	}

	if len(normalizedAsString) > 0 {
		normalizedAsString = normalizedAsString[:len(normalizedAsString)-1]
	}

	return NormalizedValues(normalizedAsString)
}

func (aggregate *Aggregate) filterEntry(entry config.GenericMap) (NormalizedValues, Labels, error) {
	labels, allLabelsFound := aggregate.LabelsFromEntry(entry)
	if !allLabelsFound {
		return "", nil, fmt.Errorf("missing keys in entry")
	}

	normalizedValues := labels.getNormalizedValues()
	return normalizedValues, labels, nil
}

func getInitValue(operation string) float64 {
	switch operation {
	case OperationSum, OperationAvg, OperationMax, OperationCount:
		return 0
	case OperationMin:
		return math.MaxFloat64
	case OperationRawValues:
		// Actually, in OperationRawValues the value is ignored.
		return 0
	default:
		log.Panicf("unknown operation %v", operation)
		return 0
	}
}

func (aggregate *Aggregate) UpdateByEntry(entry config.GenericMap, normalizedValues NormalizedValues, labels Labels) error {

	aggregate.mutex.Lock()
	defer aggregate.mutex.Unlock()

	var groupState *GroupState
	oldEntry, ok := aggregate.cache.GetCacheEntry(string(normalizedValues))
	if !ok {
		groupState = &GroupState{normalizedValues: normalizedValues, labels: labels}
		initVal := getInitValue(string(aggregate.definition.OperationType))
		groupState.totalValue = initVal
		groupState.recentOpValue = initVal
		if aggregate.definition.OperationType == OperationRawValues {
			groupState.recentRawValues = make([]float64, 0)
		}
	} else {
		groupState = oldEntry.(*GroupState)
	}
	aggregate.cache.UpdateCacheEntry(string(normalizedValues), groupState)

	// update value
	operationKey := aggregate.definition.OperationKey
	operation := aggregate.definition.OperationType

	if operation == OperationCount {
		groupState.totalValue = float64(groupState.totalCount + 1)
		groupState.recentOpValue = float64(groupState.recentCount + 1)
	} else if operationKey != "" {
		value, ok := entry[operationKey]
		if ok {
			valueString := util.ConvertToString(value)
			if valueFloat64, err := strconv.ParseFloat(valueString, 64); err != nil {
				// Log as debug to avoid performance impact
				log.Debugf("UpdateByEntry error when parsing float '%s': %v", valueString, err)
			} else {
				switch operation {
				case OperationSum:
					groupState.totalValue += valueFloat64
					groupState.recentOpValue += valueFloat64
				case OperationMax:
					groupState.totalValue = math.Max(groupState.totalValue, valueFloat64)
					groupState.recentOpValue = math.Max(groupState.recentOpValue, valueFloat64)
				case OperationMin:
					groupState.totalValue = math.Min(groupState.totalValue, valueFloat64)
					groupState.recentOpValue = math.Min(groupState.recentOpValue, valueFloat64)
				case OperationAvg:
					groupState.totalValue = (groupState.totalValue*float64(groupState.totalCount) + valueFloat64) / float64(groupState.totalCount+1)
					groupState.recentOpValue = (groupState.recentOpValue*float64(groupState.recentCount) + valueFloat64) / float64(groupState.recentCount+1)
				case OperationRawValues:
					groupState.recentRawValues = append(groupState.recentRawValues, valueFloat64)
				}
			}
		}
	}

	// update count
	groupState.totalCount++
	groupState.recentCount++

	return nil
}

func (aggregate *Aggregate) Evaluate(entries []config.GenericMap) error {
	for _, entry := range entries {
		// filter entries matching labels with aggregates
		normalizedValues, labels, err := aggregate.filterEntry(entry)
		if err != nil {
			continue
		}

		// update aggregate group by entry
		err = aggregate.UpdateByEntry(entry, normalizedValues, labels)
		if err != nil {
			log.Debugf("UpdateByEntry error %v", err)
			continue
		}
	}

	return nil
}

func (aggregate *Aggregate) GetMetrics() []config.GenericMap {
	aggregate.mutex.Lock()
	defer aggregate.mutex.Unlock()

	var metrics []config.GenericMap

	// iterate over the items in the cache
	aggregate.cache.Iterate(func(_ string, value interface{}) {
		group := value.(*GroupState)
		newEntry := config.GenericMap{
			"name":              aggregate.definition.Name,
			"operation_type":    aggregate.definition.OperationType,
			"operation_key":     aggregate.definition.OperationKey,
			"by":                strings.Join(aggregate.definition.GroupByKeys, ","),
			"aggregate":         string(group.normalizedValues),
			"total_value":       group.totalValue,
			"total_count":       group.totalCount,
			"recent_raw_values": group.recentRawValues,
			"recent_op_value":   group.recentOpValue,
			"recent_count":      group.recentCount,
			strings.Join(aggregate.definition.GroupByKeys, "_"): string(group.normalizedValues),
		}
		// add the items in aggregate.definition.GroupByKeys individually to the entry
		for _, key := range aggregate.definition.GroupByKeys {
			newEntry[key] = group.labels[key]
		}
		metrics = append(metrics, newEntry)
		// Once reported, we reset the recentXXX fields
		if aggregate.definition.OperationType == OperationRawValues {
			group.recentRawValues = make([]float64, 0)
		}
		group.recentCount = 0
		group.recentOpValue = getInitValue(string(aggregate.definition.OperationType))
	})

	return metrics
}
