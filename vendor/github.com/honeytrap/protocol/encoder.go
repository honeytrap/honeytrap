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

func NewEncoder(w io.Writer, bo binary.ByteOrder) *Encoder {
	return &Encoder{
		bufio.NewWriter(w),
		bo,
	}
}

type Encoder struct {
	*bufio.Writer

	bo binary.ByteOrder
}

func (e *Encoder) WriteUint8(v int) {
	e.WriteByte(uint8(v))
}

func (e *Encoder) WriteUint16(v int) {
	b := [2]byte{}
	e.bo.PutUint16(b[:], uint16(v))
	e.Write(b[:])
}
