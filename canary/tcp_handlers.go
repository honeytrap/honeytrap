package canary

import (
	"bufio"
	"bytes"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/tcp"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategoryTCP = message.EventCategory("tcp")
)

const (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryHTTP = message.EventCategory("http")
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
func EventHTTP(sourceIP net.IP, method, uri, proto string, headers http.Header) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.NewEvent("canary", EventCategoryHTTP, message.ServiceStarted, map[string]interface{}{
		"source-ip": sourceIP.String(),

		"http.method":  method,
		"http.uri":     uri,
		"http.proto":   proto,
		"http.headers": headers,

		"http.host":         headers.Get("Host"),
		"http.content-type": headers.Get("Content-Type"),
		"http.user-agent":   headers.Get("User-Agent"),
	})
}
