package event

import (
	"net"
	"time"
)

//====================================================================================

// EType defines a string type for all event types available.
type EType string

// contains different sets of possible events type.
const (
	PingEvent            = "PING"
	Operational          = "OPERATIONAL:Event"
	OperationalAuth      = "OPERATIONAL:AUTH"
	DataRequest          = "DATA:REQUEST"
	DataRead             = "DATA:READ"
	DataWrite            = "DATA:WRITE"
	ServiceEnded         = "SERVICE:ENDED"
	ServiceStarted       = "SERVICE:STARTED"
	ConnectionOpened     = "CONNECTION:OPENED"
	ConnectionClosed     = "CONNECTION:CLOSED"
	UserSessionOpened    = "SESSION:USER:OPENED"
	UserSessionClosed    = "SESSION:USER:CLOSED"
	ConnectionReadError  = "CONNECTION:ERROR:READ"
	ConnectionWriteError = "CONNECTION:ERROR:WRITE"
	ContainerStarted     = "CONTAINER:STARTED"
	ContainerFrozen      = "CONTAINER:FROZEN"
	ContainerDial        = "CONTAINER:DIAL"
	ContainerError       = "CONTAINER:ERROR"
	ContainerUnfrozen    = "CONTAINER:UNFROZEN"
	ContainerCloned      = "CONTAINER:CLONED"
	ContainerStopped     = "CONTAINER:STOPPED"
	ContainerPaused      = "CONTAINER:PAUSED"
	ContainerResumed     = "CONTAINER:RESUMED"
	ContainerTarred      = "CONTAINER:TARRED"
	ContainerCheckpoint  = "CONTAINER:CHECKPOINT"
	ContainerPcaped      = "CONTAINER:PCAPED"
)

//====================================================================================

// Contains a series of sensors constants.
const (
	ContainersSensor      = "CONTAINER"
	ConnectionSensor      = "CONNECTION"
	ServiceSensor         = "SERVICE"
	SessionSensor         = "SESSIONS"
	EventSensor           = "EVENTS"
	PingSensor            = "PING"
	DataSensor            = "DATA"
	ErrorsSensor          = "ERRORS"
	DataErrorSensor       = "DATA:ERROR"
	ConnectionErrorSensor = "CONNECTION:ERROR"
)

//====================================================================================

// ECategory defines a string type for for which is used to defined event category
// for different types.
type ECategory string

//====================================================================================

// Option defines a function type for events modifications.
type Option func(Event)

// Event defines a map type for event data.
type Event map[string]interface{}

// New returns a new Event with the options applied.
func New(opts ...Option) Event {
	e := map[string]interface{}{
		"date": time.Now(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return Event(e)
}

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
		m["token"] = token
	}
}

// Category returns an option for setting the category value.
func Category(s string) Option {
	return func(m Event) {
		m["category"] = ECategory(s)
	}
}

// Type returns an option for setting the type value.
func Type(s string) Option {
	return func(m Event) {
		m["type"] = EType(s)
	}
}

// Sensor returns an option for setting the sensor value.
func Sensor(s string) Option {
	return func(m Event) {
		m["sensor"] = s
	}
}

// SourceIP returns an option for setting the source-ip value.
func SourceIP(ip net.IP) Option {
	return func(m Event) {
		m["source-ip"] = ip.String()
	}
}

// DestinationIP returns an option for setting the destination-ip value.
func DestinationIP(ip net.IP) Option {
	return func(m Event) {
		m["destination-ip"] = ip.String()
	}
}

// RemoteAddr returns an option for setting the host-addr value.
func RemoteAddr(addr string) Option {
	return func(m Event) {
		m["remote-addr"] = addr
	}
}

// HostAddr returns an option for setting the host-addr value.
func HostAddr(addr string) Option {
	return func(m Event) {
		m["host-addr"] = addr
	}
}

// RemoteAddrFrom returns an option for setting the host-addr value.
func RemoteAddrFrom(addr net.Addr) Option {
	return func(m Event) {
		m["remote-addr"] = addr.String()
	}
}

// HostAddrFrom returns an option for setting the host-addr value.
func HostAddrFrom(addr net.Addr) Option {
	return func(m Event) {
		m["host-addr"] = addr.String()
	}
}

// SourcePort returns an option for setting the source-port value.
func SourcePort(port uint16) Option {
	return func(m Event) {
		m["source-port"] = port
	}
}

// DestinationPort returns an option for setting the destination-port value.
func DestinationPort(port uint16) Option {
	return func(m Event) {
		m["destination-port"] = port
	}
}

// Payload returns an option for setting the payload value.
func Payload(data []byte) Option {
	return func(m Event) {
		m["payload"] = string(data)
	}
}

// MergeFrom copies the internal key-value pair into the event if the event lacks the
// given key.
func MergeFrom(data map[string]interface{}) Option {
	return func(m Event) {
		for name, value := range data {
			if _, ok := m[name]; !ok {
				m[name] = value
			}
		}
	}
}

// CopyFrom copies the internal key-value pair into the event, overwritten any previous
// key's value if matching key.
func CopyFrom(data map[string]interface{}) Option {
	return func(m Event) {
		for name, value := range data {
			m[name] = value
		}
	}
}

// Custom returns an option for setting the custom key-value pair.
func Custom(name string, value interface{}) Option {
	return func(m Event) {
		m[name] = value
	}
}
