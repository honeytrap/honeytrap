package pushers

import (
	"github.com/honeytrap/honeytrap/pushers/message"
)

// EventBus defines a structure which provides a pubsub bus where message.Events
// are sent along it's wires for delivery
type EventBus struct {
	subs []Channel
}

// NewEventBus returns a new instance of a EventBus.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe adds the giving channel to the list of subscribers for the giving bus.
func (e *EventBus) Subscribe(channel Channel) {
	e.subs = append(e.subs, channel)
}

// Send deliverers the slice of messages to all subscribers.
func (e *EventBus) Send(pm ...message.Event) {
	for _, bus := range e.subs[:len(e.subs)] {
		bus.Send(pm...)
	}
}
