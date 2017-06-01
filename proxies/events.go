package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ServiceStartedEvent returns a connection open event object giving the associated data values.
func ServiceStartedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   meta,
		Sensor:    message.ServiceSensor,
		Type:      message.ServiceStarted,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
		Message:   "Service has started",
	}
}

// ServiceEndedEvent returns a connection open event object giving the associated data values.
func ServiceEndedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   meta,
		Sensor:    message.ServiceSensor,
		Type:      message.ServiceStarted,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
		Message:   "Service has ended",
	}
}

// UserSessionClosedEvent returns a connection open event object giving the associated data values.
func UserSessionClosedEvent(c net.Conn, data interface{}) message.Event {
	return message.Event{
		Data:      data,
		Sensor:    message.SessionSensor,
		Type:      message.UserSessionOpened,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Session has closed",
	}
}

// UserSessionOpenedEvent returns a connection open event object giving the associated data values.
func UserSessionOpenedEvent(c net.Conn, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Sensor:    message.SessionSensor,
		Type:      message.UserSessionClosed,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Details:   meta,
		Message:   "New Session has begun",
	}
}

// ConnectionOpenedEvent returns a connection open event object giving the associated data values.
func ConnectionOpenedEvent(c net.Conn) message.Event {
	return message.Event{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionOpened,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "New connection has started",
	}
}

// ConnectionClosedEvent returns a connection open event object giving the associated data values.
func ConnectionClosedEvent(c net.Conn) message.Event {
	return message.Event{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionClosed,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Connection has closed",
	}
}

// ConnectionWriteErrorEvent returns a connection open event object giving the associated data values.
func ConnectionWriteErrorEvent(c net.Conn, data error) message.Event {
	return message.Event{
		Data:      data,
		Sensor:    message.ConnectionErrorSensor,
		Type:      message.ConnectionWriteError,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Connection has faced write error",
	}
}

// ConnectionReadErrorEvent returns a connection open event object giving the associated data values.
func ConnectionReadErrorEvent(c net.Conn, data error) message.Event {
	return message.Event{
		Data:      data,
		Sensor:    message.ConnectionErrorSensor,
		Type:      message.ConnectionReadError,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Connection has faced read error",
	}
}

// ListenerClosedEvent returns a connection open event object giving the associated data values.
func ListenerClosedEvent(c net.Listener) message.Event {
	return message.Event{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionClosed,
		HostAddr:  c.Addr().String(),
		LocalAddr: c.Addr().String(),
		Message:   "Listener has being closed",
	}
}

// ListenerOpenedEvent returns a connection open event object giving the associated data values.
func ListenerOpenedEvent(c net.Listener, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Details:   meta,
		Data:      data,
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionOpened,
		HostAddr:  c.Addr().String(),
		LocalAddr: c.Addr().String(),
		Message:   "Listener has being open",
	}
}

// AgentRequestEvent defines a function which returns a event object for a
// request connection.
func AgentRequestEvent(addr net.Addr, session string, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		SessionID: session,
		Sensor:    message.DataSensor,
		Type:      message.DataRequest,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
		Details:   detail,
		Message:   "Service Agent has report an event",
	}
}

// DataRequest returns a connection open event object giving the associated data values.
func DataRequest(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataRequest,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Data request has initiated",
	}
}

// OperationEvent returns a connection open event object giving the associated data values.
func OperationEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   detail,
		Sensor:    message.EventSensor,
		Type:      message.Operational,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Operation request has occured",
	}
}

// AuthEvent returns a connection open event object giving the associated data values.
func AuthEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   detail,
		Sensor:    message.EventSensor,
		Type:      message.OperationalAuth,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Authentication request has occured",
	}
}

// DataReadEvent returns a connection open event object giving the associated data values.
func DataReadEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataRead,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Data read request has occured",
	}
}

// DataWriteEvent returns a connection open event object giving the associated data values.
func DataWriteEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.Event{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataWrite,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Data write request has occured",
	}
}

// PingEvent returns a connection open event object giving the associated data values.
func PingEvent(c net.Conn, data interface{}) message.Event {
	return message.Event{
		Data:      data,
		Sensor:    message.PingSensor,
		Type:      message.PingEvent,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Message:   "Ping has being sent",
	}
}
