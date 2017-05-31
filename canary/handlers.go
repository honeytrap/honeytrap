package canary

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"

	"github.com/google/gopacket/layers"
	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/udp"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategoryUDP = message.EventCategory1("udp")
)

// EventSSDP will return a snmp event struct
func EventUDP(sourceIP net.IP, port uint16, payload string) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategoryUDP,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"source-ip": sourceIP.String(),
			"port":      port,
			"payload":   payload,
		},
	}
}

const (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategorySSDP = message.EventCategory1("ssdp")
)

// DecodeSSDP will decode NTP packets
func (c *Canary) DecodeSSDP(iph *ipv4.Header, udph *udp.Header) error {
	request, err := http.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(udph.Payload),
		),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(EventSSDP(iph.Src, request.Method, request.RequestURI, request.Proto, request.Header))
	return nil
}

// EventSSDP will return a snmp event struct
func EventSSDP(sourceIP net.IP, method, uri, proto string, headers http.Header) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategorySSDP,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"source-ip": sourceIP.String(),

			"method":  method,
			"uri":     uri,
			"proto":   proto,
			"headers": headers,

			"HOST": headers.Get("HOST"),
			"MAN":  headers.Get("MAN"),
			"MX":   headers.Get("MX"),
			"ST":   headers.Get("ST"),
		},
	}
}

const (
	// EventCategorySIP contains events for ntp traffic
	EventCategorySIP = message.EventCategory1("sip")
)

// DecodeSIP will decode NTP packets
func (c *Canary) DecodeSIP(iph *ipv4.Header, udph *udp.Header) error {
	request, err := http.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(udph.Payload),
		),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(EventSIP(iph.Src, request.Method, request.RequestURI, request.Proto, request.Header))

	return nil
}

// EventSIP will return a snmp event struct
func EventSIP(sourceIP net.IP, method, uri, proto string, headers http.Header) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategorySIP,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"source-ip":  sourceIP.String(),
			"method":     method,
			"uri":        uri,
			"proto":      proto,
			"headers":    headers,
			"from":       headers.Get("From"),
			"to":         headers.Get("To"),
			"via":        headers.Get("Via"),
			"contact":    headers.Get("Contact"),
			"call-id":    headers.Get("Call-ID"),
			"user-agent": headers.Get("User-Agent"),
		},
	}
}

const (
	// EventCategorySNMPTrap contains events for ntp traffic
	EventCategorySNMPTrap = message.EventCategory1("snmp-trap")
)

// DecodeSNMPTrap will decode NTP packets
func (c *Canary) DecodeSNMPTrap(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(EventSNMPTrap(iph.Src))

	return nil
}

// EventSNMPTrap will return a snmp event struct
func EventSNMPTrap(sourceIP net.IP) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategorySNMPTrap,
		Type:     message.ServiceStarted,
		Details:  map[string]interface{}{},
	}
}

const (
	// EventCategorySNMP contains events for ntp traffic
	EventCategorySNMP = message.EventCategory1("snmp")
)

// DecodeSNMP will decode NTP packets
func (c *Canary) DecodeSNMP(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(EventSNMP(iph.Src))

	return nil
}

// EventSNMP will return a snmp event struct
func EventSNMP(sourceIP net.IP) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategorySNMP,
		Type:     message.ServiceStarted,
		Details:  map[string]interface{}{},
	}
}

const (
	// EventCategoryNTP contains events for ntp traffic
	EventCategoryNTP = message.EventCategory1("ntp")
)

// DecodeNTP will decode NTP packets
func (c *Canary) DecodeNTP(iph *ipv4.Header, udph *udp.Header) error {
	feedback := DummyFeedback{}

	// gopacket
	layer := &layers.NTP{}
	if err := layer.DecodeFromBytes(udph.Payload, feedback); err != nil {
		return err
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(EventNTP(iph.Src, *layer))

	return nil
}

// EventNTP will return a ntp query event struct
func EventNTP(sourceIP net.IP, ntp layers.NTP) message.Event {
	// what to do with other modes?
	modes := map[layers.NTPMode]string{
		layers.NTPMode(0): "reserved",
		layers.NTPMode(1): "Symmetric active",
		layers.NTPMode(2): "Symmetric passive",
		layers.NTPMode(3): "Client",
		layers.NTPMode(4): "Server",
		layers.NTPMode(5): "Broadcast",
		layers.NTPMode(6): "NTP control message",
		layers.NTPMode(7): "private",
	}

	// TODO: message should go into String() / Message, where message.Event will become interface
	mode := fmt.Sprintf("%q", ntp.Mode)
	if m, ok := modes[ntp.Mode]; ok {
		mode = m
	}

	return message.Event{
		Sensor:   "Canary",
		Category: EventCategoryNTP,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"source-ip": sourceIP.String(),
			"message":   fmt.Sprintf("NTP packet received, version=%d, mode=%s", ntp.Version, mode),
			"version":   ntp.Version,
			"mode":      mode,
		},
	}
}

const (
	// EventCategoryDNSQuery contains the category for dns query events
	EventCategoryDNSQuery = message.EventCategory1("dns-query")
	// EventCategoryDNSOther contains the category for dns other events
	EventCategoryDNSOther = message.EventCategory1("dns-other")
)

// DecodeDNS will decode DNS packets
func (c *Canary) DecodeDNS(iph *ipv4.Header, udph *udp.Header) error {
	feedback := DummyFeedback{}

	// gopacket
	layer := &layers.DNS{}
	if err := layer.DecodeFromBytes(udph.Payload, feedback); err != nil {
		return err
	}

	switch layer.OpCode {
	case layers.DNSOpCodeQuery:
		c.events.Send(EventDNSQuery(iph.Src, *layer))

	default:
		c.events.Send(EventDNSOther(iph.Src, *layer))
	}

	// add specific detections, reflection attack detection etc

	return nil
}

// EventDNSQuery will return a dns query event struct
func EventDNSQuery(sourceIP net.IP, dns layers.DNS) message.Event {
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategoryDNSQuery,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"message":   fmt.Sprintf("Querying for: %s", dns.Questions),
			"questions": dns.Questions,
		},
	}
}

// EventDNSOther will return a dns query event struct
func EventDNSOther(sourceIP net.IP, dns layers.DNS) message.Event {
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategoryDNSOther,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"source-ip": sourceIP.String(),
			"message":   fmt.Sprintf("opcode=%s questions=%s", dns.OpCode, dns.Questions),
			"opcode":    dns.OpCode,
			"questions": dns.Questions,
		},
	}
}

// DummyFeedback is a Dummy Feedback struct
type DummyFeedback struct {
}

// SetTruncated will suffice the FeedbackDecoder method
func (f DummyFeedback) SetTruncated() {

}
