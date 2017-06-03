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
func (eb *EventBus) Subscribe(channels ...Channel) {
	eb.subscribers = append(eb.subscribers, channels...)
}

// Send deliverers the slice of messages to all subscribers.
func (eb *EventBus) Send(e event.Event) {
	for _, subscriber := range eb.subscribers {
		subscriber.Send(e)
	}
}
