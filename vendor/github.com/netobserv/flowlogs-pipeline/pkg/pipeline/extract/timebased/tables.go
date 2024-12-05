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
	"bytes"
	"container/list"
	"fmt"
	"strings"
	"time"

	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func AddEntryToTables(indexKeyStructs map[string]*IndexKeyTable, entry config.GenericMap, nowInSecs time.Time) {
	for tableKey, recordTable := range indexKeyStructs {
		keys := strings.Split(tableKey, ",")

		validValuesCount := 0
		var b bytes.Buffer
		for _, key := range keys {
			if b.Len() > 0 {
				b.WriteRune(',')
			}
			if val, ok := entry[key]; ok {
				valStr := utils.ConvertToString(val)
				if len(valStr) > 0 {
					b.WriteString(valStr)
					validValuesCount++
				}
			}
		}

		// add entry to the table only if all values are non empty
		if len(keys) == validValuesCount {
			val := b.String()
			log.Debugf("ExtractTimebased addEntryToTables: key = %s, recordTable = %v", tableKey, recordTable)
			cEntry := &TableEntry{
				timeStamp: nowInSecs,
				entry:     entry,
			}
			// allocate list if it does not yet exist
			if recordTable.dataTableMap[val] == nil {
				recordTable.dataTableMap[val] = list.New()
			}
			log.Debugf("ExtractTimebased addEntryToTables: adding to table %s", val)
			AddEntryToTable(cEntry, recordTable.dataTableMap[val])
		}
	}
}

func AddEntryToTable(cEntry *TableEntry, tableList *list.List) {
	log.Debugf("AddEntryToTable: adding table entry %v", cEntry)
	tableList.PushBack(cEntry)
}

func DeleteOldEntriesFromTables(indexKeyStructs map[string]*IndexKeyTable, nowInSecs time.Time) {
	for _, recordTable := range indexKeyStructs {
		oldestTime := nowInSecs.Add(-recordTable.maxTimeInterval)
		for _, tableMap := range recordTable.dataTableMap {
			for {
				head := tableMap.Front()
				if head == nil {
					break
				}
				tableEntry := head.Value.(*TableEntry)
				if tableEntry.timeStamp.Before(oldestTime) {
					tableMap.Remove(head)
					continue
				}
				break
			}
			// TODO: if tableMap is empty, we should clean it up and remove it from recordTable.dataTableMap
		}
	}
}

func PrintTable(l *list.List) {
	fmt.Printf("start PrintTable: \n")
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Printf("PrintTable: e = %v, Value = %v \n", e, e.Value)
	}
	fmt.Printf("end PrintTable: \n")
}
