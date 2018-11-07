// Copyright 2018 Google LLC
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

package ipv6

import (
	"encoding/binary"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/stack"
)

// handleControl handles the case when an ICMP packet contains the headers of
// the original packet that caused the ICMP one to be sent. This information is
// used to find out which transport endpoint must be notified about the ICMP
// packet.
func (e *endpoint) handleControl(typ stack.ControlType, extra uint32, vv buffer.VectorisedView) {
	h := header.IPv6(vv.First())

	// We don't use IsValid() here because ICMP only requires that up to
	// 1280 bytes of the original packet be included. So it's likely that it
	// is truncated, which would cause IsValid to return false.
	//
	// Drop packet if it doesn't have the basic IPv6 header or if the
	// original source address doesn't match the endpoint's address.
	if len(h) < header.IPv6MinimumSize || h.SourceAddress() != e.id.LocalAddress {
		return
	}

	// Skip the IP header, then handle the fragmentation header if there
	// is one.
	vv.TrimFront(header.IPv6MinimumSize)
	p := h.TransportProtocol()
	if p == header.IPv6FragmentHeader {
		f := header.IPv6Fragment(vv.First())
		if !f.IsValid() || f.FragmentOffset() != 0 {
			// We can't handle fragments that aren't at offset 0
			// because they don't have the transport headers.
			return
		}

		// Skip fragmentation header and find out the actual protocol
		// number.
		vv.TrimFront(header.IPv6FragmentHeaderSize)
		p = f.TransportProtocol()
	}

	// Deliver the control packet to the transport endpoint.
	e.dispatcher.DeliverTransportControlPacket(e.id.LocalAddress, h.DestinationAddress(), ProtocolNumber, p, typ, extra, vv)
}

func (e *endpoint) handleICMP(r *stack.Route, vv buffer.VectorisedView) {
	v := vv.First()
	if len(v) < header.ICMPv6MinimumSize {
		return
	}
	h := header.ICMPv6(v)

	switch h.Type() {
	case header.ICMPv6PacketTooBig:
		if len(v) < header.ICMPv6PacketTooBigMinimumSize {
			return
		}
		vv.TrimFront(header.ICMPv6PacketTooBigMinimumSize)
		mtu := binary.BigEndian.Uint32(v[header.ICMPv6MinimumSize:])
		e.handleControl(stack.ControlPacketTooBig, calculateMTU(mtu), vv)

	case header.ICMPv6DstUnreachable:
		if len(v) < header.ICMPv6DstUnreachableMinimumSize {
			return
		}
		vv.TrimFront(header.ICMPv6DstUnreachableMinimumSize)
		switch h.Code() {
		case header.ICMPv6PortUnreachable:
			e.handleControl(stack.ControlPortUnreachable, 0, vv)
		}

	case header.ICMPv6NeighborSolicit:
		if len(v) < header.ICMPv6NeighborSolicitMinimumSize {
			return
		}
		targetAddr := tcpip.Address(v[8 : 8+16])
		if e.linkAddrCache.CheckLocalAddress(e.nicid, ProtocolNumber, targetAddr) == 0 {
			// We don't have a useful answer; the best we can do is ignore the request.
			return
		}
		hdr := buffer.NewPrependable(int(r.MaxHeaderLength()) + header.IPv6MinimumSize + header.ICMPv6NeighborAdvertSize)
		pkt := header.ICMPv6(hdr.Prepend(header.ICMPv6NeighborAdvertSize))
		pkt.SetType(header.ICMPv6NeighborAdvert)
		pkt[icmpV6FlagOffset] = ndpSolicitedFlag | ndpOverrideFlag
		copy(pkt[icmpV6OptOffset-len(targetAddr):], targetAddr)
		pkt[icmpV6OptOffset] = ndpOptDstLinkAddr
		pkt[icmpV6LengthOffset] = 1
		copy(pkt[icmpV6LengthOffset+1:], r.LocalLinkAddress[:])

		// ICMPv6 Neighbor Solicit messages are always sent to
		// specially crafted IPv6 multicast addresses. As a result, the
		// route we end up with here has as its LocalAddress such a
		// multicast address. It would be nonsense to claim that our
		// source address is a multicast address, so we manually set
		// the source address to the target address requested in the
		// solicit message. Since that requires mutating the route, we
		// must first clone it.
		r := r.Clone()
		defer r.Release()
		r.LocalAddress = targetAddr
		pkt.SetChecksum(icmpChecksum(pkt, r.LocalAddress, r.RemoteAddress, buffer.VectorisedView{}))
		r.WritePacket(hdr, buffer.VectorisedView{}, header.ICMPv6ProtocolNumber, r.DefaultTTL())

		e.linkAddrCache.AddLinkAddress(e.nicid, r.RemoteAddress, r.RemoteLinkAddress)

	case header.ICMPv6NeighborAdvert:
		if len(v) < header.ICMPv6NeighborAdvertSize {
			return
		}
		targetAddr := tcpip.Address(v[8 : 8+16])
		e.linkAddrCache.AddLinkAddress(e.nicid, targetAddr, r.RemoteLinkAddress)
		if targetAddr != r.RemoteAddress {
			e.linkAddrCache.AddLinkAddress(e.nicid, r.RemoteAddress, r.RemoteLinkAddress)
		}

	case header.ICMPv6EchoRequest:
		if len(v) < header.ICMPv6EchoMinimumSize {
			return
		}
		vv.TrimFront(header.ICMPv6EchoMinimumSize)
		hdr := buffer.NewPrependable(int(r.MaxHeaderLength()) + header.IPv6MinimumSize + header.ICMPv6EchoMinimumSize)
		pkt := header.ICMPv6(hdr.Prepend(header.ICMPv6EchoMinimumSize))
		copy(pkt, h)
		pkt.SetType(header.ICMPv6EchoReply)
		pkt.SetChecksum(icmpChecksum(pkt, r.LocalAddress, r.RemoteAddress, vv))
		r.WritePacket(hdr, vv, header.ICMPv6ProtocolNumber, r.DefaultTTL())

	case header.ICMPv6EchoReply:
		if len(v) < header.ICMPv6EchoMinimumSize {
			return
		}
		e.dispatcher.DeliverTransportPacket(r, header.ICMPv6ProtocolNumber, vv)

	}
}

