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
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/pipeline/utils"
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
	Definition api.AggregateDefinition
	cache      *utils.TimedCache
	mutex      *sync.Mutex
	expiryTime int64
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

func (aggregate Aggregate) LabelsFromEntry(entry config.GenericMap) (Labels, bool) {
	allLabelsFound := true
	labels := Labels{}

	for _, key := range aggregate.Definition.By {
		value, ok := entry[key]
		if !ok {
			allLabelsFound = false
		}
		labels[key] = fmt.Sprintf("%v", value)
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

func (aggregate Aggregate) FilterEntry(entry config.GenericMap) (error, NormalizedValues, Labels) {
	labels, allLabelsFound := aggregate.LabelsFromEntry(entry)
	if !allLabelsFound {
		return fmt.Errorf("missing keys in entry"), "", nil
	}

	normalizedValues := labels.getNormalizedValues()
	return nil, normalizedValues, labels
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
		log.Panicf("unkown operation %v", operation)
		return 0
	}
}

func (aggregate Aggregate) UpdateByEntry(entry config.GenericMap, normalizedValues NormalizedValues, labels Labels) error {

	aggregate.mutex.Lock()
	defer aggregate.mutex.Unlock()

	var groupState *GroupState
	oldEntry, ok := aggregate.cache.GetCacheEntry(string(normalizedValues))
	if !ok {
		groupState = &GroupState{normalizedValues: normalizedValues, labels: labels}
		initVal := getInitValue(string(aggregate.Definition.Operation))
		groupState.totalValue = initVal
		groupState.recentOpValue = initVal
		if aggregate.Definition.Operation == OperationRawValues {
			groupState.recentRawValues = make([]float64, 0)
		}
	} else {
		groupState = oldEntry.(*GroupState)
	}
	aggregate.cache.UpdateCacheEntry(string(normalizedValues), groupState)

	// update value
	recordKey := aggregate.Definition.RecordKey
	operation := aggregate.Definition.Operation

	if operation == OperationCount {
		groupState.totalValue = float64(groupState.totalCount + 1)
		groupState.recentOpValue = float64(groupState.recentCount + 1)
	} else {
		if recordKey != "" {
			value, ok := entry[recordKey]
			if ok {
				valueString := fmt.Sprintf("%v", value)
				valueFloat64, _ := strconv.ParseFloat(valueString, 64)
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
	groupState.totalCount += 1
	groupState.recentCount += 1

	return nil
}

func (aggregate Aggregate) Evaluate(entries []config.GenericMap) error {
	for _, entry := range entries {
		// filter entries matching labels with aggregates
		err, normalizedValues, labels := aggregate.FilterEntry(entry)
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

func (aggregate Aggregate) GetMetrics() []config.GenericMap {
	aggregate.mutex.Lock()
	defer aggregate.mutex.Unlock()

	var metrics []config.GenericMap

	// iterate over the items in the cache
	aggregate.cache.Iterate(func(key string, value interface{}) {
		group := value.(*GroupState)
		newEntry := config.GenericMap{
			"name":              aggregate.Definition.Name,
			"operation":         aggregate.Definition.Operation,
			"record_key":        aggregate.Definition.RecordKey,
			"by":                strings.Join(aggregate.Definition.By, ","),
			"aggregate":         string(group.normalizedValues),
			"total_value":       group.totalValue,
			"total_count":       group.totalCount,
			"recent_raw_values": group.recentRawValues,
			"recent_op_value":   group.recentOpValue,
			"recent_count":      group.recentCount,
			strings.Join(aggregate.Definition.By, "_"): string(group.normalizedValues),
		}
		// add the items in aggregate.Definition.By individually to the entry
		for _, key := range aggregate.Definition.By {
			newEntry[key] = group.labels[key]
		}
		metrics = append(metrics, newEntry)
		// Once reported, we reset the recentXXX fields
		if aggregate.Definition.Operation == OperationRawValues {
			group.recentRawValues = make([]float64, 0)
		}
		group.recentCount = 0
		group.recentOpValue = getInitValue(string(aggregate.Definition.Operation))
	})

	if aggregate.Definition.TopK > 0 {
		metrics = aggregate.computeTopK(metrics)
	}

	return metrics
}

func (aggregate Aggregate) Cleanup(entry interface{}) {
	// nothing special to do in this callback function
}

// functions to manipulate a heap to generate TopK entries
// We need to implement the heap interface: Len(), Less(), Swap(), Push(), Pop()

type heapItem struct {
	value   float64
	metrics *config.GenericMap
}

type topkHeap []heapItem

func (h topkHeap) Len() int {
	return len(h)
}

func (h topkHeap) Less(i, j int) bool {
	return h[i].value < h[j].value
}

func (h topkHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *topkHeap) Push(x interface{}) {
	*h = append(*h, x.(heapItem))
}

func (h *topkHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (aggregate Aggregate) computeTopK(inputMetrics []config.GenericMap) []config.GenericMap {
	// maintain a heap with k items, always dropping the lowest
	// we will be left with the TopK items
	var prevMin float64
	prevMin = -math.MaxFloat64
	topk := aggregate.Definition.TopK
	h := &topkHeap{}
	for index, metricMap := range inputMetrics {
		val := metricMap["total_value"].(float64)
		if val < prevMin {
			continue
		}
		item := heapItem{
			metrics: &inputMetrics[index],
			value:   val,
		}
		heap.Push(h, item)
		if h.Len() > topk {
			x := heap.Pop(h)
			prevMin = x.(heapItem).value
		}
	}
	log.Debugf("heap: %v", h)

	// convert the remaining heap to a sorted array
	result := make([]config.GenericMap, h.Len())
	heapLen := h.Len()
	for i := heapLen; i > 0; i-- {
		poppedItem := heap.Pop(h).(heapItem)
		log.Debugf("poppedItem: %v", poppedItem)
		result[i-1] = *poppedItem.metrics
	}
	log.Debugf("topk items: %v", result)
	return result
}
