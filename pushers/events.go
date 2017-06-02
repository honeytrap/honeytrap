package pushers

import "github.com/honeytrap/honeytrap/pushers/event"

// EventBus defines a structure which provides a pubsub bus where message.Events
// are sent along it's wires for delivery
type EventBus struct {
	subscribers []Channel
}

// NewEventBus returns a new instance of a EventBus.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe adds the giving channel to the list of subscribers for the giving bus.
func (e *EventBus) Subscribe(channels ...Channel) {
	e.subscribers = append(e.subscribers, channels...)
}

// Send deliverers the slice of messages to all subscribers.
func (e *EventBus) Send(pm event.Event) {
	for _, subscriber := range e.subscribers {
		subscriber.Send(pm)
	}
}
