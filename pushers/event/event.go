package event

import (
	"net"
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

type Option func(Event)

type Event map[string]interface{}

func New(opts ...Option) Event {
	e := map[string]interface{}{
		"date": time.Now(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return Event(e)
}

func Category(s string) Option {
	return func(m Event) {
		m["category"] = message.EventCategory(s)
	}
}

func Sensor(s string) Option {
	return func(m Event) {
		m["sensor"] = s
	}
}

func SourceIP(ip net.IP) Option {
	return func(m Event) {
		m["source-ip"] = ip.String()
	}
}

func DestinationIP(ip net.IP) Option {
	return func(m Event) {
		m["destination-ip"] = ip.String()
	}
}

func SourcePort(port uint16) Option {
	return func(m Event) {
		m["source-port"] = port
	}
}

func DestinationPort(port uint16) Option {
	return func(m Event) {
		m["destination-port"] = port
	}
}

func Payload(data []byte) Option {
	return func(m Event) {
		m["payload"] = string(data)
	}
}

func Custom(name string, value interface{}) Option {
	return func(m Event) {
		m[name] = value
	}
}
