// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// wrapper.go: sniffer provides the implementation of data-link layer endpoints that
// wrap another endpoint and logs inbound packets.
//
// Sniffer endpoints can be used in the networking stack by calling
// WrapLinkEndpoint(eID, pushers.Channel, chan KnockGrouper) to
// create a new endpoint, where eID is the ID of the endpoint being wrapped,
// and then passing it as an argument to Stack.CreateNIC().

package nscanary

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/nested"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Flags to use for LogProtos.
const (
	ProtoIPv4 uint32 = 1 << iota
	ProtoIPv6        = 1 << iota
	ProtoARP         = 1 << iota
	ProtoTCP         = 1 << iota
	ProtoUDP         = 1 << iota
	ProtoICMP        = 1 << iota
	ProtoAll         = ProtoIPv4 | ProtoIPv6 | ProtoTCP | ProtoUDP | ProtoICMP | ProtoARP
)

// LogProtos is a flag-set used to enable(1) events for a protocol.
// LogProtos = 0; disables event logging.
//
// LogProtos must be accessed atomically.
var LogProtos uint32 = ProtoAll

type endpoint struct {
	nested.Endpoint

	sniffer *sniffer
}

type sniffer struct {
	events    pushers.Channel
	knockChan chan KnockGrouper
	ourMAC    tcpip.LinkAddress //our network interface hardware address.
}

var _ stack.GSOEndpoint = (*endpoint)(nil)
var _ stack.LinkEndpoint = (*endpoint)(nil)
var _ stack.NetworkDispatcher = (*endpoint)(nil)

// WrapLinkEndpoint creates a new sniffer link-layer endpoint. It wraps around another
// endpoint and logs packets as they traverse the endpoint.
// Created events are send by 'e', an configured pushers.Chanel.
// Knock detection is send in 'knocks' listened on by RunKnockDetector.
func WrapLinkEndpoint(lower stack.LinkEndpoint, e pushers.Channel, knocks chan KnockGrouper) stack.LinkEndpoint {
	wrapper := &endpoint{
		sniffer: &sniffer{
			events:    e,
			knockChan: knocks,
			ourMAC:    lower.LinkAddress(),
		},
	}
	wrapper.Endpoint.Init(lower, wrapper)
	return wrapper
}

// DeliverNetworkPacket implements the stack.NetworkDispatcher interface. It is
// called by the link-layer endpoint being wrapped when a packet arrives, and
// logs the packet before forwarding to the actual dispatcher.
func (e *endpoint) DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	e.dumpPacket("recv", nil, protocol, pkt)
	e.Endpoint.DeliverNetworkPacket(remote, local, protocol, pkt)
}

func (e *endpoint) dumpPacket(prefix string, gso *stack.GSO, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	if atomic.LoadUint32(&LogProtos) > 0 {
		e.sniffer.logPacket(prefix, protocol, pkt, gso)
	}
}

// WritePacket implements the stack.LinkEndpoint interface. It is called by
// higher-level protocols to write packets; it just
// forwards the request to the lower endpoint.
func (e *endpoint) WritePacket(r *stack.Route, gso *stack.GSO, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) *tcpip.Error {
	//don't send an event on outgoing packets.
	//e.dumpPacket("send", gso, protocol, pkt)
	return e.Endpoint.WritePacket(r, gso, protocol, pkt)
}

// WritePackets implements the stack.LinkEndpoint interface. It is called by
// higher-level protocols to write packets; it just
// forwards the request to the lower endpoint.
func (e *endpoint) WritePackets(r *stack.Route, gso *stack.GSO, pkts stack.PacketBufferList, protocol tcpip.NetworkProtocolNumber) (int, *tcpip.Error) {
	// don't send an event on outgoing packets.
	// for pkt := pkts.Front(); pkt != nil; pkt = pkt.Next() {
	// 	e.dumpPacket("send", gso, protocol, pkt)
	// }
	return e.Endpoint.WritePackets(r, gso, pkts, protocol)
}

// WriteRawPacket implements stack.LinkEndpoint.WriteRawPacket.
func (e *endpoint) WriteRawPacket(vv buffer.VectorisedView) *tcpip.Error {
	// don't send an event on outgoing packets.
	// e.dumpPacket("send", nil, 0, &stack.PacketBuffer{
	// 	Data: vv,
	// })
	return e.Endpoint.WriteRawPacket(vv)
}

