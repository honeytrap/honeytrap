// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package web

import (
	"encoding/json"
	"sync"
)

// NewLimitedSafeArray returns a SafeArray with a max limit items.
func NewLimitedSafeArray(limit int) *SafeArray {
	return &SafeArray{
		array: []interface{}{},
		limit: limit,
	}
}

// NewSafeArray returns a unlimited SafeArray.
func NewSafeArray() *SafeArray {
	return &SafeArray{
		array: []interface{}{},
		limit: 0,
	}
}

// SafeArray is a thread safe array implementation.
type SafeArray struct {
	array []interface{}
	limit int
	m     sync.Mutex
}

// Append will append an item.
func (sa *SafeArray) Append(v interface{}) {
	sa.m.Lock()
	defer sa.m.Unlock()

	if sa.limit == 0 {
	} else if len(sa.array) > sa.limit {
		sa.array = sa.array[1:]
	}

	sa.array = append(sa.array, v)
}

// Range will enumerate through all array items.
func (sa *SafeArray) Range(fn func(interface{}) bool) {
	sa.m.Lock()
	defer sa.m.Unlock()

	for _, v := range sa.array {
		if fn(v) {
			continue
		}

		break
	}
}

// MarshalJSON will marshall the array contents to JSON.
func (sa *SafeArray) MarshalJSON() ([]byte, error) {
	sa.m.Lock()
	defer sa.m.Unlock()

	return json.Marshal(sa.array)
}
