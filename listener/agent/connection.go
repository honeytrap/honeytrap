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

type agentConnection struct {
	Laddr net.Addr
	Raddr net.Addr

	buff   []byte
	closed bool

	in chan []byte

	out chan ReadWrite

	m sync.Mutex
}

func (dc *agentConnection) Read(b []byte) (int, error) {
	if len(dc.buff) >= len(b) {
		dc.m.Lock()
		defer dc.m.Unlock()

		n := copy(b[:], dc.buff[0:])
		dc.buff = dc.buff[n:]

		return n, nil
	}

	data, ok := <-dc.in
	if !ok {
		return 0, io.EOF
	}

	dc.m.Lock()
	defer dc.m.Unlock()

	dc.buff = append(dc.buff, data[:]...)

	n := copy(b[:], dc.buff[0:])
	dc.buff = dc.buff[n:]

	return n, nil
}

func (dc *agentConnection) Write(b []byte) (int, error) {
	dc.m.Lock()
	defer dc.m.Unlock()

	payload := make([]byte, len(b))

	copy(payload, b)

	p := ReadWrite{
		Laddr:   dc.LocalAddr(),
		Raddr:   dc.RemoteAddr(),
		Payload: payload[:],
	}

	dc.out <- p
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

	if data, err := p.MarshalBinary(); err == nil {
		dc.in <- data
	} else {
		log.Error("Error marshaling hello: %s", err.Error())
		return err
	}

	close(dc.in)

	dc.closed = true
	return nil
}

func (dc *agentConnection) LocalAddr() net.Addr {
	return dc.Laddr
}

func (dc *agentConnection) RemoteAddr() net.Addr {
	return dc.Raddr
}

func (dc *agentConnection) SetDeadline(t time.Time) error {
	return nil
}

func (dc *agentConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (dc *agentConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
