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
