package event

import (
	"encoding/hex"
	"fmt"
	"net"
	"runtime/debug"
)

//====================================================================================

// contains different sets of possible events type.
var (
	PingEvent            = Type("PING")
	Operational          = Type("OPERATIONAL:Event")
	OperationalAuth      = Type("OPERATIONAL:AUTH")
	DataRequest          = Type("DATA:REQUEST")
	DataRead             = Type("DATA:READ")
	DataWrite            = Type("DATA:WRITE")
	ServiceEnded         = Type("SERVICE:ENDED")
	SeverityFatal        = Type("fatal")
	SeverityError        = Type("error")
	ServiceStarted       = Type("SERVICE:STARTED")
	ConnectionOpened     = Type("CONNECTION:OPENED")
	ConnectionClosed     = Type("CONNECTION:CLOSED")
	UserSessionOpened    = Type("SESSION:USER:OPENED")
	UserSessionClosed    = Type("SESSION:USER:CLOSED")
	ConnectionReadError  = Type("CONNECTION:ERROR:READ")
	ConnectionWriteError = Type("CONNECTION:ERROR:WRITE")
	ContainerStarted     = Type("CONTAINER:STARTED")
	ContainerFrozen      = Type("CONTAINER:FROZEN")
	ContainerDial        = Type("CONTAINER:DIAL")
	ContainerError       = Type("CONTAINER:ERROR")
	ContainerUnfrozen    = Type("CONTAINER:UNFROZEN")
	ContainerCloned      = Type("CONTAINER:CLONED")
	ContainerStopped     = Type("CONTAINER:STOPPED")
	ContainerPaused      = Type("CONTAINER:PAUSED")
	ContainerResumed     = Type("CONTAINER:RESUMED")
	ContainerTarred      = Type("CONTAINER:TARRED")
	ContainerCheckpoint  = Type("CONTAINER:CHECKPOINT")
	ContainerPcaped      = Type("CONTAINER:PCAPED")
)

//====================================================================================

// Contains a series of sensors variables.
var (
	ContainersSensorName      = "CONTAINER"
	ConnectionSensorName      = "CONNECTION"
	ServiceSensorName         = "SERVICE"
	SessionSensorName         = "SESSIONS"
	EventSensorName           = "EVENTS"
	PingSensorName            = "PING"
	DataSensorName            = "DATA"
	ErrorsSensorName          = "ERRORS"
	DataErrorSensorName       = "DATA:ERROR"
	ConnectionErrorSensorName = "CONNECTION:ERROR"

	ContainersSensor      = Sensor("CONTAINER")
	ConnectionSensor      = Sensor("CONNECTION")
	ServiceSensor         = Sensor("SERVICE")
	SessionSensor         = Sensor("SESSIONS")
	EventSensor           = Sensor("EVENTS")
	PingSensor            = Sensor("PING")
	DataSensor            = Sensor("DATA")
	ErrorsSensor          = Sensor("ERRORS")
	DataErrorSensor       = Sensor("DATA:ERROR")
	ConnectionErrorSensor = Sensor("CONNECTION:ERROR")
)

//====================================================================================

// Option defines a function type for events modifications.
type Option func(Event)

// Apply applies all options to the Event returning it after it's done.
func Apply(e Event, opts ...Option) Event {
	for _, option := range opts {
		option(e)
	}

	return e
}

// NewWith combines the set of option into a single option which
// applies all the series when called.
func NewWith(opts ...Option) Option {
	return func(e Event) {
		for _, option := range opts {
			option(e)
		}
	}
}

// Token adds the provided token into the giving Event.
func Token(token string) Option {
	return func(m Event) {
		m.Store("token", token)
	}
}

// Category returns an option for setting the category value.
func Category(s string) Option {
	return func(m Event) {
		m.Store("category", s)
	}
}

// Error returns an option for setting the error value.
func Error(err error) Option {
	return func(m Event) {
		m.Store("error", err)
	}
}

// Type returns an option for setting the type value.
func Type(s string) Option {
	return func(m Event) {
		m.Store("type", s)
	}
}

// Sensor returns an option for setting the sensor value.
func Sensor(s string) Option {
	return func(m Event) {
		m.Store("sensor", s)
	}
}

// SourceAddr returns an option for setting the source-ip value.
func SourceAddr(addr net.Addr) Option {
	return func(m Event) {
		m.Store("source-ip", addr.(*net.TCPAddr).IP.String())
		m.Store("source-port", addr.(*net.TCPAddr).Port)
	}
}

// DestinationAddr returns an option for setting the destination-ip value.
func DestinationAddr(addr net.Addr) Option {
	return func(m Event) {
		m.Store("destination-ip", addr.(*net.TCPAddr).IP.String())
		m.Store("destination-port", addr.(*net.TCPAddr).Port)
	}
}

// SourceIP returns an option for setting the source-ip value.
func SourceIP(ip net.IP) Option {
	return func(m Event) {
		m.Store("source-ip", ip.String())
	}
}

// DestinationIP returns an option for setting the destination-ip value.
func DestinationIP(ip net.IP) Option {
	return func(m Event) {
		m.Store("destination-ip", ip.String())
	}
}

// RemoteAddr returns an option for setting the host-addr value.
func RemoteAddr(addr string) Option {
	return func(m Event) {
		m.Store("remote-addr", addr)
	}
}

// HostAddr returns an option for setting the host-addr value.
func HostAddr(addr string) Option {
	return func(m Event) {
		m.Store("host-addr", addr)
	}
}

// RemoteAddrFrom returns an option for setting the host-addr value.
func RemoteAddrFrom(addr net.Addr) Option {
	return func(m Event) {
		m.Store("remote-addr", addr.String())
	}
}

// HostAddrFrom returns an option for setting the host-addr value.
func HostAddrFrom(addr net.Addr) Option {
	return func(m Event) {
		m.Store("host-addr", addr.String())
	}
}

// SourcePort returns an option for setting the source-port value.
func SourcePort(port uint16) Option {
	return func(m Event) {
		m.Store("source-port", port)
	}
}

// DestinationPort returns an option for setting the destination-port value.
func DestinationPort(port uint16) Option {
	return func(m Event) {
		m.Store("destination-port", port)
	}
}

// Message returns an option for setting the payload value.
// should this be just a formatter? eg Bla Bla {src-ip}
func Message(format string, a ...interface{}) Option {
	return func(m Event) {
		m.Store("message", fmt.Sprintf(format, a...))
	}
}

// Stack returns a stacktrace
func Stack() Option {
	return func(m Event) {
		data := debug.Stack()
		m.Store("stacktrace", string(data))
	}
}

// Payload returns an option for setting the payload value.
func Payload(data []byte) Option {
	return func(m Event) {
		m.Store("payload", string(data))
		m.Store("payload-hex", hex.EncodeToString(data))
		m.Store("payload-length", len(data))
	}
}

// MergeFrom copies the internal key-value pair into the event if the event lacks the
// given key.
func MergeFrom(data map[string]interface{}) Option {
	return func(m Event) {
		for name, value := range data {
			if !m.Has(name) {
				m.Store(name, value)
			}
		}
	}
}

// CopyFrom copies the internal key-value pair into the event, overwritten any previous
// key's value if matching key.
func CopyFrom(data map[string]interface{}) Option {
	return func(m Event) {
		for name, value := range data {
			m.Store(name, value)
		}
	}
}

// Custom returns an option for setting the custom key-value pair.
func Custom(name string, value interface{}) Option {
	return func(m Event) {
		m.Store(name, value)
	}
}
