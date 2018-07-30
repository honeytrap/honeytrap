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
package agent

import (
	"io"
	"net"
	"sync"
	"time"
)

var (
	noDeadline = make(<-chan time.Time)
)

type agentConnection struct {
	Laddr net.Addr
	Raddr net.Addr

	buff   []byte
	closed bool

	readTimeout  time.Time
	writeTimeout time.Time

	in chan []byte

	out chan interface{}

	m sync.Mutex
}

func (dc *agentConnection) receive(data []byte) {
	dc.m.Lock()
	defer dc.m.Unlock()

	if dc.closed {
		return
	}

	dc.buff = append(dc.buff, data...)

	select {
	case dc.in <- []byte{}: // v.Payload {
	default:
	}
}

func (dc *agentConnection) Read(b []byte) (int, error) {
	dc.m.Lock()
	if len(dc.buff) != 0 {
		n := copy(b[:], dc.buff[0:])
		dc.buff = dc.buff[n:]
		dc.m.Unlock()
		return n, nil
	}
	dc.m.Unlock()

	after := noDeadline

	if !dc.readTimeout.IsZero() {
		after = time.After(time.Until(dc.readTimeout))
	}

	select {
	case <-after:
		return 0, ErrTimeout
	case _, ok := <-dc.in:
		if !ok {
			log.Errorf("Error reading from channel, return EOF")
			return 0, io.EOF
		}

		dc.m.Lock()
		n := copy(b[:], dc.buff[0:])
		dc.buff = dc.buff[n:]
		dc.m.Unlock()

		return n, nil
	}
}

func (dc *agentConnection) Write(b []byte) (int, error) {
	dc.m.Lock()
	defer dc.m.Unlock()

	payload := make([]byte, len(b))

	copy(payload, b)

	p := ReadWriteTCP{
		Laddr:   dc.LocalAddr(),
		Raddr:   dc.RemoteAddr(),
		Payload: payload[:],
	}

	after := noDeadline
	if !dc.writeTimeout.IsZero() {
		after = time.After(time.Until(dc.writeTimeout))
	}

	select {
	case <-after:
		return 0, ErrTimeout
	case dc.out <- p:
	}

	return len(b), nil
}

func (dc *agentConnection) Close() error {
	dc.m.Lock()
	defer dc.m.Unlock()

	if dc.closed {
		return nil
	}

	p := EOF{
		Laddr: dc.LocalAddr(),
		Raddr: dc.RemoteAddr(),
	}

	dc.out <- p

	dc.closed = true
	close(dc.in)

	return nil
}

func (dc *agentConnection) LocalAddr() net.Addr {
	return dc.Laddr
}

func (dc *agentConnection) RemoteAddr() net.Addr {
	return dc.Raddr
}

func (dc *agentConnection) SetDeadline(t time.Time) error {
	dc.SetReadDeadline(t)
	dc.SetWriteDeadline(t)
	return nil
}

func (dc *agentConnection) SetReadDeadline(t time.Time) error {
	dc.readTimeout = t
	return nil
}

func (dc *agentConnection) SetWriteDeadline(t time.Time) error {
	dc.writeTimeout = t
	return nil
}
