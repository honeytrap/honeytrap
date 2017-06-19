// +build !go1.9

package event

import (
	"sync"
	"time"
)

// Event defines a object which adds key-value pairs into a map type for event data.
type Event struct {
	ml sync.Mutex
	mp map[string]interface{}
}

// New returns a new Event with the options applied.
func New(opts ...Option) *Event {
	e := &Event{
		mp: map[string]interface{}{
			"date": time.Now(),
		},
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Add adds the key and value into the event.
func (e *Event) Add(s string, v interface{}) {
	e.ml.Lock()
	e.mp[s] = v
	e.ml.Unlock()
}

// Map returns the underline map for the giving object.
func (e *Event) Map() Map {
	var em Map

	e.ml.Lock()
	{
		em = e.mp
		e.mp = make(map[string]interface{})
	}
	e.ml.Unlock()

	return em
}

// Get retrieves a giving value for a key has string.
func (e *Event) Get(s string) string {
	e.ml.Lock()
	defer e.ml.Unlock()

	if v, ok := e.mp[s]; !ok {
		return ""
	} else if v, ok := v.(string); !ok {
		return ""
	} else {
		return v
	}
}
