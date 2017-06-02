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
	"github.com/honeytrap/honeytrap/pushers/message"
)

var (
	SensorCanary = event.Sensor("Canary")

	// EventCategorySSDP contains events for ssdp traffic
	EventCategoryUDP = event.Category("udp")
)

// EventSSDP will return a snmp event struct
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

const (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategorySSDP = message.EventCategory("ssdp")
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
	return message.NewEvent("Canary", EventCategorySSDP, message.ServiceStarted, map[string]interface{}{
		"source-ip": sourceIP.String(),

		"ssdp.method":  method,
		"ssdp.uri":     uri,
		"ssdp.proto":   proto,
		"ssdp.headers": headers,

		"ssdp.host": headers.Get("HOST"),
		"ssdp.man":  headers.Get("MAN"),
		"ssdp.mx":   headers.Get("MX"),
		"ssdp.st":   headers.Get("ST"),
	})
}

const (
	// EventCategorySIP contains events for ntp traffic
	EventCategorySIP = message.EventCategory("sip")
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
	return message.NewEvent("Canary", EventCategorySIP, message.ServiceStarted, map[string]interface{}{
		"source-ip":      sourceIP.String(),
		"sip.method":     method,
		"sip.uri":        uri,
		"sip.proto":      proto,
		"sip.headers":    headers,
		"sip.from":       headers.Get("From"),
		"sip.to":         headers.Get("To"),
		"sip.via":        headers.Get("Via"),
		"sip.contact":    headers.Get("Contact"),
		"sip.call-id":    headers.Get("Call-ID"),
		"sip.user-agent": headers.Get("User-Agent"),
	})
}

const (
	// EventCategorySNMPTrap contains events for ntp traffic
	EventCategorySNMPTrap = message.EventCategory("snmp-trap")
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
	return message.NewEvent("Canary", EventCategorySNMPTrap, message.ServiceStarted, map[string]interface{}{
		"source-ip": sourceIP.String(),
	})
}

const (
	// EventCategorySNMP contains events for ntp traffic
	EventCategorySNMP = message.EventCategory("snmp")
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
	return message.NewEvent("Canary", EventCategorySNMP, message.ServiceStarted, map[string]interface{}{
		"source-ip": sourceIP.String(),
	})
}

const (
	// EventCategoryNTP contains events for ntp traffic
	EventCategoryNTP = message.EventCategory("ntp")
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

	return message.NewEvent("Canary", EventCategoryNTP, message.ServiceStarted, map[string]interface{}{
		"source-ip":   sourceIP.String(),
		"ntp.message": fmt.Sprintf("NTP packet received, version=%d, mode=%s", ntp.Version, mode),
		"ntp.version": ntp.Version,
		"ntp.mode":    mode,
	})
}

const (
	// EventCategoryDNSQuery contains the category for dns query events
	EventCategoryDNSQuery = message.EventCategory("dns-query")
	// EventCategoryDNSOther contains the category for dns other events
	EventCategoryDNSOther = message.EventCategory("dns-other")
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
	return message.NewEvent("Canary", EventCategoryDNSQuery, message.ServiceStarted, map[string]interface{}{
		"message":       fmt.Sprintf("Querying for: %s", dns.Questions),
		"dns.questions": dns.Questions,
	})
}

// EventDNSOther will return a dns query event struct
func EventDNSOther(sourceIP net.IP, dns layers.DNS) message.Event {
	return message.NewEvent("Canary", EventCategoryDNSOther, message.ServiceStarted, map[string]interface{}{
		"source-ip":     sourceIP.String(),
		"dns.message":   fmt.Sprintf("opcode=%s questions=%s", dns.OpCode, dns.Questions),
		"dns.opcode":    dns.OpCode,
		"dns.questions": dns.Questions,
	})
}

// DummyFeedback is a Dummy Feedback struct
type DummyFeedback struct {
}

// SetTruncated will suffice the FeedbackDecoder method
func (f DummyFeedback) SetTruncated() {

}
