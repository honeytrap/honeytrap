package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ServiceStartedEvent returns a connection open event object giving the associated data values.
func ServiceStartedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   meta,
		Sensor:    message.ServiceSensor,
		Type:      message.ServiceStarted,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
	}
}

// ServiceEndedEvent returns a connection open event object giving the associated data values.
func ServiceEndedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   meta,
		Sensor:    message.ServiceSensor,
		Type:      message.ServiceStarted,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
	}
}

// UserSessionClosedEvent returns a connection open event object giving the associated data values.
func UserSessionClosedEvent(c net.Conn, data interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Sensor:    message.SessionSensor,
		Type:      message.UserSessionOpened,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// UserSessionOpenedEvent returns a connection open event object giving the associated data values.
func UserSessionOpenedEvent(c net.Conn, data interface{}, meta map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Sensor:    message.SessionSensor,
		Type:      message.UserSessionClosed,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
		Details:   meta,
	}
}

// ConnectionOpenedEvent returns a connection open event object giving the associated data values.
func ConnectionOpenedEvent(c net.Conn) message.Event {
	return message.BasicEvent{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionOpened,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// ConnectionClosedEvent returns a connection open event object giving the associated data values.
func ConnectionClosedEvent(c net.Conn) message.Event {
	return message.BasicEvent{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionClosed,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// ConnectionWriteErrorEvent returns a connection open event object giving the associated data values.
func ConnectionWriteErrorEvent(c net.Conn, data error) message.Event {
	return message.BasicEvent{
		Data:      data,
		Sensor:    message.ConnectionErrorSensor,
		Type:      message.ConnectionWriteError,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// ConnectionReadErrorEvent returns a connection open event object giving the associated data values.
func ConnectionReadErrorEvent(c net.Conn, data error) message.Event {
	return message.BasicEvent{
		Data:      data,
		Sensor:    message.ConnectionErrorSensor,
		Type:      message.ConnectionReadError,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// ListenerClosedEvent returns a connection open event object giving the associated data values.
func ListenerClosedEvent(c net.Listener) message.Event {
	return message.BasicEvent{
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionClosed,
		HostAddr:  c.Addr().String(),
		LocalAddr: c.Addr().String(),
	}
}

// ListenerOpenedEvent returns a connection open event object giving the associated data values.
func ListenerOpenedEvent(c net.Listener, data interface{}, meta map[string]interface{}) message.Event {
	return message.BasicEvent{
		Details:   meta,
		Data:      data,
		Sensor:    message.ConnectionSensor,
		Type:      message.ConnectionOpened,
		HostAddr:  c.Addr().String(),
		LocalAddr: c.Addr().String(),
	}
}

// AgentRequestEvent defines a function which returns a event object for a
// request connection.
func AgentRequestEvent(addr net.Addr, session string, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		SessionID: session,
		Sensor:    message.DataSensor,
		Type:      message.DataRequest,
		HostAddr:  addr.String(),
		LocalAddr: addr.String(),
		Details:   detail,
	}
}

// DataRequest returns a connection open event object giving the associated data values.
func DataRequest(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataRequest,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// OperationEvent returns a connection open event object giving the associated data values.
func OperationEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   detail,
		Sensor:    message.EventSensor,
		Type:      message.Operational,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// AuthEvent returns a connection open event object giving the associated data values.
func AuthEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   detail,
		Sensor:    message.EventSensor,
		Type:      message.OperationalAuth,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// DataReadEvent returns a connection open event object giving the associated data values.
func DataReadEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataRead,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// DataWriteEvent returns a connection open event object giving the associated data values.
func DataWriteEvent(c net.Conn, data interface{}, detail map[string]interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Details:   detail,
		Sensor:    message.DataSensor,
		Type:      message.DataWrite,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}

// PingEvent returns a connection open event object giving the associated data values.
func PingEvent(c net.Conn, data interface{}) message.Event {
	return message.BasicEvent{
		Data:      data,
		Sensor:    message.PingSensor,
		Type:      message.PingEvent,
		HostAddr:  c.RemoteAddr().String(),
		LocalAddr: c.LocalAddr().String(),
	}
}
