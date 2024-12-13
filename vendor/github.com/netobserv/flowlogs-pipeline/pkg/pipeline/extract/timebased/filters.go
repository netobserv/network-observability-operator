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

package timebased

import (
	"container/list"
	"math"
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func (fs *FilterStruct) CalculateResults(nowInSecs time.Time) {
	log.Debugf("CalculateResults nowInSecs = %v", nowInSecs)
	oldestValidTime := nowInSecs.Add(-fs.Rule.TimeInterval.Duration)
	for tableKey, l := range fs.IndexKeyDataTable.dataTableMap {
		var valueFloat64 = float64(0)
		var err error
		//nolint:exhaustive
		switch fs.Rule.OperationType {
		case api.FilterOperationLast:
			// handle empty list
			if l.Len() == 0 {
				continue
			}
			valueFloat64, err = utils.ConvertToFloat64(l.Back().Value.(*TableEntry).entry[fs.Rule.OperationKey])
			if err != nil {
				continue
			}
		case api.FilterOperationDiff:
			for e := l.Front(); e != nil; e = e.Next() {
				cEntry := e.Value.(*TableEntry)
				if cEntry.timeStamp.Before(oldestValidTime) {
					// entry is out of time range; ignore it
					continue
				}
				first, err := utils.ConvertToFloat64(e.Value.(*TableEntry).entry[fs.Rule.OperationKey])
				if err != nil {
					continue
				}
				last, err := utils.ConvertToFloat64(l.Back().Value.(*TableEntry).entry[fs.Rule.OperationKey])
				if err != nil {
					continue
				}
				valueFloat64 = last - first
				break
			}
		default:
			valueFloat64 = fs.CalculateValue(l, oldestValidTime)
		}
		fs.Results[tableKey] = &filterOperationResult{
			values:          tableKey,
			operationResult: valueFloat64,
		}
	}
	log.Debugf("CalculateResults Results = %v", fs.Results)
}

func (fs *FilterStruct) CalculateValue(l *list.List, oldestValidTime time.Time) float64 {
	log.Debugf("CalculateValue nowInSecs = %v", oldestValidTime)
	currentValue := getInitValue(fs.Rule.OperationType)
	nItems := 0
	for e := l.Front(); e != nil; e = e.Next() {
		cEntry := e.Value.(*TableEntry)
		if cEntry.timeStamp.Before(oldestValidTime) {
			// entry is out of time range; ignore it
			continue
		}
		if valueFloat64, err := utils.ConvertToFloat64(cEntry.entry[fs.Rule.OperationKey]); err != nil {
			// Log as debug to avoid performance impact
			log.Debugf("CalculateValue error with OperationKey %s: %v", fs.Rule.OperationKey, err)
		} else {
			nItems++
			switch fs.Rule.OperationType {
			case api.FilterOperationSum, api.FilterOperationAvg:
				currentValue += valueFloat64
			case api.FilterOperationMax:
				currentValue = math.Max(currentValue, valueFloat64)
			case api.FilterOperationMin:
				currentValue = math.Min(currentValue, valueFloat64)
			case api.FilterOperationCnt, api.FilterOperationLast, api.FilterOperationDiff:
			}
		}
	}
	if fs.Rule.OperationType == api.FilterOperationAvg && nItems > 0 {
		currentValue /= float64(nItems)
	}
	if fs.Rule.OperationType == api.FilterOperationCnt {
		currentValue = float64(nItems)
	}
	return currentValue
}

func getInitValue(operation api.FilterOperationEnum) float64 {
	switch operation {
	case api.FilterOperationSum,
		api.FilterOperationAvg,
		api.FilterOperationCnt,
		api.FilterOperationLast,
		api.FilterOperationDiff:
		return 0
	case api.FilterOperationMax:
		return (-math.MaxFloat64)
	case api.FilterOperationMin:
		return math.MaxFloat64
	default:
		log.Panicf("unknown operation %v", operation)
		return 0
	}
}

func (fs *FilterStruct) ComputeTopkBotk() {
	var output []filterOperationResult
	if fs.Rule.TopK > 0 {
		if fs.Rule.Reversed {
			output = fs.computeBotK(fs.Results)
		} else {
			output = fs.computeTopK(fs.Results)
		}
	} else {
		// return all Results; convert map to array
		output = make([]filterOperationResult, len(fs.Results))
		i := 0
		for _, item := range fs.Results {
			output[i] = *item
			i++
		}
	}
	fs.Output = output
}

func (fs *FilterStruct) CreateGenericMap() []config.GenericMap {
	output := make([]config.GenericMap, 0)
	for _, result := range fs.Output {
		t := config.GenericMap{
			"name":      fs.Rule.Name,
			"index_key": fs.Rule.IndexKey,
			"operation": fs.Rule.OperationType,
		}

		// append operation key and result as key / value
		t[fs.Rule.OperationKey] = result.operationResult

		// append index key / value pairs
		values := strings.Split(result.values, ",")
		if len(fs.Rule.IndexKeys) == len(values) {
			for i, k := range fs.Rule.IndexKeys {
				t[k] = values[i]
			}
		}

		log.Debugf("FilterStruct CreateGenericMap: %v", t)
		output = append(output, t)
	}
	log.Debugf("FilterStruct CreateGenericMap: output = %v \n", output)
	return output
}
