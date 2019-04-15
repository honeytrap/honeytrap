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
package udp

import (
	"encoding/binary"
	"fmt"
)

type Header struct {
	Source      uint16
	Destination uint16
	Length      uint16
	Checksum    uint16
	Payload     []byte
}

func Unmarshal(data []byte) (*Header, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("Incorrect UDP header size: %d", len(data))
	}

	hdr := Header{}
	hdr.Source = binary.BigEndian.Uint16(data[0:2])
	hdr.Destination = binary.BigEndian.Uint16(data[2:4])
	hdr.Length = binary.BigEndian.Uint16(data[4:6])
	hdr.Checksum = binary.BigEndian.Uint16(data[6:8])
	hdr.Payload = data[8:]

	if len(data) != int(hdr.Length) {
		return nil, fmt.Errorf("UDP payload length and size doesn't match, got %d, expected %d", len(data), hdr.Length)
	}

	return &hdr, nil
}

func (hdr *Header) String() string {
	return fmt.Sprintf("sport=%d, sdest=%d, length=%d, checksum=%x",
		hdr.Source, hdr.Destination, hdr.Length, hdr.Checksum)
}

func (hdr *Header) Marshal() ([]byte, error) {
	buf := make([]byte, 8+len(hdr.Payload))
	binary.BigEndian.PutUint16(buf[0:2], hdr.Source)
	binary.BigEndian.PutUint16(buf[2:4], hdr.Destination)
	binary.BigEndian.PutUint16(buf[4:6], hdr.Length)
	binary.BigEndian.PutUint16(buf[6:8], hdr.Checksum)
	copy(buf[8:], hdr.Payload)
	return buf, nil
}
