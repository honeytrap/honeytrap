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
package decoder

import (
	"bytes"
	"encoding/binary"
)

type EncoderType interface {
	WriteUint8(b byte)

	WriteUint16(v int16)
	WriteUint32(v int32)

	WriteData(v string, zero bool)
}

type Encoder struct {
	bytes.Buffer
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) WriteUint8(b byte) {
	_ = e.WriteByte(b)
}

func (e *Encoder) WriteUint16(v int16) {
	b := [2]byte{}
	binary.BigEndian.PutUint16(b[:], uint16(v))
	e.Write(b[:])
}

func (e *Encoder) WriteUint32(v int32) {
	b := [4]byte{}
	binary.BigEndian.PutUint32(b[:], uint32(v))
	e.Write(b[:])
}

// if zero is true, write zero length
func (e *Encoder) WriteData(v string, zero bool) {

	if zero {
		e.WriteUint16(0)
	} else {
		e.WriteUint16(int16(len(v)))
		e.Write([]byte(v))
	}
}
