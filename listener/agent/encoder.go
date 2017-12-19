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
	"bytes"
	"encoding/binary"
	"net"
)

type Encoder struct {
	bytes.Buffer
}

func (e *Encoder) WriteUint8(v int) {
	e.WriteByte(uint8(v))
}

func (e *Encoder) WriteUint16(v int) {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	e.Write(b[:])
}

func (e *Encoder) WriteString(s string) {
	e.WriteData([]byte(s))
}

func (e *Encoder) WriteData(data []byte) {
	e.WriteUint16(len(data))
	e.Write(data)
}

func (e *Encoder) WriteAddr(address net.Addr) {
	var ip net.IP
	var port int

	if ta, ok := address.(*net.TCPAddr); ok {
		ip = ta.IP
		port = ta.Port
	}

	e.WriteData(ip)
	e.WriteUint16(port)
}
