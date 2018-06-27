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
	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/stack"
)

func NewFilter(lower tcpip.LinkEndpointID) tcpip.LinkEndpointID {
	return stack.RegisterLinkEndpoint(&filterEndpoint{
		lower: stack.FindLinkEndpoint(lower),
	})
}

type filterEndpoint struct {
	lower      stack.LinkEndpoint
	dispatcher stack.NetworkDispatcher
}

// WritePacket writes outbound packets to the file descriptor. If it is not
// currently writable, the packet is dropped.
func (e *filterEndpoint) WritePacket(r *stack.Route, hdr *buffer.Prependable, payload buffer.View, protocol tcpip.NetworkProtocolNumber) *tcpip.Error {
	// https://godoc.org/golang.org/x/net/bpf
	return e.lower.WritePacket(r, hdr, payload, protocol)
}

// Attach implements the stack.LinkEndpoint interface. It saves the dispatcher
// and registers with the lower filterEndpoint as its dispatcher so that "e" is called
// for inbound packets.
func (e *filterEndpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.dispatcher = dispatcher
	e.lower.Attach(e)
}

func (e *filterEndpoint) DeliverNetworkPacket(linkEP stack.LinkEndpoint, remoteLinkAddr tcpip.LinkAddress, protocol tcpip.NetworkProtocolNumber, vv *buffer.VectorisedView) {
	e.dispatcher.DeliverNetworkPacket(e, remoteLinkAddr, protocol, vv)
}

// IsAttached implements stack.LinkEndpoint.IsAttached.
func (e *filterEndpoint) IsAttached() bool {
	return e.dispatcher != nil
}

// MTU implements stack.LinkEndpoint.MTU. It just forwards the request to the
// lower filterEndpoint.
func (e *filterEndpoint) MTU() uint32 {
	return e.lower.MTU()
}

// Capabilities implements stack.LinkEndpoint.Capabilities. It just forwards the
// request to the lower filterEndpoint.
func (e *filterEndpoint) Capabilities() stack.LinkEndpointCapabilities {
	return e.lower.Capabilities()
}

// MaxHeaderLength implements the stack.LinkEndpoint interface. It just forwards
// the request to the lower filterEndpoint.
func (e *filterEndpoint) MaxHeaderLength() uint16 {
	return e.lower.MaxHeaderLength()
}

func (e *filterEndpoint) LinkAddress() tcpip.LinkAddress {
	return e.lower.LinkAddress()
}
