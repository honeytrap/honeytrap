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
	"encoding/binary"
	"io"
	"net"

	"github.com/honeytrap/protocol"
)

func NewEncoder(w io.Writer, bo binary.ByteOrder) *Encoder {
	return &Encoder{
		protocol.NewEncoder(w, bo),
	}
}

type Encoder struct {
	*protocol.Encoder
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
		e.WriteUint8(6)
	} else if ua, ok := address.(*net.UDPAddr); ok {
		ip = ua.IP
		port = ua.Port
		e.WriteUint8(17)
	}

	e.WriteData(ip)
	e.WriteUint16(port)
}
