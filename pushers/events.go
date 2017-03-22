package pushers

import (
	"github.com/honeytrap/honeytrap/pushers/message"
)

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
