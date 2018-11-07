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
package netstack

import (
	"sync"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/udp"
	"github.com/google/netstack/waiter"
)

// Forwarder is a connection request forwarder, which allows clients to decide
// what to do with a connection request, for example: ignore it, send a RST, or
// attempt to complete the 3-way handshake.
//
// The canonical way of using it is to pass the Forwarder.HandlePacket function
// to stack.SetTransportProtocolHandler.
type UDPForwarder struct {
	handler func(*UDPForwarderRequest)

	s *stack.Stack

	mu sync.Mutex
}

// NewForwarder allocates and initializes a new forwarder with the given
// maximum number of in-flight connection attempts. Once the maximum is reached
// new incoming connection requests will be ignored.
//
// If rcvWnd is set to zero, the default buffer size is used instead.
func NewUDPForwarder(s *stack.Stack, handler func(*UDPForwarderRequest)) *UDPForwarder {
	return &UDPForwarder{
		s:       s,
		handler: handler,
	}
}

// HandlePacket handles a packet if it is of interest to the forwarder (i.e., if
// it's a SYN packet), returning true if it's the case. Otherwise the packet
// is not handled and false is returned.
//
// This function is expected to be passed as an argument to the
// stack.SetTransportProtocolHandler function.
func (f *UDPForwarder) HandlePacket(r *stack.Route, id stack.TransportEndpointID, vv buffer.VectorisedView) bool {
	// Get the header then trim it from the view.
	hdr := header.UDP(vv.First())
	if int(hdr.Length()) > vv.Size() {
		// Malformed packet.
		return false
	}

	vv.TrimFront(header.UDPMinimumSize)

	var wq waiter.Queue

	ep, err := udp.NewConnectedEndpoint(f.s, r, id, &wq)
	if err != nil {
		panic(err)
	}

	go f.handler(&UDPForwarderRequest{
		forwarder: f,
		ep:        ep,
		wq:        &wq,
		payload:   hdr.Payload(),
		la: tcpip.FullAddress{
			Addr: id.LocalAddress,
			Port: id.LocalPort,
			NIC:  r.NICID(),
		},
		ra: tcpip.FullAddress{
			Addr: id.RemoteAddress,
			Port: id.RemotePort,
			NIC:  r.NICID(),
		},
	})

	return true
}

type UDPForwarderRequest struct {
	forwarder *UDPForwarder
	payload   []byte

	wq *waiter.Queue
	ep tcpip.Endpoint

	la tcpip.FullAddress
	ra tcpip.FullAddress
}

// ID returns the 4-tuple (src address, src port, dst address, dst port) that
// represents the connection request.
func (r *UDPForwarderRequest) ID() stack.TransportEndpointID {
	return stack.TransportEndpointID{
		LocalPort:     r.la.Port,
		LocalAddress:  r.la.Addr,
		RemotePort:    r.ra.Port,
		RemoteAddress: r.ra.Addr,
	}
}
