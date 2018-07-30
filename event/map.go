/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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

	e.sm.Store("date", time.Now())

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
