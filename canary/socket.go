package canary

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/glycerine/rbuf"
)

type SocketState int

const (
	SocketClosed SocketState = iota
	SocketListen
	SocketSynReceived
	SocketSynSent
	SocketEstablished
	SocketFinWait1
	SocketFinWait2
	SocketClosing
	SocketTimeWait
	SocketCloseWait
	SocketLastAck
)

func (ss SocketState) String() string {
	switch ss {
	case SocketClosed:
		return "SocketClosed"
	case SocketListen:
		return "SocketListen"
	case SocketSynReceived:
		return "SocketSynReceived"
	case SocketSynSent:
		return "SocketSynSent"
	case SocketEstablished:
		return "SocketEstablished"
	case SocketFinWait1:
		return "SocketFinWait1"
	case SocketFinWait2:
		return "SocketFinWait2"
	case SocketClosing:
		return "SocketClosing"
	case SocketTimeWait:
		return "SocketTimeWait"
	case SocketCloseWait:
		return "SocketCloseWait"
	case SocketLastAck:
		return "SocketLastAck"
	default:
		return fmt.Sprintf("Unknown state: %d", int(ss))
	}
}

type Socket struct {
	laddr net.Addr
	raddr net.Addr

	rchan chan interface{}

	rbuffer *rbuf.FixedSizeRingBuf
	wbuffer *rbuf.FixedSizeRingBuf

	closed bool
}

func (s Socket) LocalAddr() net.Addr {
	return s.laddr
}

func (s Socket) RemoteAddr() net.Addr {
	return s.raddr
}

func (s Socket) flush() {
	// non blocking channel
	select {
	case s.rchan <- []byte{}:
	default:
	}
}

func (s Socket) Read(p []byte) (n int, err error) {
	if !s.closed {
	} else if s.rbuffer.Avail() == 0 {
		return 0, io.EOF
	}

	n, _ = s.rbuffer.Read(p)
	if n > 0 {
		return
	}

	// timeout
	// close (io.EOF)
	select {
	case <-s.rchan:
		/*
			case <-time.After(time.Second * 1):
				return 0, errors.New("Timeout occured")
		*/
	}

	n, _ = s.rbuffer.Read(p)
	return
}

func (s Socket) write(p []byte) (n int, err error) {
	s.rbuffer.Write(p)
	return len(p), nil
}

func (s Socket) read(p []byte) (n int, err error) {
	// read will enable write
	return len(p), nil
}

func (s Socket) Write(p []byte) (n int, err error) {
	// all writes are buffered in herer
	// shold trigger channel that will update state
	// don't send by socket self
	return len(p), nil
}

func (s Socket) close() {
	if s.closed {
		return
	}

	close(s.rchan)

	s.closed = true
}

func (s Socket) Close() error {
	return nil
}

func NewSocket(src, dst net.Addr) *Socket {
	return &Socket{
		laddr: src,
		raddr: dst,

		rchan: make(chan interface{}),

		rbuffer: rbuf.NewFixedSizeRingBuf(65535),
		wbuffer: rbuf.NewFixedSizeRingBuf(65535),

		closed: false,
	}
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (s Socket) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (s Socket) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (s Socket) SetWriteDeadline(t time.Time) error {
	return nil
}
