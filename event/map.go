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
package event

import (
	"encoding/json"
	"sync"
	"time"
)

// Event defines a object which adds key-value pairs into a map type for event data.
type Event struct {
	sm *sync.Map
}

func (e Event) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			m[keyName] = value
		}
		return true
	})

	return json.Marshal(m)
}

// New returns a new Event with the options applied.
func New(opts ...Option) Event {
	e := Event{
		sm: new(sync.Map),
	}

	e.sm.Store("date", time.Now().Format(time.RFC3339))

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(e)
	}

	return e
}

// Range defines a function which ranges the underline key-values with
// the provided syncmap.
func (e Event) Range(fx func(interface{}, interface{}) bool) {
	e.sm.Range(fx)
}

// Store adds the key and value into the event.
func (e Event) Store(s string, v interface{}) {
	e.sm.Store(s, v)
}

// Has returns true/false if the giving key exists.
func (e Event) Has(s string) bool {
	_, ok := e.sm.Load(s)
	return ok
}

// Get retrieves a giving value for a key has string.
func (e Event) Get(s string) string {
	if v, ok := e.sm.Load(s); !ok {
		return ""
	} else if v, ok := v.(string); !ok {
		return ""
	} else {
		return v
	}
}