func (s *sniffer) logPacket(prefix string, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer, gso *stack.GSO) {
	// Figure out the network layer info.
	var (
		transProto     uint8
		fragmentOffset uint16
		moreFragments  bool
		srcMAC         tcpip.LinkAddress
		destMAC        tcpip.LinkAddress
	)

	eoptions := make([]event.Option, 0, 16)
	eoptions = append(eoptions, CanaryOptions)

	// set the hardware addresses.
	// skip logging if we are the source.
	if len(pkt.LinkHeader) > 0 {
		eth := header.Ethernet(pkt.LinkHeader)
		srcMAC = eth.SourceAddress()
		if srcMAC == s.ourMAC {
			return
		}
		destMAC = eth.DestinationAddress()
		eoptions = append(eoptions,
			event.DestinationHardwareAddr(net.HardwareAddr(destMAC)),
			event.SourceHardwareAddr(net.HardwareAddr(srcMAC)),
			event.Custom("network-protocol-number", protocol),
		)
	}

	src := tcpip.Address("unknown")
	dst := tcpip.Address("unknown")
	id := 0
	size := uint16(0)

	// Create a clone of pkt, including any headers if present. Avoid allocating
	// backing memory for the clone.
	views := [8]buffer.View{}
	vv := buffer.NewVectorisedView(0, views[:0])
	vv.AppendView(pkt.Header.View())
	vv.Append(pkt.Data)

	switch protocol {
	case header.IPv4ProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&ProtoIPv4 == 0 {
			return
		}
		hdr := header.IPv4(vv.ToView())
		if !hdr.IsValid(len(hdr)) {
			//TODO (jerry): log invalid header??
			return
		}
		fragmentOffset = hdr.FragmentOffset()
		moreFragments = hdr.Flags()&header.IPv4FlagMoreFragments == header.IPv4FlagMoreFragments
		src = hdr.SourceAddress()
		dst = hdr.DestinationAddress()
		transProto = hdr.Protocol()
		size = hdr.PayloadLength()
		vv.TrimFront(int(hdr.HeaderLength()))
		id = int(hdr.ID())
		eoptions = append(eoptions,
			event.Custom("ip-version", "4"),
			event.Custom("source-ip", hdr.SourceAddress().String()),
			event.Custom("destination-ip", hdr.DestinationAddress().String()),
			event.Payload(hdr.Payload()),
		)

	case header.IPv6ProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&ProtoIPv6 == 0 {
			return
		}
		hdr := header.IPv6(vv.ToView())
		if !hdr.IsValid(len(hdr)) {
			//TODO (jerry): log invalid header??
			return
		}
		src = hdr.SourceAddress()
		dst = hdr.DestinationAddress()
		transProto = hdr.NextHeader()
		size = hdr.PayloadLength()
		vv.TrimFront(vv.Size() - int(size))
		eoptions = append(eoptions,
			event.Custom("ip-version", "6"),
			event.Custom("source-ip", hdr.SourceAddress().String()),
			event.Custom("destination-ip", hdr.DestinationAddress().String()),
			event.Payload(hdr.Payload()),
		)

	case header.ARPProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&ProtoARP == 0 {
			return
		}
		hdr := header.ARP(vv.ToView())
		if !hdr.IsValid() {
			return
		}
		line := fmt.Sprintf(
			"%s arp %s (%s) -> %s (%s) valid:%t",
			prefix,
			tcpip.Address(hdr.ProtocolAddressSender()), tcpip.LinkAddress(hdr.HardwareAddressSender()),
			tcpip.Address(hdr.ProtocolAddressTarget()), tcpip.LinkAddress(hdr.HardwareAddressTarget()),
			hdr.IsValid(),
		)

		s.events.Send(event.New(
			CanaryOptions,
			event.Category("ARP"),
			event.DestinationHardwareAddr(net.HardwareAddr(hdr.HardwareAddressTarget())),
			event.SourceHardwareAddr(net.HardwareAddr(hdr.HardwareAddressSender())),
			event.SourceIP(hdr.ProtocolAddressSender()),
			event.DestinationIP(hdr.ProtocolAddressTarget()),
			event.Custom("arp-opcode", hdr.Op()),
			event.Custom("arp-isvalid", hdr.IsValid()),
			//event.Custom("arp-hardware-type", hdr.HardwareType),
			//event.Custom("arp-hardware-size", hdr.HardwareSize),
			//event.Custom("arp-protocol-type", hdr.ProtocolType),
			//event.Custom("arp-protocol-size", hdr.ProtocolSize),
			event.Message(line),
		))
		return
	default:
		eoptions = append(eoptions,
			event.Message("unknown network protocol"),
		)
	}

	// Figure out the transport layer info.
	eoptions = append(eoptions,
		event.Custom("transport-protocol-number", transProto),
	)
	transName := "unknown"
	srcPort := uint16(0)
	dstPort := uint16(0)
	details := ""
	switch tcpip.TransportProtocolNumber(transProto) {
	case header.ICMPv4ProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&(ProtoICMP|ProtoIPv4) != ProtoICMP|ProtoIPv4 {
			return
		}
		transName = "icmp"
		if vv.Size() < header.ICMPv4MinimumSize {
			break
		}
		hdr := header.ICMPv4(vv.ToView())
		icmpType := "unknown"
		if fragmentOffset == 0 {
			switch hdr.Type() {
			case header.ICMPv4EchoReply:
				icmpType = "echo reply"
			case header.ICMPv4DstUnreachable:
				icmpType = "destination unreachable"
			case header.ICMPv4SrcQuench:
				icmpType = "source quench"
			case header.ICMPv4Redirect:
				icmpType = "redirect"
			case header.ICMPv4Echo:
				icmpType = "echo"
			case header.ICMPv4TimeExceeded:
				icmpType = "time exceeded"
			case header.ICMPv4ParamProblem:
				icmpType = "param problem"
			case header.ICMPv4Timestamp:
				icmpType = "timestamp"
			case header.ICMPv4TimestampReply:
				icmpType = "timestamp reply"
			case header.ICMPv4InfoRequest:
				icmpType = "info request"
			case header.ICMPv4InfoReply:
				icmpType = "info reply"
			}
		}
		line := fmt.Sprintf("%s %s %s -> %s %s len:%d id:%04x code:%d", prefix, transName, src, dst, icmpType, size, id, hdr.Code())
		//TODO (jerry): Add communty-id
		eoptions = append(eoptions,
			event.Category("icmp"),
			event.Protocol("ICMPv4"),
			event.Custom("icmp-type", icmpType),
			event.Custom("icmp-code", hdr.Code()),
			event.Message(line),
			event.Payload(hdr.Payload()),
		)

		s.knockChan <- KnockICMP{
			IPVersion:               4,
			SourceHardwareAddr:      net.HardwareAddr(srcMAC),
			DestinationHardwareAddr: net.HardwareAddr(destMAC),
			SourceIP:                src,
			DestinationIP:           dst,
		}

	case header.ICMPv6ProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&(ProtoICMP|ProtoIPv6) != ProtoICMP|ProtoIPv6 {
			return
		}
		transName = "icmp"
		if vv.Size() < header.ICMPv6MinimumSize {
			break
		}
		hdr := header.ICMPv6(vv.ToView())
		icmpType := "unknown"
		switch hdr.Type() {
		case header.ICMPv6DstUnreachable:
			icmpType = "destination unreachable"
		case header.ICMPv6PacketTooBig:
			icmpType = "packet too big"
		case header.ICMPv6TimeExceeded:
			icmpType = "time exceeded"
		case header.ICMPv6ParamProblem:
			icmpType = "param problem"
		case header.ICMPv6EchoRequest:
			icmpType = "echo request"
		case header.ICMPv6EchoReply:
			icmpType = "echo reply"
		case header.ICMPv6RouterSolicit:
			icmpType = "router solicit"
		case header.ICMPv6RouterAdvert:
			icmpType = "router advert"
		case header.ICMPv6NeighborSolicit:
			icmpType = "neighbor solicit"
		case header.ICMPv6NeighborAdvert:
			icmpType = "neighbor advert"
		case header.ICMPv6RedirectMsg:
			icmpType = "redirect message"
		}
		line := fmt.Sprintf("%s %s %s -> %s %s len:%d id:%04x code:%d", prefix, transName, src, dst, icmpType, size, id, hdr.Code())
		//TODO (jerry): Add communty-id
		eoptions = append(eoptions,
			event.Category("icmp"),
			event.Protocol("ICMPv6"),
			event.Custom("icmp-type", icmpType),
			event.Custom("icmp-code", hdr.Code()),
			event.Message(line),
			event.Payload(hdr.Payload()),
		)

		s.knockChan <- KnockICMP{
			IPVersion:               6,
			SourceHardwareAddr:      net.HardwareAddr(srcMAC),
			DestinationHardwareAddr: net.HardwareAddr(destMAC),
			SourceIP:                src,
			DestinationIP:           dst,
		}

	case header.UDPProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&ProtoUDP == 0 {
			return
		}
		transName = "udp"
		if vv.Size() < header.UDPMinimumSize {
			break
		}
		hdr := header.UDP(vv.ToView())
		if fragmentOffset == 0 {
			srcPort = hdr.SourcePort()
			dstPort = hdr.DestinationPort()
			details = fmt.Sprintf("xsum: 0x%x", hdr.Checksum())
			size -= header.UDPMinimumSize
		}
		//TODO (jerry): Add communty-id
		eoptions = append(eoptions,
			event.Category("udp"),
			event.Protocol("UDP"),
			event.SourcePort(hdr.SourcePort()),
			event.DestinationPort(hdr.DestinationPort()),
			event.Payload(hdr.Payload()),
		)

		s.knockChan <- KnockUDPPort{
			SourceHardwareAddr:      net.HardwareAddr(srcMAC),
			DestinationHardwareAddr: net.HardwareAddr(destMAC),
			SourceIP:                src,
			DestinationIP:           dst,
			DestinationPort:         dstPort,
		}

	case header.TCPProtocolNumber:
		if atomic.LoadUint32(&LogProtos)&ProtoTCP == 0 {
			return
		}
		transName = "tcp"
		if vv.Size() < header.TCPMinimumSize {
			break
		}
		hdr := header.TCP(vv.ToView())
		if fragmentOffset == 0 {
			offset := int(hdr.DataOffset())
			if offset < header.TCPMinimumSize {
				details += fmt.Sprintf("invalid packet: tcp data offset too small %d", offset)
				break
			}
			if offset > vv.Size() && !moreFragments {
				details += fmt.Sprintf("invalid packet: tcp data offset %d larger than packet buffer length %d", offset, vv.Size())
				break
			}

			srcPort = hdr.SourcePort()
			dstPort = hdr.DestinationPort()
			size -= uint16(offset)

			// Initialize the TCP flags.
			flags := hdr.Flags()
			flagsStr := []byte("FSRPAU")
			for i := range flagsStr {
				if flags&(1<<uint(i)) == 0 {
					flagsStr[i] = ' '
				}
			}
			details = fmt.Sprintf("flags:0x%02x (%s) seqnum: %d ack: %d win: %d xsum:0x%x", flags, string(flagsStr), hdr.SequenceNumber(), hdr.AckNumber(), hdr.WindowSize(), hdr.Checksum())
			if flags&header.TCPFlagSyn != 0 {
				details += fmt.Sprintf(" options: %+v", header.ParseSynOptions(hdr.Options(), flags&header.TCPFlagAck != 0))
			} else {
				details += fmt.Sprintf(" options: %+v", hdr.ParsedOptions())
			}
		}

		//TODO (jerry): Add communty-id
		eoptions = append(eoptions,
			event.Category("tcp"),
			event.Protocol("TCP"),
			event.SourcePort(hdr.SourcePort()),
			event.DestinationPort(hdr.DestinationPort()),
			event.Payload(hdr.Payload()),
		)

		s.knockChan <- KnockTCPPort{
			SourceHardwareAddr:      net.HardwareAddr(srcMAC),
			DestinationHardwareAddr: net.HardwareAddr(destMAC),
			SourceIP:                src,
			DestinationIP:           dst,
			DestinationPort:         dstPort,
		}

	default:
		s.events.Send(event.New(
			CanaryOptions,
			EventCategoryUnknown,
			event.Message("unknown transport protocol"),
			event.Payload(vv.ToView()),
		))
	}

	line := fmt.Sprintf("%s %s %s:%d -> %s:%d len:%d id:%04x %s", prefix, transName, src, srcPort, dst, dstPort, size, id, details)

	eoptions = append(eoptions,
		event.Message(line),
	)

	s.events.Send(event.New(
		eoptions...,
	))
}

// ExcludeLogProtos exclude protos from event logging.
// recognized options for protos: ["ip4", "ip6", "arp", "udp", "tcp", "icmp"]
//
// This sets the global 'LogProtos'
func ExcludeLogProtos(protos []string) {
	flags := ProtoAll
	for _, proto := range protos {
		switch proto {
		case "ip4":
			flags &^= ProtoIPv4
		case "ip6":
			flags &^= ProtoIPv6
		case "arp":
			flags &^= ProtoARP
		case "udp":
			flags &^= ProtoUDP
		case "tcp":
			flags &^= ProtoTCP
		case "icmp":
			flags &^= ProtoICMP
		}
	}
	LogProtos = flags
}
