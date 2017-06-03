package canary

import (
	"io"
	"net"

	"github.com/glycerine/rbuf"
)

type Socket struct {
	laddr net.IP
	raddr net.IP

	rchan chan interface{}

	rbuffer *rbuf.FixedSizeRingBuf
	wbuffer *rbuf.FixedSizeRingBuf

	closed bool
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

func NewSocket(src, dst net.IP) *Socket {
	return &Socket{
		laddr: src,
		raddr: dst,

		rchan: make(chan interface{}),

		rbuffer: rbuf.NewFixedSizeRingBuf(65535),
		wbuffer: rbuf.NewFixedSizeRingBuf(65535),

		closed: false,
	}
}
