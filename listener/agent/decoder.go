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

	"github.com/honeytrap/protocol"
)

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		protocol.NewDecoder(bytes.NewBuffer(data), binary.LittleEndian),
	}
}

type Decoder struct {
	*protocol.Decoder
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

func (d *Decoder) ReadAddr() net.Addr {
	if d.LastError != nil {
		return nil
	}

	proto := d.ReadUint8()

	data := d.ReadData()
	ip := net.IP(data)

	port := d.ReadUint16()

	if proto == 6 {
		return &net.TCPAddr{
			IP:   ip,
			Port: port,
		}
	} else if proto == 17 {
		return &net.UDPAddr{
			IP:   ip,
			Port: port,
		}
	} else {
		// unsupported protocol
		return nil
	}
}
