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

	RecvNext uint32
	SendNext uint32

	LastAcked uint32

	// /proc/net/tcp

	socket *Socket

	// contains tx_queue
	// contains rx_queue
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

		RecvNext: 0,
		SendNext: rand.Uint32(),
	}
}
