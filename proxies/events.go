package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/pushers/event"
)

// ServiceStartedEvent returns a connection open event object giving the associated data values.
func ServiceStartedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ServiceStarted),
		event.Sensor(event.ServiceSensor),
		event.Custom("data", data),
		event.HostAddr(addr.String()),
		event.CopyFrom(meta),
	)
}

// ServiceEndedEvent returns a connection open event object giving the associated data values.
func ServiceEndedEvent(addr net.Addr, data interface{}, meta map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ServiceEnded),
		event.Sensor(event.ServiceSensor),
		event.Custom("data", data),
		event.HostAddr(addr.String()),
		event.CopyFrom(meta),
	)
}

// UserSessionClosedEvent returns a connection open event object giving the associated data values.
func UserSessionClosedEvent(c net.Conn, data interface{}) event.Event {
	return event.New(
		event.Type(event.UserSessionOpened),
		event.Sensor(event.SessionSensor),
		event.Custom("data", data),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// UserSessionOpenedEvent returns a connection open event object giving the associated data values.
func UserSessionOpenedEvent(c net.Conn, data interface{}, meta map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.UserSessionClosed),
		event.Sensor(event.SessionSensor),
		event.Custom("data", data),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
		event.CopyFrom(meta),
	)
}

// ConnectionOpenedEvent returns a connection open event object giving the associated data values.
func ConnectionOpenedEvent(c net.Conn) event.Event {
	return event.New(
		event.Type(event.ConnectionOpened),
		event.Sensor(event.ConnectionSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
		// event.CopyFrom(meta),
	)
}

// ConnectionClosedEvent returns a connection open event object giving the associated data values.
func ConnectionClosedEvent(c net.Conn) event.Event {
	return event.New(
		event.Type(event.ConnectionClosed),
		event.Sensor(event.ConnectionSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
		// event.CopyFrom(meta),
	)
}

// ConnectionWriteErrorEvent returns a connection open event object giving the associated data values.
func ConnectionWriteErrorEvent(c net.Conn, data error) event.Event {
	return event.New(
		event.Type(event.ConnectionWriteError),
		event.Sensor(event.ConnectionErrorSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
		event.Custom("error", data),
	)
}

// ConnectionReadErrorEvent returns a connection open event object giving the associated data values.
func ConnectionReadErrorEvent(c net.Conn, data error) event.Event {
	return event.New(
		event.Type(event.ConnectionReadError),
		event.Sensor(event.ConnectionErrorSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
		event.Custom("error", data),
	)
}

// ListenerClosedEvent returns a connection open event object giving the associated data values.
func ListenerClosedEvent(c net.Listener) event.Event {
	return event.New(
		event.Type(event.ConnectionClosed),
		event.Sensor(event.ConnectionSensor),
		event.HostAddrFrom(c.Addr()),
	)
}

// ListenerOpenedEvent returns a connection open event object giving the associated data values.
func ListenerOpenedEvent(c net.Listener, data interface{}, meta map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(meta),
		event.Type(event.ConnectionOpened),
		event.Sensor(event.ConnectionSensor),
		event.HostAddrFrom(c.Addr()),
		event.Custom("data", data),
	)
}

// AgentRequestEvent defines a function which returns a event object for a
// request connection.
func AgentRequestEvent(addr net.Addr, session string, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Type(event.DataRequest),
		event.Sensor(event.DataSensor),
		event.HostAddrFrom(addr),
		event.Custom("data", data),
		event.Custom("session-id", session),
	)
}

// DataRequest returns a connection open event object giving the associated data values.
func DataRequest(c net.Conn, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Custom("data", data),
		event.Type(event.DataRequest),
		event.Sensor(event.DataSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// OperationEvent returns a connection open event object giving the associated data values.
func OperationEvent(c net.Conn, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Custom("data", data),
		event.Type(event.Operational),
		event.Sensor(event.EventSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// AuthEvent returns a connection open event object giving the associated data values.
func AuthEvent(c net.Conn, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Custom("data", data),
		event.Type(event.OperationalAuth),
		event.Sensor(event.EventSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// DataReadEvent returns a connection open event object giving the associated data values.
func DataReadEvent(c net.Conn, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Custom("data", data),
		event.Type(event.DataRead),
		event.Sensor(event.DataSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// DataWriteEvent returns a connection open event object giving the associated data values.
func DataWriteEvent(c net.Conn, data interface{}, detail map[string]interface{}) event.Event {
	return event.New(
		event.CopyFrom(detail),
		event.Custom("data", data),
		event.Type(event.DataWrite),
		event.Sensor(event.DataSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}

// PingEvent returns a connection open event object giving the associated data values.
func PingEvent(c net.Conn, data interface{}) event.Event {
	return event.New(
		event.Custom("data", data),
		event.Type(event.PingEvent),
		event.Sensor(event.PingSensor),
		event.HostAddrFrom(c.LocalAddr()),
		event.RemoteAddrFrom(c.RemoteAddr()),
	)
}
