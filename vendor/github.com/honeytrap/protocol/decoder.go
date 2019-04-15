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
package protocol

import (
	"bufio"
	"encoding/binary"
	"io"
)

func NewDecoder(r io.Reader, bo binary.ByteOrder) *Decoder {
	return &Decoder{
		Reader:    bufio.NewReader(r),
		bo:        bo,
		LastError: nil,
	}
}

type Decoder struct {
	*bufio.Reader

	bo        binary.ByteOrder
	LastError error
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

func (d *Decoder) ReadUint16() int {
	if d.LastError != nil {
		return 0
	}

	buffer := [2]byte{}
	if _, err := d.Read(buffer[:]); err != nil {
		d.LastError = err
		return 0
	}

	return int(d.bo.Uint16(buffer[:]))
}
