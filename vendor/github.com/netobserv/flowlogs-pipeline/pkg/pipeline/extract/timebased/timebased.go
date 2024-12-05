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
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	log "github.com/sirupsen/logrus"
)

type FilterStruct struct {
	Rule              api.TimebasedFilterRule
	IndexKeyDataTable *IndexKeyTable
	Results           filterOperationResults
	Output            []filterOperationResult
}

type filterOperationResults map[string]*filterOperationResult

type filterOperationResult struct {
	values          string
	operationResult float64
}

type DataTableMap map[string]*list.List

type IndexKeyTable struct {
	maxTimeInterval time.Duration
	dataTableMap    DataTableMap
}

type TableEntry struct {
	timeStamp time.Time
	entry     config.GenericMap
}

// CreateIndexKeysAndFilters creates structures for each IndexKey that appears in the rules.
// Note that the same IndexKey might appear in more than one Rule.
// Connect IndexKey structure to its filters.
// For each IndexKey, we need a table of history to handle the largest TimeInterval.
func CreateIndexKeysAndFilters(rules []api.TimebasedFilterRule) (map[string]*IndexKeyTable, []FilterStruct) {
	tmpIndexKeyStructs := make(map[string]*IndexKeyTable)
	tmpFilters := make([]FilterStruct, 0)
	for _, filterRule := range rules {
		log.Debugf("CreateIndexKeysAndFilters: filterRule = %v", filterRule)
		if len(filterRule.IndexKeys) > 0 {
			// reuse indexKey as table index
			filterRule.IndexKey = strings.Join(filterRule.IndexKeys, ",")
		} else if len(filterRule.IndexKey) > 0 {
			// append indexKey to indexKeys
			filterRule.IndexKeys = append(filterRule.IndexKeys, filterRule.IndexKey)
		} else {
			log.Errorf("missing IndexKey(s) for filter %s", filterRule.Name)
			continue
		}
		rStruct, ok := tmpIndexKeyStructs[filterRule.IndexKey]
		if !ok {
			rStruct = &IndexKeyTable{
				maxTimeInterval: filterRule.TimeInterval.Duration,
				dataTableMap:    make(DataTableMap),
			}
			tmpIndexKeyStructs[filterRule.IndexKey] = rStruct
			log.Debugf("new IndexKeyTable: name = %s = %v", filterRule.IndexKey, *rStruct)
		} else if filterRule.TimeInterval.Duration > rStruct.maxTimeInterval {
			rStruct.maxTimeInterval = filterRule.TimeInterval.Duration
		}
		// verify the validity of the OperationType field in the filterRule
		switch filterRule.OperationType {
		case api.FilterOperationLast,
			api.FilterOperationDiff,
			api.FilterOperationCnt,
			api.FilterOperationAvg,
			api.FilterOperationMax,
			api.FilterOperationMin,
			api.FilterOperationSum:
			// OK; nothing to do
		default:
			log.Errorf("illegal operation type %s", filterRule.OperationType)
			continue
		}
		tmpFilter := FilterStruct{
			Rule:              filterRule,
			IndexKeyDataTable: rStruct,
			Results:           make(filterOperationResults),
		}
		log.Debugf("new Rule = %v", tmpFilter)
		tmpFilters = append(tmpFilters, tmpFilter)
	}
	return tmpIndexKeyStructs, tmpFilters
}
