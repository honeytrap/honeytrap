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
	"bytes"
	"encoding/binary"
	"net"

	"github.com/honeytrap/honeytrap/protocol"
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
