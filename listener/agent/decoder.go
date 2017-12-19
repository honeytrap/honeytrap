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

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		Buffer:    bytes.NewBuffer(data),
		LastError: nil,
	}
}

type Decoder struct {
	*bytes.Buffer

	LastError error
}

func (d *Decoder) ReadData() []byte {
	if d.LastError != nil {
		return []byte{}
	}

	l := d.ReadUint16()

	buffer := make([]byte, l)
	if _, err := d.Read(buffer[:]); err != nil {
		d.LastError = err
		return []byte{}
	}

	return buffer
}

func (d *Decoder) ReadString() string {
	if d.LastError != nil {
		return ""
	}

	l := d.ReadUint16()

	buffer := make([]byte, l)
	if _, err := d.Read(buffer[:]); err != nil {
		d.LastError = err
		return ""
	}

	return string(buffer)
}

func (d *Decoder) ReadUint16() int {
	if d.LastError != nil {
		return 0
	}

	buffer := [2]byte{}
	if _, err := d.Read(buffer[:]); err != nil {
		d.LastError = err
		return 0
	}

	return int(binary.LittleEndian.Uint16(buffer[:]))
}

func (d *Decoder) ReadUint8() int {
	if d.LastError != nil {
		return 0
	}

	b, err := d.ReadByte()
	if err != nil {
		d.LastError = err
	}

	return int(b)
}

func (d *Decoder) ReadAddr() net.Addr {
	if d.LastError != nil {
		return nil
	}

	data := d.ReadData()
	ip := net.IP(data)
	port := d.ReadUint16()

	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}
}
