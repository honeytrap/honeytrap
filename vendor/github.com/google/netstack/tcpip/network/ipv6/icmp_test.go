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
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/link/channel"
	"github.com/google/netstack/tcpip/link/sniffer"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/ping"
	"github.com/google/netstack/waiter"
)

const (
	linkAddr0 = tcpip.LinkAddress("\x01\x02\x03\x04\x05\x06")
	linkAddr1 = tcpip.LinkAddress("\x0a\x0b\x0c\x0d\x0e\x0f")
)

var (
	lladdr0 = header.LinkLocalAddr(linkAddr0)
	lladdr1 = header.LinkLocalAddr(linkAddr1)
)

type icmpInfo struct {
	typ header.ICMPv6Type
	src tcpip.Address
}

type testContext struct {
	t  *testing.T
	s0 *stack.Stack
	s1 *stack.Stack

	linkEP0 *channel.Endpoint
	linkEP1 *channel.Endpoint

	icmpCh chan icmpInfo
}

type endpointWithResolutionCapability struct {
	stack.LinkEndpoint
}

func (e endpointWithResolutionCapability) Capabilities() stack.LinkEndpointCapabilities {
	return e.LinkEndpoint.Capabilities() | stack.CapabilityResolutionRequired
}

func newTestContext(t *testing.T) *testContext {
	c := &testContext{
		t:      t,
		s0:     stack.New([]string{ProtocolName}, []string{ping.ProtocolName6}, stack.Options{}),
		s1:     stack.New([]string{ProtocolName}, []string{ping.ProtocolName6}, stack.Options{}),
		icmpCh: make(chan icmpInfo, 10),
	}

	const defaultMTU = 65536
	_, linkEP0 := channel.New(256, defaultMTU, linkAddr0)
	c.linkEP0 = linkEP0
	wrappedEP0 := endpointWithResolutionCapability{LinkEndpoint: linkEP0}
	id0 := stack.RegisterLinkEndpoint(wrappedEP0)
	if testing.Verbose() {
		id0 = sniffer.New(id0)
	}
	if err := c.s0.CreateNIC(1, id0); err != nil {
		t.Fatalf("CreateNIC s0: %v", err)
	}
	if err := c.s0.AddAddress(1, ProtocolNumber, lladdr0); err != nil {
		t.Fatalf("AddAddress lladdr0: %v", err)
	}
	if err := c.s0.AddAddress(1, ProtocolNumber, header.SolicitedNodeAddr(lladdr0)); err != nil {
		t.Fatalf("AddAddress sn lladdr0: %v", err)
	}

	_, linkEP1 := channel.New(256, defaultMTU, linkAddr1)
	c.linkEP1 = linkEP1
	wrappedEP1 := endpointWithResolutionCapability{LinkEndpoint: linkEP1}
	id1 := stack.RegisterLinkEndpoint(wrappedEP1)
	if err := c.s1.CreateNIC(1, id1); err != nil {
		t.Fatalf("CreateNIC failed: %v", err)
	}
	if err := c.s1.AddAddress(1, ProtocolNumber, lladdr1); err != nil {
		t.Fatalf("AddAddress lladdr1: %v", err)
	}
	if err := c.s1.AddAddress(1, ProtocolNumber, header.SolicitedNodeAddr(lladdr1)); err != nil {
		t.Fatalf("AddAddress sn lladdr1: %v", err)
	}

	c.s0.SetRouteTable(
		[]tcpip.Route{{
			Destination: lladdr1,
			Mask:        tcpip.AddressMask(strings.Repeat("\xff", 16)),
			NIC:         1,
		}},
	)
	c.s1.SetRouteTable(
		[]tcpip.Route{{
			Destination: lladdr0,
			Mask:        tcpip.AddressMask(strings.Repeat("\xff", 16)),
			NIC:         1,
		}},
	)

	go c.routePackets(linkEP0.C, linkEP1)
	go c.routePackets(linkEP1.C, linkEP0)

	return c
}

func (c *testContext) countPacket(pkt channel.PacketInfo) {
	if pkt.Proto != ProtocolNumber {
		return
	}
	ipv6 := header.IPv6(pkt.Header)
	transProto := tcpip.TransportProtocolNumber(ipv6.NextHeader())
	if transProto != header.ICMPv6ProtocolNumber {
		return
	}
	b := pkt.Header[header.IPv6MinimumSize:]
	icmp := header.ICMPv6(b)
	c.icmpCh <- icmpInfo{
		typ: icmp.Type(),
		src: ipv6.SourceAddress(),
	}
}

func (c *testContext) routePackets(ch <-chan channel.PacketInfo, ep *channel.Endpoint) {
	for pkt := range ch {
		c.countPacket(pkt)
		views := []buffer.View{pkt.Header, pkt.Payload}
		size := len(pkt.Header) + len(pkt.Payload)
		vv := buffer.NewVectorisedView(size, views)
		ep.InjectLinkAddr(pkt.Proto, ep.LinkAddress(), vv)
	}
}

func (c *testContext) cleanup() {
	close(c.linkEP0.C)
	close(c.linkEP1.C)
}

func TestLinkResolution(t *testing.T) {
	c := newTestContext(t)
	defer c.cleanup()
	r, err := c.s0.FindRoute(1, lladdr0, lladdr1, ProtocolNumber)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Release()

	hdr := buffer.NewPrependable(int(r.MaxHeaderLength()) + header.IPv6MinimumSize + header.ICMPv6EchoMinimumSize)
	pkt := header.ICMPv6(hdr.Prepend(header.ICMPv6EchoMinimumSize))
	pkt.SetType(header.ICMPv6EchoRequest)
	pkt.SetChecksum(icmpChecksum(pkt, r.LocalAddress, r.RemoteAddress, buffer.VectorisedView{}))
	payload := tcpip.SlicePayload(hdr.View())

	// We can't send our payload directly over the route because that
	// doesn't provoke NDP discovery.
	var wq waiter.Queue
	ep, err := c.s0.NewEndpoint(header.ICMPv6ProtocolNumber, ProtocolNumber, &wq)
	if err != nil {
		t.Fatal(err)
	}

	// This actually takes about 10 milliseconds, so no need to wait for
	// a multi-minute go test timeout if something is broken.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for {
		if ctx.Err() != nil {
			break
		}
		if _, _, err := ep.Write(payload, tcpip.WriteOptions{To: &tcpip.FullAddress{NIC: 1, Addr: lladdr1}}); err == tcpip.ErrNoLinkAddress {
			// There's something asynchronous going on; yield to let it do its thing.
			runtime.Gosched()
		} else if err == nil {
			break
		} else {
			t.Fatal(err)
		}
	}

	stats := make(map[header.ICMPv6Type]int)
	for {
		select {
		case <-ctx.Done():
			t.Errorf("timeout waiting for ICMP, got: %#+v", stats)
			return
		case icmpInfo := <-c.icmpCh:
			switch icmpInfo.typ {
			case header.ICMPv6NeighborAdvert:
				if got, want := icmpInfo.src, lladdr1; got != want {
					t.Errorf("got ICMPv6NeighborAdvert.sourceAddress = %v, want = %v", got, want)
				}
			}
			stats[icmpInfo.typ]++

			if stats[header.ICMPv6NeighborSolicit] > 0 &&
				stats[header.ICMPv6NeighborAdvert] > 0 &&
				stats[header.ICMPv6EchoRequest] > 0 &&
				stats[header.ICMPv6EchoReply] > 0 {
				return
			}
		}
	}
}
