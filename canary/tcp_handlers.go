package canary

import (
	"bufio"
	"bytes"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/tcp"
	"github.com/honeytrap/honeytrap/pushers/event"
)

var (
	// EventCategoryTCP contains events for ssdp traffic
	EventCategoryTCP = event.Category("tcp")
)

// EventTCPPayload will return a snmp event struct
func EventTCPPayload(sourceIP net.IP, port uint16, payload string) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface
	return event.New(
		CanaryOptions,
		EventCategoryHTTP,
		event.Type(event.ServiceStarted),
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("tcp.port", port),
		event.Custom("tcp.payload", payload),
		event.Custom("tcp.length", len(payload)),
	)
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryHTTP = event.Category("http")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeHTTP(iph *ipv4.Header, tcph *tcp.Header) error {
	request, err := http.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(tcph.Payload),
		),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(EventHTTP(iph.Src, request.Method, request.RequestURI, request.Proto, request.Header))
	return nil
}

// EventHTTP will return a snmp event struct
func EventHTTP(sourceIP net.IP, method, uri, proto string, headers http.Header) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface
	return event.New(
		CanaryOptions,
		EventCategoryHTTP,
		event.Type(event.ServiceStarted),
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("http.method", method),
		event.Custom("http.uri", uri),
		event.Custom("http.proto", proto),
		event.Custom("http.headers", headers),

		event.Custom("http.host", headers.Get("Host")),
		event.Custom("http.content-type", headers.Get("Content-Type")),
		event.Custom("http.user-agent", headers.Get("User-Agent")),
	)
}

// port 139 -> http://s11.invisionfree.com/dongsongbang/ar/t170.htm
