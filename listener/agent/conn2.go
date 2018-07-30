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
	"encoding"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
)

func Conn2(c net.Conn) *conn2 {
	return &conn2{c}
}

type conn2 struct {
	net.Conn
}

func (c *conn2) Handshake() error {
	return nil
}

func (c *conn2) receive() (interface{}, error) {
	buff := make([]byte, 1)

	if _, err := c.Conn.Read(buff); err != nil {
		return nil, err
	}

	msgType := int(buff[0])

	var o encoding.BinaryUnmarshaler

	switch msgType {
	case TypeHello:
		o = &Hello{}
	case TypeHandshake:
		o = &Handshake{}
	case TypeHandshakeResponse:
		o = &HandshakeResponse{}
	case TypeReadWriteTCP:
		o = &ReadWriteTCP{}
	case TypeReadWriteUDP:
		o = &ReadWriteUDP{}
	case TypePing:
		o = &Ping{}
	case TypeEOF:
		o = &EOF{}
	default:
		return nil, fmt.Errorf("Unsupported message receive type %d", msgType)
	}

	buff = make([]byte, 2)

	if _, err := c.Conn.Read(buff); err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint16(buff)

	buff = make([]byte, size)

	if _, err := c.Conn.Read(buff); err != nil {
		return nil, err
	}

	if err := o.UnmarshalBinary(buff[:]); err != nil {
		return nil, err
	}

	return o, nil
}

func (c conn2) send(o encoding.BinaryMarshaler) error {
	// write type
	switch o.(type) {
	case Hello:
		c.Conn.Write([]byte{uint8(TypeHello)})
	case Handshake:
		c.Conn.Write([]byte{uint8(TypeHandshake)})
	case HandshakeResponse:
		c.Conn.Write([]byte{uint8(TypeHandshakeResponse)})
	case Ping:
		c.Conn.Write([]byte{uint8(TypePing)})
	case ReadWriteTCP:
		c.Conn.Write([]byte{uint8(TypeReadWriteTCP)})
	case ReadWriteUDP:
		c.Conn.Write([]byte{uint8(TypeReadWriteUDP)})
	case EOF:
		c.Conn.Write([]byte{uint8(TypeEOF)})
	default:
		return fmt.Errorf("Unsupported message type send %s", reflect.TypeOf(o))
	}

	data, err := o.MarshalBinary()
	if err != nil {
		return err
	}

	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff[0:2], uint16(len(data)))

	if _, err := c.Conn.Write(buff); err != nil {
		return err
	}

	if _, err := c.Conn.Write(data[:]); err != nil {
		return err
	}

	return nil
}
