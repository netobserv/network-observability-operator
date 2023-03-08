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

// This file defines a multi-ordered map data structure. It supports insertion, deletion and retrieval in O(1) on
// average like a regular map. In addition, it allows iterating over the records by multiple orders.
// New records are pushed back to each of the defined orders. Existing records can be moved to the back of a specific
// order by MoveToBack()
// Note: MultiOrderedMap isn't responsible for keeping the records sorted. The user should take care of that.

package utils

import (
	"container/list"
	"fmt"
)

type Key uint64
type Record interface{}
type OrderID string
type processRecordFunc func(Record) (delete, stop bool)

type recordWrapper struct {
	record          Record
	key             Key
	orderID2element map[OrderID]*list.Element
}

type MultiOrderedMap struct {
	m      map[Key]*recordWrapper
	orders map[OrderID]*list.List
}

// NewMultiOrderedMap returns an initialized MultiOrderedMap.
func NewMultiOrderedMap(orderIDs ...OrderID) *MultiOrderedMap {
	mom := &MultiOrderedMap{
		m:      map[Key]*recordWrapper{},
		orders: map[OrderID]*list.List{},
	}
	for _, id := range orderIDs {
		mom.orders[id] = list.New()
	}
	return mom
}

// Len returns the number of records of the multi-ordered map mom.
func (mom MultiOrderedMap) Len() int {
	return len(mom.m)
}

// AddRecord adds a record to the multi-ordered map.
func (mom MultiOrderedMap) AddRecord(key Key, record Record) error {
	if _, found := mom.GetRecord(key); found {
		return fmt.Errorf("record with key %x already exists", key)
	}
	rw := &recordWrapper{key: key, record: record, orderID2element: map[OrderID]*list.Element{}}
	mom.m[key] = rw
	for orderID, orderList := range mom.orders {
		elem := orderList.PushBack(rw)
		rw.orderID2element[orderID] = elem
	}
	return nil
}

// GetRecord returns the record of key `key` and true if the key exists. Otherwise, nil and false is returned.
func (mom MultiOrderedMap) GetRecord(key Key) (Record, bool) {
	rw, found := mom.m[key]
	if !found {
		return nil, false
	}
	return rw.record, true
}

// RemoveRecord removes the record of key `key`. If the key doesn't exist, RemoveRecord is a no-op.
func (mom MultiOrderedMap) RemoveRecord(key Key) {
	rw, found := mom.m[key]
	if !found {
		return
	}
	for orderID, elem := range rw.orderID2element {
		mom.orders[orderID].Remove(elem)
	}
	delete(mom.m, key)
}

// MoveToBack moves the record of key `key` to the back of orderID. If the key or the orderID doesn't exist, an error
// is returned.
func (mom MultiOrderedMap) MoveToBack(key Key, orderID OrderID) error {
	rw, found := mom.m[key]
	if !found {
		return fmt.Errorf("can't MoveToBack non-existing key %x (order id %q)", key, orderID)
	}
	elem, found := rw.orderID2element[orderID]
	if !found {
		return fmt.Errorf("can't MoveToBack non-existing order id %q (key %x)", orderID, key)
	}
	mom.orders[orderID].MoveToBack(elem)
	return nil
}

// MoveToFront moves the record of key `key` to the front of orderID. If the key or the orderID doesn't exist, an error
// is returned.
func (mom MultiOrderedMap) MoveToFront(key Key, orderID OrderID) error {
	rw, found := mom.m[key]
	if !found {
		return fmt.Errorf("can't MoveToFront non-existing key %x (order id %q)", key, orderID)
	}
	elem, found := rw.orderID2element[orderID]
	if !found {
		return fmt.Errorf("can't MoveToFront non-existing order id %q (key %x)", orderID, key)
	}
	mom.orders[orderID].MoveToFront(elem)
	return nil
}

// IterateFrontToBack iterates over the records by orderID. It applies function f() on each record.
// f() returns two booleans `delete` and `stop` that control whether to remove the record from the multi-ordered map
// and whether to stop the iteration respectively.
func (mom MultiOrderedMap) IterateFrontToBack(orderID OrderID, f processRecordFunc) {
	if _, found := mom.orders[orderID]; !found {
		panic(fmt.Sprintf("Unknown order id %q", orderID))
	}
	// How to remove element from list while iterating the same list in golang
	// https://stackoverflow.com/a/27662823/2749989
	var next *list.Element
	for e := mom.orders[orderID].Front(); e != nil; e = next {
		rw := e.Value.(*recordWrapper)
		next = e.Next()
		shouldDelete, shouldStop := f(rw.record)
		if shouldDelete {
			mom.RemoveRecord(rw.key)
		}
		if shouldStop {
			break
		}
	}
}
