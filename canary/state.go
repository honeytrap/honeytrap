package canary

import (
	"math/rand"
	"net"
)

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

type StateTable []*State

func (st *StateTable) Add(state *State) {
	*st = append(*st, state)
}

// GetState will return the state for the ip, port combination
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
