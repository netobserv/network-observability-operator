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
	"container/heap"
	"math"

	log "github.com/sirupsen/logrus"
)

// functions to manipulate a heap to generate TopK/BotK entries
// We need to implement the heap interface: Len(), Less(), Swap(), Push(), Pop()

type heapItem struct {
	value  float64
	result *filterOperationResult
}

type topkHeap []heapItem
type botkHeap []heapItem

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

func (fs *FilterStruct) computeTopK(inputs filterOperationResults) []filterOperationResult {
	// maintain a heap with k items, always dropping the lowest
	// we will be left with the TopK items
	var prevMin float64
	prevMin = -math.MaxFloat64
	topk := fs.Rule.TopK
	h := &topkHeap{}
	for key, metricMap := range inputs {
		val := metricMap.operationResult
		if val < prevMin {
			continue
		}
		item := heapItem{
			result: inputs[key],
			value:  val,
		}
		heap.Push(h, item)
		if h.Len() > topk {
			x := heap.Pop(h)
			prevMin = x.(heapItem).value
		}
	}
	log.Debugf("heap: %v", h)

	// convert the remaining heap to a sorted array
	result := make([]filterOperationResult, h.Len())
	heapLen := h.Len()
	for i := heapLen; i > 0; i-- {
		poppedItem := heap.Pop(h).(heapItem)
		log.Debugf("poppedItem: %v", poppedItem)
		result[i-1] = *poppedItem.result
	}
	log.Debugf("topk items: %v", result)
	return result
}

func (h botkHeap) Len() int {
	return len(h)
}

// For a botk heap, we reverse the order of the Less() operation
func (h botkHeap) Less(i, j int) bool {
	return h[i].value > h[j].value
}

func (h botkHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *botkHeap) Push(x interface{}) {
	*h = append(*h, x.(heapItem))
}

func (h *botkHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (fs *FilterStruct) computeBotK(inputs filterOperationResults) []filterOperationResult {
	// maintain a heap with k items, always dropping the highest
	// we will be left with the BotK items
	var prevMax float64
	prevMax = math.MaxFloat64
	botk := fs.Rule.TopK
	h := &botkHeap{}
	for key, metricMap := range inputs {
		val := metricMap.operationResult
		if val > prevMax {
			continue
		}
		item := heapItem{
			result: inputs[key],
			value:  val,
		}
		heap.Push(h, item)
		if h.Len() > botk {
			x := heap.Pop(h)
			prevMax = x.(heapItem).value
		}
	}
	log.Debugf("heap: %v", h)

	// convert the remaining heap to a sorted array
	result := make([]filterOperationResult, h.Len())
	heapLen := h.Len()
	for i := heapLen; i > 0; i-- {
		poppedItem := heap.Pop(h).(heapItem)
		log.Debugf("poppedItem: %v", poppedItem)
		result[i-1] = *poppedItem.result
	}
	log.Debugf("botk items: %v", result)
	return result
}
