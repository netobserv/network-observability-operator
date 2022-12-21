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

package utils

import (
	"container/list"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "utils.TimedCache")

// Functions to manage an LRU cache with an expiry
// When an item expires, allow a callback to allow the specific implementation to perform its particular cleanup

type CacheCallback interface {
	Cleanup(entry interface{})
}

type cacheEntry struct {
	key             string
	lastUpdatedTime int64
	e               *list.Element
	SourceEntry     interface{}
}

type TimedCacheMap map[string]*cacheEntry

type TimedCache struct {
	mu        sync.RWMutex
	cacheList *list.List
	cacheMap  TimedCacheMap
}

func (tc *TimedCache) GetCacheEntry(key string) (interface{}, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	cEntry, ok := tc.cacheMap[key]
	if ok {
		return cEntry.SourceEntry, ok
	} else {
		return nil, ok
	}
}

var uclog = log.WithField("method", "UpdateCacheEntry")

func (tc *TimedCache) UpdateCacheEntry(key string, entry interface{}) *cacheEntry {
	nowInSecs := time.Now().Unix()
	tc.mu.Lock()
	defer tc.mu.Unlock()
	cEntry, ok := tc.cacheMap[key]
	if ok {
		// item already exists in cache; update the element and move to end of list
		cEntry.lastUpdatedTime = nowInSecs
		// move to end of list
		tc.cacheList.MoveToBack(cEntry.e)
	} else {
		// create new entry for cache
		cEntry = &cacheEntry{
			lastUpdatedTime: nowInSecs,
			key:             key,
			SourceEntry:     entry,
		}
		uclog.Debugf("adding entry: %#v", cEntry)
		// place at end of list
		cEntry.e = tc.cacheList.PushBack(cEntry)
		tc.cacheMap[key] = cEntry
	}
	return cEntry
}

func (tc *TimedCache) GetCacheLen() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.cacheList.Len()
}

// We expect that the function calling Iterate might make updates to the entries by calling UpdateCacheEntry()
// We therefore cannot take the lock at this point since it will conflict with the call in UpdateCacheEntry()
// TODO: If the callback needs to update the cache, then we need a method to perform it without taking the lock again.
func (tc *TimedCache) Iterate(f func(key string, value interface{})) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	for k, v := range tc.cacheMap {
		f(k, v.SourceEntry)
	}
}

// CleanupExpiredEntries removes items from cache that were last touched more than expiryTime seconds ago
func (tc *TimedCache) CleanupExpiredEntries(expiryTime int64, callback CacheCallback) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	clog := log.WithFields(logrus.Fields{
		"mapLen":  len(tc.cacheMap),
		"listLen": tc.cacheList.Len(),
	})
	clog.Debugf("cleaning up expried entries")

	nowInSecs := time.Now().Unix()
	expireTime := nowInSecs - expiryTime
	deleted := 0
	// go through the list until we reach recently used entries
	for {
		listEntry := tc.cacheList.Front()
		if listEntry == nil {
			return
		}
		pCacheInfo := listEntry.Value.(*cacheEntry)
		if pCacheInfo.lastUpdatedTime > expireTime {
			// no more expired items
			clog.Debugf("deleted %d expired entries", deleted)
			return
		}
		deleted++
		callback.Cleanup(pCacheInfo.SourceEntry)
		delete(tc.cacheMap, pCacheInfo.key)
		tc.cacheList.Remove(listEntry)
	}
}

func NewTimedCache() *TimedCache {
	l := &TimedCache{
		cacheList: list.New(),
		cacheMap:  make(TimedCacheMap),
	}
	return l
}
