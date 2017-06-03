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
	"github.com/honeytrap/honeytrap/pushers/event"
)

// contains different variables in use.
var (
	SensorCanary = event.Sensor("Canary")

	// EventCategorySSDP contains events for ssdp traffic
	EventCategoryUDP = event.Category("udp")

	CanaryOptions = event.NewWith(
		SensorCanary,
	)
)

// EventUDP will return a snmp event struct
func EventUDP(sourceIP, destinationIP net.IP, srcport, dstport uint16, payload []byte) event.Event {
	return event.New(
		SensorCanary,
		EventCategoryUDP,

		event.SourceIP(sourceIP),
		event.DestinationIP(destinationIP),

		event.SourcePort(srcport),
		event.DestinationPort(dstport),

		event.Payload(payload),
	)
}

var (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategorySSDP = event.Category("ssdp")
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
func EventSSDP(sourceIP net.IP, method, uri, proto string, headers http.Header) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface

	return event.New(
		CanaryOptions,
		EventCategorySSDP,
		event.ServiceStarted,

		event.Custom("source-ip", sourceIP.String()),
		event.Custom("ssdp.method", method),
		event.Custom("ssdp.uri", uri),
		event.Custom("ssdp.proto", proto),
		event.Custom("ssdp.headers", headers),

		event.Custom("ssdp.host", headers.Get("HOST")),
		event.Custom("ssdp.man", headers.Get("MAN")),
		event.Custom("ssdp.mx", headers.Get("MX")),
		event.Custom("ssdp.st", headers.Get("ST")),
	)
}

var (
	// EventCategorySIP contains events for ntp traffic
	EventCategorySIP = event.Category("sip")
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
func EventSIP(sourceIP net.IP, method, uri, proto string, headers http.Header) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface

	return event.New(
		CanaryOptions,
		EventCategorySNMPTrap,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("sip.method", method),
		event.Custom("sip.uri", uri),
		event.Custom("sip.proto", proto),
		event.Custom("sip.headers", headers),
		event.Custom("sip.from", headers.Get("From")),
		event.Custom("sip.to", headers.Get("To")),
		event.Custom("sip.via", headers.Get("Via")),
		event.Custom("sip.contact", headers.Get("Contact")),
		event.Custom("sip.call-id", headers.Get("Call-ID")),
		event.Custom("sip.user-agent", headers.Get("User-Agent")),
	)
}

var (
	// EventCategorySNMPTrap contains events for ntp traffic
	EventCategorySNMPTrap = event.Category("snmp-trap")
)

// DecodeSNMPTrap will decode NTP packets
func (c *Canary) DecodeSNMPTrap(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(EventSNMPTrap(iph.Src))

	return nil
}

// EventSNMPTrap will return a snmp event struct
func EventSNMPTrap(sourceIP net.IP) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface

	return event.New(
		CanaryOptions,
		EventCategorySNMPTrap,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
	)
}

var (
	// EventCategorySNMP contains events for ntp traffic
	EventCategorySNMP = event.Category("snmp")
)

// DecodeSNMP will decode NTP packets
func (c *Canary) DecodeSNMP(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(EventSNMP(iph.Src))
	return nil
}

// EventSNMP will return a snmp event struct
func EventSNMP(sourceIP net.IP) event.Event {
	// TODO: message should go into String() / Message, where event.Event will become interface

	return event.New(
		CanaryOptions,
		EventCategorySNMP,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
	)
}

var (
	// EventCategoryNTP contains events for ntp traffic
	EventCategoryNTP = event.Category("ntp")
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
func EventNTP(sourceIP net.IP, ntp layers.NTP) event.Event {
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

	// TODO: message should go into String() / Message, where event.Event will become interface
	mode := fmt.Sprintf("%q", ntp.Mode)
	if m, ok := modes[ntp.Mode]; ok {
		mode = m
	}

	return event.New(
		CanaryOptions,
		EventCategoryNTP,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("ntp.message", fmt.Sprintf("NTP packet received, version=%d, mode=%s", ntp.Version, mode)),
		event.Custom("ntp.version", ntp.Version),
		event.Custom("ntp.mode", mode),
	)
}

var (
	// EventCategoryDNSQuery contains the category for dns query events
	EventCategoryDNSQuery = event.Category("dns-query")
	// EventCategoryDNSOther contains the category for dns other events
	EventCategoryDNSOther = event.Category("dns-other")
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
func EventDNSQuery(sourceIP net.IP, dns layers.DNS) event.Event {
	return event.New(
		CanaryOptions,
		EventCategoryDNSQuery,
		event.ServiceStarted,
		event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", dns.Questions)),
		event.Custom("dns.questions", dns.Questions),
	)
}

// EventDNSOther will return a dns query event struct
func EventDNSOther(sourceIP net.IP, dns layers.DNS) event.Event {
	return event.New(
		CanaryOptions,
		EventCategoryDNSOther,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("dns.message", fmt.Sprintf("opcode=%+q questions=%#q", dns.OpCode, dns.Questions)),
		event.Custom("dns.opcode", dns.OpCode),
		event.Custom("dns.questions", dns.Questions),
	)
}

// DummyFeedback is a Dummy Feedback struct
type DummyFeedback struct {
}

// SetTruncated will suffice the FeedbackDecoder method
func (f DummyFeedback) SetTruncated() {

}
