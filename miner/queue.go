// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package miner

import (
	"container/heap"
	"sync"
	"xfsgo"
)

type txPrioItem struct {
	tx       *xfsgo.Transaction
	priority float32
}

type qs []*txPrioItem

func (q qs) Len() int {
	return len(q)
}

func (q qs) Less(i, j int) bool {
	return q[i].priority < q[j].priority
}

func (q qs) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *qs) Pop() interface{} {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[0 : n-1]
	return x
}
func (q *qs) Push(x interface{}) {
	*q = append(*q, x.(*txPrioItem))
}

type txPriorityQueue struct {
	mu    sync.RWMutex
	items *qs
}

func newTxPriorityQueue() *txPriorityQueue {
	queue := &txPriorityQueue{
		items: &qs{},
	}
	heap.Init(queue.items)
	return queue
}

func (queue *txPriorityQueue) PushTxs(item *txPrioItem) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	heap.Push(queue.items, item)
}

func (queue *txPriorityQueue) Push(item *txPrioItem) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	heap.Push(queue.items, item)
}

func (queue *txPriorityQueue) Pop() *txPrioItem {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return heap.Pop(queue.items).(*txPrioItem)
}
func (queue *txPriorityQueue) Len() int {
	queue.mu.RLock()
	defer queue.mu.RUnlock()
	return len(*queue.items)
}
