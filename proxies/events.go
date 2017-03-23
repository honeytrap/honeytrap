package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ConnectionEvent defines a function to return an event object for related
// connection events.
func ConnectionEvent(ip net.Addr, ev message.EventType, sensor string, data interface{}) message.Event {
	return message.Event{
		Sensor:   sensor,
		Category: "Connections",
		Type:     ev,
		Data:     data,
		Details: map[string]interface{}{
			"addr": ip.String(),
		},
	}
}

// EventConnectionPing defines a function which returns a event object for a
// ping request with a connection.
func EventConnectionPing(ip net.Addr, sensor string, data error) message.Event {
	return ConnectionEvent(ip, message.Ping, sensor, data)
}

// EventConnectionError defines a function which returns a event object for a
// error occurence with a connection.
func EventConnectionError(ip net.Addr, sensor string, data error) message.Event {
	return ConnectionEvent(ip, message.ConnectionError, sensor, data)
}

// EventConnectionRequest defines a function which returns a event object for a
// request connection.
func EventConnectionRequest(ip net.Addr, sensor string, data interface{}) message.Event {
	return ConnectionEvent(ip, message.ConnectionRequest, sensor, data)
}

// EventConnectionClosed defines a function which returns a event object for a
// closed connection.
func EventConnectionClosed(ip net.Addr, sensor string, data interface{}) message.Event {
	return ConnectionEvent(ip, message.ConnectionClosed, sensor, data)
}

// EventConnectionOpened defines a function which returns a event object for a
// closed connection.
func EventConnectionOpened(ip net.Addr, sensor string, data interface{}) message.Event {
	return ConnectionEvent(ip, message.ConnectionStarted, sensor, data)
}
