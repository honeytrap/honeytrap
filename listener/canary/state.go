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
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/listener/canary/tcp"
)

// State defines a struct for holding connection data and address.
type State struct {
	// interface?
	c *Canary

	m sync.Mutex

	SrcIP   net.IP
	SrcPort uint16

	DestIP   net.IP
	DestPort uint16

	ID uint32

	LastAcked uint32

	// /proc/net/tcp

	socket *Socket

	State SocketState
	// contains tx_queue
	// contains rx_queue

	// SND.UNA - send unacknowledged
	SendUnacknowledged uint32
	// SND.NXT - send next
	SendNext uint32
	// SND.WND - send window
	SendWindow uint32
	// SND.UP  - send urgent pointer
	SendUrgentPointer uint32

	// SND.WL1 - segment sequence number used for last window update
	SendWL1 uint32

	// SND.WL2 - segment acknowledgment number used for last window update
	SendWL2 uint32

	// ISS     - initial send sequence number
	InitialSendSequenceNumber uint32

	// RCV.NXT - receive next
	RecvNext uint32
	// RCV.WND - receive window
	ReceiveWindow uint16
	// RCV.UP  - receive urgent pointer
	ReceiveUrgentPointer uint32

	// IRS     - initial receive sequence number
	InitialReceiveSequenceNumber uint32

	t time.Time
}

func (s *State) write(data []byte) {
	// I think tstate should not write packets directly,
	// instead to write queue or buffer
	s.m.Lock()
	defer s.m.Unlock()

	s.c.send(s, data, tcp.PSH|tcp.ACK)
	s.SendNext += uint32(len(data))
}

func (s *State) close() {
	// I think tstate should not write packets directly,
	// instead to write queue or buffer
	s.m.Lock()
	defer s.m.Unlock()

	// Queue this until all preceding SENDs have been segmentized, then
	// form a FIN segment and send it.  In any case, enter FIN-WAIT-1
	// state.
	s.c.send(s, []byte{}, tcp.FIN|tcp.ACK)
	s.SendNext++

	s.State = SocketFinWait1
}

// StateTable defines a slice of States type.
type StateTable [65535]*State

// Add adds the state into the table.
func (st *StateTable) Add(state *State) {
	for i := range *st {
		if (*st)[i] == nil {
			// slot not taken
		} else if (*st)[i].State == SocketTimeWait {
			// reuse socket timewait
		} else {
			continue
		}

		(*st)[i] = state
		return
	}

	now := time.Now()

	for i := range *st {
		if now.Sub((*st)[i].t) > 30*time.Second {
			// inactive
		} else {
			continue
		}

		(*st)[i] = state
		return
	}

	// we don't have enough space in the state table, and
	// there are no inactive entries
	panic("Statetable full")
}

// Get will return the state for the ip, port combination
func (st *StateTable) Get(SrcIP, DestIP net.IP, SrcPort, DestPort uint16) *State {
	for _, state := range *st {
		if state == nil {
			continue
		}

		if state.SrcPort != SrcPort && state.DestPort != SrcPort {
			continue
		}

		if state.DestPort != DestPort && state.SrcPort != DestPort {
			continue
		}

		// comparing ipv6 with ipv4 now
		if !state.SrcIP.Equal(SrcIP) && !state.DestIP.Equal(SrcIP) {
			continue
		}

		if !state.DestIP.Equal(DestIP) && !state.SrcIP.Equal(DestIP) {
			continue
		}

		return state
	}

	return nil // state
}

func (st *StateTable) Remove(s *State) {
	for i := range *st {
		if (*st)[i] != s {
			continue
		}

		(*st)[i] = nil
		break
	}
}

// NewState returns a new instance of a State.
func (c *Canary) NewState(src net.IP, srcPort uint16, dest net.IP, dstPort uint16) *State {
	return &State{
		c: c,

		SrcIP:   src,
		SrcPort: srcPort,

		DestIP:   dest,
		DestPort: dstPort,

		ID: rand.Uint32(),

		ReceiveWindow: 65535,

		RecvNext:                  0,
		InitialSendSequenceNumber: rand.Uint32(),

		t: time.Now(),

		m: sync.Mutex{},
	}
}
