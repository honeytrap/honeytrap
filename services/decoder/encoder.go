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
