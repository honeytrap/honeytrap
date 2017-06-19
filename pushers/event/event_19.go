// +build go1.9

package event

import (
	"sync/syncmap"
	"time"
)

// Event defines a object which adds key-value pairs into a map type for event data.
type Event struct {
	sm *syncmap.Map
}

// New returns a new Event with the options applied.
func New(opts ...Option) *Event {
	e := &Event{
		sm: new(syncmap.Map),
	}

	e.sm.Store("date", time.Now())

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Add adds the key and value into the event.
func (e *Event) Add(s string, v interface{}) {
	e.sm.Store(s, v)
}

// Map returns the underline map for the giving object.
func (e *Event) Map() Map {
	mp := make(map[string]interface{})

	e.sm.Range(func(key, value interface{}) {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
	})

	return Map(mp)
}

// Get retrieves a giving value for a key has string.
func (e *Event) Get(s string) string {
	if v, ok := e.sm.Load(s); !ok {
		return ""
	} else if v, ok := v.(string); !ok {
		return ""
	} else {
		return v
	}
}
