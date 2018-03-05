/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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
	PingEvent       = Type("PING")
	Operational     = Type("OPERATIONAL:Event")
	OperationalAuth = Type("OPERATIONAL:AUTH")
	DataRequest     = Type("DATA:REQUEST")
	DataRead        = Type("DATA:READ")
	DataWrite       = Type("DATA:WRITE")
	ServiceEnded    = Type("SERVICE:ENDED")

	SeverityFatal = Type("fatal")
	SeverityError = Type("error")
	SeverityInfo  = Type("info")

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
func SourceHardwareAddr(addr net.HardwareAddr) Option {
	return func(m Event) {
		m.Store("source-mac", addr.String())
	}
}

// DestinationAddr returns an option for setting the destination-ip value.
func DestinationHardwareAddr(addr net.HardwareAddr) Option {
	return func(m Event) {
		m.Store("destination-mac", addr.String())
	}
}

// SourceAddr returns an option for setting the source-ip value.
func SourceAddr(addr net.Addr) Option {
	return func(m Event) {
		if ta, ok := addr.(*net.TCPAddr); ok {
			m.Store("source-ip", ta.IP.String())
			m.Store("source-port", ta.Port)
		} else if ua, ok := addr.(*net.UDPAddr); ok {
			m.Store("source-ip", ua.IP.String())
			m.Store("source-port", ua.Port)
		}
	}
}

// DestinationAddr returns an option for setting the destination-ip value.
func DestinationAddr(addr net.Addr) Option {
	return func(m Event) {
		if ta, ok := addr.(*net.TCPAddr); ok {
			m.Store("destination-ip", ta.IP.String())
			m.Store("destination-port", ta.Port)
		} else if ua, ok := addr.(*net.UDPAddr); ok {
			m.Store("destination-ip", ua.IP.String())
			m.Store("destination-port", ua.Port)
		}
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

// Service sets the service of the event
func Service(v string) Option {
	return func(m Event) {
		m.Store("service", v)
	}
}

// Protocol sets the protocol of the event
func Protocol(v string) Option {
	return func(m Event) {
		m.Store("protocol", v)
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

// ToMap returns a map containing all available data which map
// a string key and value type.
func ToMap(ev Event) map[string]interface{} {
	mp := make(map[string]interface{})

	ev.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	return mp
}
