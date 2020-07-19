package nscanary

import (
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

type EventConn struct {
	net.Conn

	events    pushers.Channel
	connEvent event.Option
}

func NewEventConn(conn net.Conn, events pushers.Channel) *EventConn {

	p := &EventConn{
		Conn:   conn,
		events: events,
		connEvent: event.NewWith(
			CanaryOptions,
			event.Category("tcp"),
			event.SourceAddr(conn.LocalAddr()),
			event.DestinationAddr(conn.RemoteAddr()),
		),
	}

	return p
}

func (c *EventConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if err != nil {
		return n, err
	}

	payload := make([]byte, n)
	copy(payload, p)

	c.events.Send(event.New(
		c.connEvent,
		event.Payload(payload)),
	)

	return n, nil
}
