// +build linux

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
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/glycerine/rbuf"
)

// SocketState defines a int type.
type SocketState int

// contains different SocketState.
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

// Socket defines a object for representing a giving underrline socket
type Socket struct {
	laddr net.Addr
	raddr net.Addr

	rchan chan interface{}

	rbuffer *rbuf.FixedSizeRingBuf
	wbuffer *rbuf.FixedSizeRingBuf

	closed bool

	state *State
}

// LocalAddr returns local net.Addr.
func (s Socket) LocalAddr() net.Addr {
	return s.laddr
}

// RemoteAddr returns remote net.Addr.
func (s Socket) RemoteAddr() net.Addr {
	return s.raddr
}

func (s Socket) flush() {
	if s.closed {
		return
	}

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
	case <-time.After(time.Second * 60):
		return 0, errors.New("Read timeout occurred")
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
	s.state.write(p)
	// all writes are buffered in herer
	// shold trigger channel that will update state
	// don't send by socket self
	return len(p), nil
}

func (s Socket) close() {
	if s.closed {
		return
	}

	// close(s.rchan)

	s.closed = true
}

// Close closes the underline connection.
func (s Socket) Close() error {
	s.state.close()
	return nil
}

// NewSocket returns a new instance of Socket.
func (state *State) NewSocket(src, dst net.Addr) *Socket {
	return &Socket{
		state: state,

		laddr: dst,
		raddr: src,

		rchan: make(chan interface{}),

		// rbuffer: rbuf.NewFixedSizeRingBuf(65535),
		// wbuffer: rbuf.NewFixedSizeRingBuf(65535),
		rbuffer: rbuf.NewFixedSizeRingBuf(4096),
		wbuffer: nil, /*rbuf.NewFixedSizeRingBuf(10000), */

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
