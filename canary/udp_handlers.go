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
	SensorCanary = event.Sensor("canary")

	// EventCategorySSDP contains events for ssdp traffic
	EventCategoryUDP = event.Category("udp")

	CanaryOptions = event.NewWith(
		SensorCanary,
	)
)

// EventUDP will return a snmp event struct
func EventUDP(sourceIP, destinationIP net.IP, srcport, dstport uint16, payload []byte) *event.Event {
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
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySSDP,
		event.ServiceStarted,

		event.SourceIP(iph.Src),
		event.DestinationIP(iph.Dst),
		event.SourcePort(udph.Source),
		event.DestinationPort(udph.Destination),

		event.Custom("ssdp.method", request.Method),
		event.Custom("ssdp.uri", request.RequestURI),
		event.Custom("ssdp.proto", request.Proto),
		event.Custom("ssdp.headers", request.Header),

		event.Custom("ssdp.host", request.Header.Get("HOST")),
		event.Custom("ssdp.man", request.Header.Get("MAN")),
		event.Custom("ssdp.mx", request.Header.Get("MX")),
		event.Custom("ssdp.st", request.Header.Get("ST")),
	))

	return nil
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
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySIP,
		event.ServiceStarted,

		event.SourceIP(iph.Src),
		event.DestinationIP(iph.Dst),
		event.SourcePort(udph.Source),
		event.DestinationPort(udph.Destination),

		event.Custom("sip.method", request.Method),
		event.Custom("sip.uri", request.RequestURI),
		event.Custom("sip.proto", request.Proto),
		event.Custom("sip.headers", request.Header),
		event.Custom("sip.from", request.Header.Get("From")),
		event.Custom("sip.to", request.Header.Get("To")),
		event.Custom("sip.via", request.Header.Get("Via")),
		event.Custom("sip.contact", request.Header.Get("Contact")),
		event.Custom("sip.call-id", request.Header.Get("Call-ID")),
		event.Custom("sip.user-agent", request.Header.Get("User-Agent")),
	))

	return nil
}

var (
	// EventCategorySNMPTrap contains events for ntp traffic
	EventCategorySNMPTrap = event.Category("snmp-trap")
)

// DecodeSNMPTrap will decode NTP packets
func (c *Canary) DecodeSNMPTrap(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySNMPTrap,
		event.ServiceStarted,

		event.SourceIP(iph.Src),
		event.DestinationIP(iph.Dst),
		event.SourcePort(udph.Source),
		event.DestinationPort(udph.Destination),
	))

	return nil
}

var (
	// EventCategorySNMP contains events for ntp traffic
	EventCategorySNMP = event.Category("snmp")
)

// DecodeSNMP will decode NTP packets
func (c *Canary) DecodeSNMP(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySNMP,
		event.ServiceStarted,

		event.SourceIP(iph.Src),
		event.DestinationIP(iph.Dst),
		event.SourcePort(udph.Source),
		event.DestinationPort(udph.Destination),
	))

	return nil
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
	ntp := *layer

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

	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryNTP,
		event.ServiceStarted,

		event.SourceIP(iph.Src),
		event.DestinationIP(iph.Dst),
		event.SourcePort(udph.Source),
		event.DestinationPort(udph.Destination),

		event.Custom("ntp.message", fmt.Sprintf("NTP packet received, version=%d, mode=%s", ntp.Version, mode)),
		event.Custom("ntp.version", ntp.Version),
		event.Custom("ntp.mode", mode),
	))

	return nil
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

	// DNSTypeANY -> Amplification attack (https://www.us-cert.gov/ncas/alerts/TA13-088A)

	dns := *layer

	switch layer.OpCode {
	case layers.DNSOpCodeQuery:
		c.events.Send(event.New(
			CanaryOptions,
			EventCategoryDNSQuery,
			event.ServiceStarted,

			event.SourceIP(iph.Src),
			event.DestinationIP(iph.Dst),
			event.SourcePort(udph.Source),
			event.DestinationPort(udph.Destination),

			event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", dns.Questions)),
			event.Custom("dns.questions", dns.Questions),
		))
	default:
		c.events.Send(event.New(
			CanaryOptions,
			EventCategoryDNSOther,
			event.ServiceStarted,

			event.SourceIP(iph.Src),
			event.DestinationIP(iph.Dst),
			event.SourcePort(udph.Source),
			event.DestinationPort(udph.Destination),

			event.Message("opcode=%+q questions=%#q", dns.OpCode, dns.Questions),
			event.Custom("dns.opcode", dns.OpCode),
			event.Custom("dns.questions", dns.Questions),
		))
	}

	// add specific detections, reflection attack detection etc

	return nil
}

// DummyFeedback is a Dummy Feedback struct
type DummyFeedback struct {
}

// SetTruncated will suffice the FeedbackDecoder method
func (f DummyFeedback) SetTruncated() {

}
