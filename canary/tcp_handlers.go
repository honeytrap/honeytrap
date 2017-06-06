package canary

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/pushers/event"
)

var (
	// EventCategoryTCP contains events for ssdp traffic
	EventCategoryTCP = event.Category("tcp")
)

// EventTCPPayload will return a snmp event struct
func EventTCPPayload(src, dst net.IP, srcport, dstport uint16, payload []byte) event.Event {
	return event.New(
		CanaryOptions,
		EventCategoryTCP,
		event.ServiceStarted,
		event.SourceIP(src),
		event.DestinationIP(dst),
		event.SourcePort(srcport),
		event.DestinationPort(dstport),
		event.Payload(payload),
	)
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryHTTP = event.Category("http")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeHTTP(conn net.Conn) error {
	defer conn.Close()

	request, err := http.ReadRequest(
		bufio.NewReader(conn),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryHTTP,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.method", request.Method),
		event.Custom("http.uri", request.URL.String()),
		event.Custom("http.proto", request.Proto),
		event.Custom("http.headers", request.Header),
		event.Custom("http.host", request.Header.Get("Host")),
		event.Custom("http.content-type", request.Header.Get("Content-Type")),
		event.Custom("http.user-agent", request.Header.Get("User-Agent")),
	))

	fmt.Printf("%+v", request)
	return nil
}

// port 139 -> http://s11.invisionfree.com/dongsongbang/ar/t170.htm
