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
