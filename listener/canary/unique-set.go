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
package canary

type equalFn func(interface{}, interface{}) bool

// NewUniqueSet returns a new instance of UniqueSet.
func NewUniqueSet(fn equalFn) *UniqueSet {
	return &UniqueSet{
		uniqueFunc: fn,
	}
}

// UniqueSet defines a type to create a unique set of values.
type UniqueSet struct {
	items []interface{}

	uniqueFunc equalFn
}

// Count returns count of all elements.
func (us *UniqueSet) Count() int {
	return len(us.items)
}

// Add adds the given item into the set if it not yet included.
func (us *UniqueSet) Add(item interface{}) interface{} {
	for _, item2 := range us.items {
		if !us.uniqueFunc(item, item2) {
			continue
		}

		return item2
	}

	us.items = append(us.items, item)
	return item
}

// Remove removes the item from the interal set.
func (us *UniqueSet) Remove(item interface{}) {
	for i, item2 := range us.items {
		if item != item2 {
			continue
		}

		us.items[i] = nil

		us.items = append(us.items[:i], us.items[i+1:]...)
		return
	}
}

// Each runs the function against all items in set.
func (us *UniqueSet) Each(fn func(int, interface{})) {
	items := us.items[:]

	for i, item := range items {
		fn(i, item)
	}
}

// Find runs the function against all items in set.
func (us *UniqueSet) Find(fn func(interface{}) bool) interface{} {
	for _, item := range us.items {
		if fn(item) {
			return item
		}
	}

	return nil
}
