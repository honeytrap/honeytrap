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
