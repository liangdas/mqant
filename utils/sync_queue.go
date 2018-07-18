// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package utils

import (
	"sync"
)

// Synchronous FIFO queue
type SyncQueue struct {
	lock    sync.Mutex
	popable *sync.Cond
	buffer  *Queue
	closed  bool
}

// Create a new SyncQueue
func NewSyncQueue() *SyncQueue {
	ch := &SyncQueue{
		buffer: NewQueue(),
	}
	ch.popable = sync.NewCond(&ch.lock)
	return ch
}

// Pop an item from SyncQueue, will block if SyncQueue is empty
func (q *SyncQueue) Pop() (v interface{}) {
	c := q.popable
	buffer := q.buffer

	q.lock.Lock()
	for buffer.Length() == 0 && !q.closed {
		c.Wait()
	}

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
	}

	q.lock.Unlock()
	return
}

// Try to pop an item from SyncQueue, will return immediately with bool=false if SyncQueue is empty
func (q *SyncQueue) TryPop() (v interface{}, ok bool) {
	buffer := q.buffer

	q.lock.Lock()

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
		ok = true
	} else if q.closed {
		ok = true
	}

	q.lock.Unlock()
	return
}

// Push an item to SyncQueue. Always returns immediately without blocking
func (q *SyncQueue) Push(v interface{}) {
	q.lock.Lock()
	if !q.closed {
		q.buffer.Add(v)
		q.popable.Signal()
	}
	q.lock.Unlock()
}

// Get the length of SyncQueue
func (q *SyncQueue) Len() (l int) {
	q.lock.Lock()
	l = q.buffer.Length()
	q.lock.Unlock()
	return
}

// Close SyncQueue
//
// After close, Pop will return nil without block, and TryPop will return v=nil, ok=True
func (q *SyncQueue) Close() {
	q.lock.Lock()
	if !q.closed {
		q.closed = true
		q.popable.Signal()
	}
	q.lock.Unlock()
}

func (q *SyncQueue) IsClose() (v bool) {
	q.lock.Lock()
	v = q.closed
	q.lock.Unlock()
	return
}
