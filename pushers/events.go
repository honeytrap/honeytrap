package pushers

import (
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// Events defines an interface which exposes a method for the delivery of message.Event
// object.
type Events interface {
	Deliver(message.Event)
}

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
	a.Deliver(ev)
}

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
	ev.Time = time.Now()

	d.sync.Send([]*message.PushMessage{
		{
			Sensor:      ev.Sensor,
			Category:    ev.Category,
			SessionID:   ev.SessionID,
			ContainerID: ev.ContainerID,
			Data:        ev,
		},
	})
}