const (
	ndpSolicitedFlag = 1 << 6
	ndpOverrideFlag  = 1 << 5

	ndpOptSrcLinkAddr = 1
	ndpOptDstLinkAddr = 2

	icmpV6FlagOffset   = 4
	icmpV6OptOffset    = 24
	icmpV6LengthOffset = 25
)

var broadcastMAC = tcpip.LinkAddress([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

var _ stack.LinkAddressResolver = (*protocol)(nil)

// LinkAddressProtocol implements stack.LinkAddressResolver.
func (*protocol) LinkAddressProtocol() tcpip.NetworkProtocolNumber {
	return header.IPv6ProtocolNumber
}

// LinkAddressRequest implements stack.LinkAddressResolver.
func (*protocol) LinkAddressRequest(addr, localAddr tcpip.Address, linkEP stack.LinkEndpoint) *tcpip.Error {
	snaddr := header.SolicitedNodeAddr(addr)
	r := &stack.Route{
		LocalAddress:      localAddr,
		RemoteAddress:     snaddr,
		RemoteLinkAddress: broadcastMAC,
	}
	hdr := buffer.NewPrependable(int(linkEP.MaxHeaderLength()) + header.IPv6MinimumSize + header.ICMPv6NeighborAdvertSize)
	pkt := header.ICMPv6(hdr.Prepend(header.ICMPv6NeighborAdvertSize))
	pkt.SetType(header.ICMPv6NeighborSolicit)
	copy(pkt[icmpV6OptOffset-len(addr):], addr)
	pkt[icmpV6OptOffset] = ndpOptSrcLinkAddr
	pkt[icmpV6LengthOffset] = 1
	copy(pkt[icmpV6LengthOffset+1:], linkEP.LinkAddress())
	pkt.SetChecksum(icmpChecksum(pkt, r.LocalAddress, r.RemoteAddress, buffer.VectorisedView{}))

	length := uint16(hdr.UsedLength())
	ip := header.IPv6(hdr.Prepend(header.IPv6MinimumSize))
	ip.Encode(&header.IPv6Fields{
		PayloadLength: length,
		NextHeader:    uint8(header.ICMPv6ProtocolNumber),
		HopLimit:      defaultIPv6HopLimit,
		SrcAddr:       r.LocalAddress,
		DstAddr:       r.RemoteAddress,
	})

	return linkEP.WritePacket(r, hdr, buffer.VectorisedView{}, ProtocolNumber)
}

// ResolveStaticAddress implements stack.LinkAddressResolver.
func (*protocol) ResolveStaticAddress(addr tcpip.Address) (tcpip.LinkAddress, bool) {
	return "", false
}

func icmpChecksum(h header.ICMPv6, src, dst tcpip.Address, vv buffer.VectorisedView) uint16 {
	// Calculate the IPv6 pseudo-header upper-layer checksum.
	xsum := header.Checksum([]byte(src), 0)
	xsum = header.Checksum([]byte(dst), xsum)
	var upperLayerLength [4]byte
	binary.BigEndian.PutUint32(upperLayerLength[:], uint32(len(h)+vv.Size()))
	xsum = header.Checksum(upperLayerLength[:], xsum)
	xsum = header.Checksum([]byte{0, 0, 0, uint8(header.ICMPv6ProtocolNumber)}, xsum)
	for _, v := range vv.Views() {
		xsum = header.Checksum(v, xsum)
	}

	// h[2:4] is the checksum itself, set it aside to avoid checksumming the checksum.
	h2, h3 := h[2], h[3]
	h[2], h[3] = 0, 0
	xsum = ^header.Checksum(h, xsum)
	h[2], h[3] = h2, h3

	return xsum
}
