package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ConnectionEvent defines a function to return an event object for related
// connection events.
func ConnectionEvent(host, local net.Addr, ev message.EventType, sensor string, data interface{}, dt map[string]interface{}) message.Event {
	return message.Event{
		Sensor:    sensor,
		Category:  "Connections",
		Type:      ev,
		Data:      data,
		HostAddr:  host.String(),
		LocalAddr: local.String(),
		Details:   dt,
	}
}

// AgentRequestEvent defines a function which returns a event object for a
// request connection.
func AgentRequestEvent(host, local net.Addr, sensor string, session string, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		SessionID: session,
		Sensor:    sensor,
		Category:  "agent-request",
		Type:      message.Ping,
		HostAddr:  host.String(),
		LocalAddr: local.String(),
		Details:   detail,
	}
}

// EventConnectionError defines a function which returns a event object for a
// error occurence with a connection.
func EventConnectionError(ip, local net.Addr, sensor string, data error, dt map[string]interface{}) message.Event {
	return ConnectionEvent(ip, local, message.ConnectionError, sensor, data, dt)
}

// EventConnectionClosed defines a function which returns a event object for a
// closed connection.
func EventConnectionClosed(ip, local net.Addr, sensor string, data interface{}, detail map[string]interface{}) message.Event {
	return ConnectionEvent(ip, local, message.ConnectionClosed, sensor, data, detail)
}

// EventConnectionOpened defines a function which returns a event object for a
// closed connection.
func EventConnectionOpened(ip, local net.Addr, sensor string, data interface{}, detail map[string]interface{}) message.Event {
	return ConnectionEvent(ip, local, message.ConnectionStarted, sensor, data, detail)
}
