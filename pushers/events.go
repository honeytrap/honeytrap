package pushers

import (
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

//================================================================================

// Events defines an interface which exposes a method for the delivery of message.Event
// object.
type Events interface {
	Deliver(message.Event)
}

// EventStream defines a type for a slice of Events implementing objects.
type EventStream []Events

// Deliver delivers the provided events to all underline set of Events implementing
// objects.
func (eset EventStream) Deliver(ev message.Event) {
	for _, es := range eset {
		es.Deliver(ev)
	}
}

//================================================================================

// TokenedEventDelivery defines a custom event delivery type which wraps the
// EventDelivery and sets the internal token value for the events passed in.
type TokenedEventDelivery struct {
	*EventDelivery
	Token string
}

// NewTokenedEventDelivery returns a new TokenedEventDelivery instanc.
func NewTokenedEventDelivery(token string, channel Channel) *TokenedEventDelivery {
	return &TokenedEventDelivery{
		EventDelivery: NewEventDelivery(channel),
		Token:         token,
	}
}

// Deliver delivers the underline event object to the underline EventDelivery
// object.
func (a TokenedEventDelivery) Deliver(ev message.Event) {
	ev.Token = a.Token
	a.EventDelivery.Deliver(ev)
}

//================================================================================

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
func (e *EventBus) Send(pm []message.PushMessage) {
	for _, bus := range e.subs[:len(e.subs)] {
		bus.Send(pm)
	}
}

//================================================================================

// EventDelivery defines a struct which embodies a delivery system which allows
// events to be piped down to a pusher library.
type EventDelivery struct {
	sync Channel
}

// NewEventDelivery returns a new EventDelivery instance which is used to deliver
// events.
func NewEventDelivery(channel Channel) *EventDelivery {
	return &EventDelivery{sync: channel}
}

// Deliver adds the giving event into the provided message.Channel for the delivery
func (d *EventDelivery) Deliver(ev message.Event) {
	// Set the time for the event
	ev.Date = time.Now()

	if ev.Location == "" {
		ev.Location = "Unknown"
	}

	d.sync.Send([]message.PushMessage{
		{
			Event:       true,
			Sensor:      ev.Sensor,
			Category:    string(ev.Type),
			SessionID:   ev.SessionID,
			ContainerID: ev.ContainerID,
			Data:        ev,
		},
	})
}
