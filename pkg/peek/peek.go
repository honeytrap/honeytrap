//Package peek provides Peek capabilty to a net.Conn
package peek

import (
	"net"
	"sync"
)

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn,
		[]byte{},
		sync.Mutex{},
	}
}

type Conn struct {
	net.Conn

	buffer []byte
	m      sync.Mutex
}

func (pc *Conn) Peek(p []byte) (int, error) {
	pc.m.Lock()
	defer pc.m.Unlock()

	n, err := pc.Conn.Read(p)

	pc.buffer = append(pc.buffer, p[:n]...)
	return n, err
}

func (pc *Conn) Read(p []byte) (n int, err error) {
	pc.m.Lock()
	defer pc.m.Unlock()

	// first serve from peek buffer
	if len(pc.buffer) > 0 {
		bn := copy(p, pc.buffer)
		pc.buffer = pc.buffer[bn:]
		return bn, nil
	}

	return pc.Conn.Read(p)
}
