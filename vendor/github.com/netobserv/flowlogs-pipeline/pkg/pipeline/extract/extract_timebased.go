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

package extract

import (
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	tb "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/extract/timebased"
	log "github.com/sirupsen/logrus"
)

type timebased struct {
	Filters         []tb.FilterStruct
	IndexKeyStructs map[string]*tb.IndexKeyTable
}

// Extract extracts a flow before being stored
func (et *timebased) Extract(entries []config.GenericMap) []config.GenericMap {
	log.Debugf("entering timebased Extract")
	nowInSecs := time.Now()
	// Populate the Table with the current entries
	for _, entry := range entries {
		log.Debugf("timebased Extract, entry = %v", entry)
		tb.AddEntryToTables(et.IndexKeyStructs, entry, nowInSecs)
	}

	output := make([]config.GenericMap, 0)
	// Calculate Filters based on time windows
	for i := range et.Filters {
		filter := &et.Filters[i]
		filter.CalculateResults(nowInSecs)
		filter.ComputeTopkBotk()
		genMap := filter.CreateGenericMap()
		output = append(output, genMap...)
	}
	log.Debugf("output of extract timebased: %v", output)

	// delete entries from tables that are outside time windows
	tb.DeleteOldEntriesFromTables(et.IndexKeyStructs, nowInSecs)

	return output
}

// NewExtractTimebased creates a new extractor
func NewExtractTimebased(params config.StageParam) (Extractor, error) {
	var rules []api.TimebasedFilterRule
	if params.Extract != nil && params.Extract.Timebased.Rules != nil {
		rules = params.Extract.Timebased.Rules
	}
	log.Debugf("NewExtractTimebased; rules = %v", rules)

	tmpIndexKeyStructs, tmpFilters := tb.CreateIndexKeysAndFilters(rules)

	return &timebased{
		Filters:         tmpFilters,
		IndexKeyStructs: tmpIndexKeyStructs,
	}, nil
}
