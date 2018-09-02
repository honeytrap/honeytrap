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
package utils

import (
	"net"
	"sync"
)

func PeekConnection(conn net.Conn) *PeekConn {
	return &PeekConn{
		conn,
		[]byte{},
		sync.Mutex{},
	}
}

// PeekConn struct, allows peeking a connection by writing peeked content to a buffer. Returning this content first when reading
type PeekConn struct {
	net.Conn

	buffer []byte
	m      sync.Mutex
}

// Peek writes the buffer from the connection, returning the amount of bytes written
func (pc *PeekConn) Peek(p []byte) (int, error) {
	pc.m.Lock()
	defer pc.m.Unlock()

	n, err := pc.Conn.Read(p)

	pc.buffer = append(pc.buffer, p[:n]...)
	return n, err
}

// Read writes the buffer from the connection, first writing bytes that have been peeked
func (pc *PeekConn) Read(p []byte) (n int, err error) {
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
