package nscanary

import (
	"errors"
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

// Reimplementation of gonet.TCPListener which can check for TLS requests
// when accepting new tcp connections.

// A TCPListener is a wrapper around a TCP tcpip.Endpoint that implements
// net.Listener.
type TCPListener struct {
	stack  *stack.Stack
	ep     tcpip.Endpoint
	wq     *waiter.Queue
	cancel chan struct{}
}

// NewTCPListener creates a new TCPListener from a listening tcpip.Endpoint.
func NewTCPListener(s *stack.Stack, wq *waiter.Queue, ep tcpip.Endpoint) *TCPListener {
	return &TCPListener{
		stack:  s,
		ep:     ep,
		wq:     wq,
		cancel: make(chan struct{}),
	}
}

// ListenTCP creates a new TCPListener.
func ListenTCP(s *stack.Stack, addr tcpip.FullAddress, network tcpip.NetworkProtocolNumber) (*TCPListener, error) {
	// Create a TCP endpoint, bind it, then start listening.
	var wq waiter.Queue
	ep, err := s.NewEndpoint(tcp.ProtocolNumber, network, &wq)
	if err != nil {
		return nil, errors.New(err.String())
	}

	fmt.Printf("network = %+v\n", network)
	fmt.Printf("addr = %+v\n", addr)
	fmt.Printf("ep.Info() = %+v\n", ep.Info())

	if err := ep.Bind(addr); err != nil {
		ep.Close()
		return nil, &net.OpError{
			Op:   "bind",
			Net:  "tcp",
			Addr: fullToTCPAddr(addr),
			Err:  errors.New(err.String()),
		}
	}

	fmt.Println("start ep.Listen(10)")

	if err := ep.Listen(10); err != nil {
		ep.Close()
		return nil, &net.OpError{
			Op:   "listen",
			Net:  "tcp",
			Addr: fullToTCPAddr(addr),
			Err:  errors.New(err.String()),
		}
	}

	return NewTCPListener(s, &wq, ep), nil
}

// Accept is a copy of gonet.TCPListener.Accept with added TLS detection.
// returns an extra boolean indicating if it is a TLS request.
func (l *TCPListener) Accept() (net.Conn, bool, error) {
	n, wq, err := l.ep.Accept()

	if err == tcpip.ErrWouldBlock {
		// Create wait queue entry that notifies a channel.
		waitEntry, notifyCh := waiter.NewChannelEntry(nil)
		l.wq.EventRegister(&waitEntry, waiter.EventIn)
		defer l.wq.EventUnregister(&waitEntry)

		for {
			n, wq, err = l.ep.Accept()

			if err != tcpip.ErrWouldBlock {
				break
			}

			select {
			case <-l.cancel:
				return nil, false, errors.New("operation canceled")
			case <-notifyCh:
			}
		}
	}

	if err != nil {
		return nil, false, &net.OpError{
			Op:   "accept",
			Net:  "tcp",
			Addr: l.Addr(),
			Err:  errors.New(err.String()),
		}
	}

	var signature [3]byte
	var isTLS bool

	vec := [][]byte{signature[:]}
	_, _, _ = l.ep.Peek(vec)
	if signature[0] == 0x16 && signature[1] == 0x03 && signature[2] <= 0x03 {
		isTLS = true
	}

	return gonet.NewTCPConn(wq, n), isTLS, nil
}

// Addr implements net.Listener.Addr.
func (l *TCPListener) Addr() net.Addr {
	a, err := l.ep.GetLocalAddress()
	if err != nil {
		return nil
	}
	return fullToTCPAddr(a)
}

func fullToTCPAddr(addr tcpip.FullAddress) *net.TCPAddr {
	return &net.TCPAddr{IP: net.IP(addr.Addr), Port: int(addr.Port)}
}
