// +build !go1.9

package event

import (
	"time"

	"golang.org/x/sync/syncmap"
)

// Event defines a object which adds key-value pairs into a map type for event data.
type Event struct {
	sm *syncmap.Map
}

// New returns a new Event with the options applied.
func New(opts ...Option) Event {
	e := Event{
		sm: new(syncmap.Map),
	}

	e.sm.Store("date", time.Now())

	for _, opt := range opts {
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
