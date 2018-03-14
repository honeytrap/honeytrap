// +build linux

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
package canary

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/google/gopacket/layers"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener/canary/ipv4"
	"github.com/honeytrap/honeytrap/listener/canary/udp"
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

var (
	// EventCategorySSDP contains events for ssdp traffic
	EventCategorySSDP = event.Category("ssdp")
)

// DecodeSSDP will decode SSDP packets
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

		event.Protocol("udp"),

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

// DecodeSIP will decode SIP packets
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

		event.Protocol("udp"),

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

// DecodeSNMPTrap will decode SNMP Trap packets
func (c *Canary) DecodeSNMPTrap(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySNMPTrap,

		event.Protocol("udp"),

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

// DecodeSNMP will decode SNMP packets
func (c *Canary) DecodeSNMP(iph *ipv4.Header, udph *udp.Header) error {
	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategorySNMP,

		event.Protocol("udp"),

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

		event.Protocol("udp"),

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

			event.Protocol("udp"),

			event.SourceIP(iph.Src),
			event.DestinationIP(iph.Dst),
			event.SourcePort(udph.Source),
			event.DestinationPort(udph.Destination),

			event.Payload(udph.Payload),

			event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", dns.Questions)),
			event.Custom("dns.questions", dns.Questions),
		))
	default:
		c.events.Send(event.New(
			CanaryOptions,
			EventCategoryDNSOther,

			event.Protocol("udp"),

			event.SourceIP(iph.Src),
			event.DestinationIP(iph.Dst),
			event.SourcePort(udph.Source),
			event.DestinationPort(udph.Destination),

			event.Payload(udph.Payload),

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
