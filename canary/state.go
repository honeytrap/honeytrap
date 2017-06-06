package canary

import (
	"math/rand"
	"net"
)

// State defines a struct for holding connection data and address.
type State struct {
	// interface?

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
	ReceiveWindow uint32
	// RCV.UP  - receive urgent pointer
	ReceiveUrgentPointer uint32

	// IRS     - initial receive sequence number
	InitialReceiveSequenceNumber uint32
}

// StateTable defines a slice of States type.
type StateTable []*State

// Add adds the state into the table.
func (st *StateTable) Add(state *State) {
	*st = append(*st, state)
}

// Get will return the state for the ip, port combination
func (st *StateTable) Get(SrcIP, DestIP net.IP, SrcPort, DestPort uint16) *State {
	for _, state := range *st {
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

// NewState returns a new instance of a State.
func NewState(src net.IP, srcPort uint16, dest net.IP, dstPort uint16) *State {
	return &State{
		SrcIP:   src,
		SrcPort: srcPort,

		DestIP:   dest,
		DestPort: dstPort,

		ID: rand.Uint32(),

		RecvNext:                  0,
		InitialSendSequenceNumber: rand.Uint32(),
	}
}
