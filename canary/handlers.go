package canary

import (
	"fmt"
	"net"

	"github.com/google/gopacket/layers"
	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/udp"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	// EventCategorySNMPTrap contains events for ntp traffic
	EventCategorySNMPTrap = message.EventCategory("snmp-trap")
)

// DecodeSNMPTrap will decode NTP packets
func (c *Canary) DecodeSNMPTrap(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Deliver(EventSNMPTrap(iph.Src))

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
	EventCategorySNMP = message.EventCategory("snmp")
)

// DecodeSNMP will decode NTP packets
func (c *Canary) DecodeSNMP(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Deliver(EventSNMP(iph.Src))

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
	c.events.Deliver(EventNTP(iph.Src, *layer))

	return nil
}

// EventNTP will return a ntp query event struct
func EventNTP(sourceIP net.IP, ntp layers.NTP) message.Event {
	// TODO: message should go into String() / Message, where message.Event will become interface
	return message.Event{
		Sensor:   "Canary",
		Category: EventCategoryNTP,
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("NTP packet received, mode=%q\n", ntp.Mode),
		},
	}
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

	switch layer.OpCode {
	case layers.DNSOpCodeQuery:
		c.events.Deliver(EventDNSQuery(iph.Src, *layer))

	default:
		c.events.Deliver(EventDNSOther(iph.Src, *layer))
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
			"message":   fmt.Sprintf("Querying for: %s\n", dns.Questions),
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
			"message":   fmt.Sprintf("opcode=%s questions=%s\n", dns.OpCode, dns.Questions),
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
