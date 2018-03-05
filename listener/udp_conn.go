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
package listener

import (
	"net"
	"time"
)

type DummyUDPConn struct {
	Buffer []byte

	Laddr *net.UDPAddr
	Raddr *net.UDPAddr

	Fn func(b []byte, addr *net.UDPAddr) (int, error)
}

func (dc *DummyUDPConn) Read(b []byte) (int, error) {
	n := copy(b, dc.Buffer)
	dc.Buffer = dc.Buffer[n:]
	return n, nil
}

func (dc *DummyUDPConn) Write(b []byte) (int, error) {
	if dc.Fn == nil {
		return len(b), nil
	}

	return dc.Fn(b[:], dc.Raddr)
}

func (dc *DummyUDPConn) Close() error {
	return nil
}

func (dc *DummyUDPConn) LocalAddr() net.Addr {
	return dc.Laddr
}

func (dc *DummyUDPConn) RemoteAddr() net.Addr {
	return dc.Raddr
}

func (dc *DummyUDPConn) SetDeadline(t time.Time) error {
	return nil
}

func (dc *DummyUDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (dc *DummyUDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
