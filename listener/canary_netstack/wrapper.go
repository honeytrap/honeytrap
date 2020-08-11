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

	"github.com/honeytrap/honeytrap/pushers"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/nested"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type endpoint struct {
	nested.Endpoint

	saf *SniffAndFilter
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
func WrapLinkEndpoint(lower stack.LinkEndpoint, opts SniffAndFilterOpts) stack.LinkEndpoint {
	opts.OurMAC = lower.LinkAddress()

	wrapper := &endpoint{
		saf: NewSniffAndFilter(opts),
	}
	wrapper.Endpoint.Init(lower, wrapper)
	return wrapper
}

// DeliverNetworkPacket implements the stack.NetworkDispatcher interface. It is
// called by the link-layer endpoint being wrapped when a packet arrives, and
// logs the packet before forwarding to the actual dispatcher.
func (e *endpoint) DeliverNetworkPacket(remote, local tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	if e.dumpPacket("recv", nil, protocol, pkt) {
		fmt.Println("handleRequest = true")
		e.Endpoint.DeliverNetworkPacket(remote, local, protocol, pkt)
		return
	}
	fmt.Println("handleRequest = false")
}

// dumpPacket logs the packets and returns a boolean if packet should be handled by netstack.
// if true let netstack handle the packet else the host.
func (e *endpoint) dumpPacket(prefix string, gso *stack.GSO, protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) bool {
	return e.saf.logPacket(prefix, protocol, pkt, gso)
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
