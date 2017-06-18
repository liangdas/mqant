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

// BeeMap is a map with lock
type BeeMap struct {
	lock *sync.RWMutex
	bm   map[interface{}]interface{}
}

// NewBeeMap return new safemap
func NewBeeMap() *BeeMap {
	return &BeeMap{
		lock: new(sync.RWMutex),
		bm:   make(map[interface{}]interface{}),
	}
}

// Get from maps return the k's value
func (m *BeeMap) Get(k interface{}) interface{} {
	m.lock.RLock()
	if val, ok := m.bm[k]; ok {
		m.lock.RUnlock()
		return val
	}
	m.lock.RUnlock()
	return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *BeeMap) Set(k interface{}, v interface{}) bool {
	m.lock.Lock()
	if val, ok := m.bm[k]; !ok {
		m.bm[k] = v
		m.lock.Unlock()
	} else if val != v {
		m.bm[k] = v
		m.lock.Unlock()
	} else {
		m.lock.Unlock()
		return false
	}
	return true
}

// Check Returns true if k is exist in the map.
func (m *BeeMap) Check(k interface{}) bool {
	m.lock.RLock()
	if _, ok := m.bm[k]; !ok {
		m.lock.RUnlock()
		return false
	}
	m.lock.RUnlock()
	return true
}

// Delete the given key and value.
func (m *BeeMap) Delete(k interface{}) {
	m.lock.Lock()
	delete(m.bm, k)
	m.lock.Unlock()
}

func (m *BeeMap) DeleteAll() {
	m.lock.Lock()
	for k, _ := range m.bm {
		delete(m.bm, k)
	}

	m.lock.Unlock()
}

// Items returns all items in safemap.
func (m *BeeMap) Items() map[interface{}]interface{} {
	m.lock.RLock()
	r := make(map[interface{}]interface{})
	for k, v := range m.bm {
		r[k] = v
	}
	m.lock.RUnlock()
	return r
}
