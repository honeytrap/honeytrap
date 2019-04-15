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
package ethernet

import (
	"encoding/binary"
	"net"
	"syscall"
)

type Frame struct {
	Source      net.HardwareAddr
	Destination net.HardwareAddr

	Type uint16

	Payload []byte
}

func Parse(data []byte) (*Frame, error) {
	eh := &Frame{
		Source:      make([]byte, 6),
		Destination: make([]byte, 6),
	}
	return eh, eh.Unmarshal(data)
}

func (f *Frame) Unmarshal(data []byte) error {
	copy(f.Destination[:], data[0:6])
	copy(f.Source[:], data[6:12])
	f.Type = binary.BigEndian.Uint16(data[12:14])
	f.Payload = data[14:]
	return nil
}

// Marshal returns the binary encoding of the IPv4 header h.
func (f *Frame) Marshal() ([]byte, error) {
	if f == nil {
		return nil, syscall.EINVAL
	}

	data := [14]byte{}
	copy(data[0:6], f.Destination)
	copy(data[6:12], f.Source)
	data[12] = uint8((f.Type >> 8) & 0xFF)
	data[13] = uint8(f.Type & 0xFF)
	return data[:], nil
}
