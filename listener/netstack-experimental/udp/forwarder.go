// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package udp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	udpg "gvisor.dev/gvisor/pkg/tcpip/transport/udp"

	"gvisor.dev/gvisor/pkg/waiter"
)

// Forwarder is a connection request forwarder, which allows clients to decide
// what to do with a connection request, for example: ignore it, send a RST, or
// attempt to complete the 3-way handshake.
//
// The canonical way of using it is to pass the Forwarder.HandlePacket function
// to stack.SetTransportProtocolHandler.
type Forwarder struct {
	handler func(*ForwarderRequest)

	s *stack.Stack

	gforwarder        *udpg.Forwarder
	gforwarderRequest *udpg.ForwarderRequest

	mu sync.Mutex
}

// NewForwarder allocates and initializes a new forwarder with the given
// maximum number of in-flight connection attempts. Once the maximum is reached
// new incoming connection requests will be ignored.
//
// If rcvWnd is set to zero, the default buffer size is used instead.
func NewForwarder(s *stack.Stack, handler func(*ForwarderRequest)) *Forwarder {

	f := &Forwarder{
		s:       s,
		handler: handler,
	}

	//This will get set when Forwarder.HandlePacket is called.
	//need to get a (gvisor)ForwarderRequest like this. Structs are made unexported by gvisor.
	gfr := udpg.NewForwarder(s, func(fr *udpg.ForwarderRequest) {
		f.gforwarderRequest = fr
	})

	f.gforwarder = gfr

	return f
}

/*
// HandlePacket handles a packet if it is of interest to the forwarder (i.e., if
// it's a SYN packet), returning true if it's the case. Otherwise the packet
// is not handled and false is returned.
//
// This function is expected to be passed as an argument to the
// stack.SetTransportProtocolHandler function.
func (f *Forwarder) HandlePacket(r *stack.Route, id stack.TransportEndpointID, netHeader buffer.View, vv buffer.VectorisedView) bool {
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

	go f.handler(&ForwarderRequest{
		forwarder: f,
		ep:        ep,
		wq:        &wq,
		payload:   hdr.Payload(),
	})

	return true
}
*/

// HandlePacket handles a packet if it is of interest to the forwarder (i.e., if
// it's a SYN packet), returning true if it's the case. Otherwise the packet
// is not handled and false is returned.
//
// This function is expected to be passed as an argument to the
// stack.SetTransportProtocolHandler function.
func (f *Forwarder) HandlePacket(r *stack.Route, id stack.TransportEndpointID, pkt tcpip.PacketBuffer) bool {
	// Get the header then trim it from the view.
	hdr := header.UDP(pkt.Data.First())
	if int(hdr.Length()) > pkt.Data.Size() {
		// Malformed packet.
		return false
	}

	pkt.Data.TrimFront(header.UDPMinimumSize)

	//set gforwarderRequest. Neccesary to use CreateEndpoint.
	f.gforwarder.HandlePacket(r, id, pkt)

	wq := &waiter.Queue{}

	ep, err := f.gforwarderRequest.CreateEndpoint(wq)
	if err != nil {
		return false
	}

	go f.handler(&ForwarderRequest{
		payload: hdr.Payload(),
		id:      id,
		route:   r,
		wq:      wq,
		ep:      ep,
	})

	return true
}

type ForwarderRequest struct {
	payload []byte
	id      stack.TransportEndpointID
	wq      *waiter.Queue
	ep      tcpip.Endpoint
	route   *stack.Route

	gfr *udpg.ForwarderRequest
}

// ID returns the 4-tuple (src address, src port, dst address, dst port) that
// represents the session request.
func (fr *ForwarderRequest) ID() stack.TransportEndpointID {
	return fr.gfr.ID()
}

/*
// ID returns the 4-tuple (src address, src port, dst address, dst port) that
// represents the connection request.
func (r *ForwarderRequest) ID() stack.TransportEndpointID {
	return stack.TransportEndpointID{
		LocalPort:     r.la.Port,
		LocalAddress:  r.la.Addr,
		RemotePort:    r.ra.Port,
		RemoteAddress: r.ra.Addr,
	}
}
*/

func (fr *ForwarderRequest) Payload() []byte {
	return fr.payload
}

func (fr *ForwarderRequest) Write(b []byte, addr *net.UDPAddr) (int, error) {
	v := buffer.NewView(len(b))
	copy(v, b)

	wopts := tcpip.WriteOptions{To: &tcpip.FullAddress{
		Addr: fr.id.RemoteAddress,
		Port: fr.id.RemotePort,
		NIC:  fr.route.NICID(),
	}}

	n, resCh, err := fr.ep.Write(tcpip.SlicePayload(v), wopts)
	if resCh != nil {
		select {
		case <-time.After(time.Millisecond * 1000):
			return int(n), fmt.Errorf("timeout")
		case <-resCh:
		}

		n, _, err = fr.ep.Write(tcpip.SlicePayload(v), wopts)
		if err != nil {
			return int(n), errors.New(err.String())
		}
		return int(n), nil
	}

	if err == tcpip.ErrWouldBlock {
		// Create wait queue entry that notifies a channel.
		waitEntry, notifyCh := waiter.NewChannelEntry(nil)
		fr.wq.EventRegister(&waitEntry, waiter.EventOut)
		defer fr.wq.EventUnregister(&waitEntry)
		for {
			select {
			case <-time.After(time.Millisecond * 1000):
				return int(n), fmt.Errorf("timeout")
			case <-notifyCh:
			}

			n, _, err = fr.ep.Write(tcpip.SlicePayload(v), wopts)
			if err != tcpip.ErrWouldBlock {
				break
			}
		}
	}

	if err == nil {
		return int(n), nil
	}

	return len(b), errors.New(err.String())
}
