package s7comm

import (
	"bytes"
	"encoding/binary"
)

func (T *TPKT) serialize(m []byte) (r []byte) {

	T.Version = 0x03
	T.Reserved = 0x00
	T.Length = uint16(len(m) + 0x04)

	rb := &bytes.Buffer{}

	TErr := binary.Write(rb, binary.BigEndian, T)
	mErr := binary.Write(rb, binary.BigEndian, m)

	if TErr != nil || mErr != nil {
		/* Print error message to console */
		return nil
	}
	return rb.Bytes()
}

func (T *TPKT) deserialize(m *[]byte) (verified bool) {
	Length := binary.BigEndian.Uint16((*m)[2:4])
	T.Version = (*m)[0]
	T.Reserved = (*m)[1]
	T.Length = Length

	if T.verify(*m) {
		*m = (*m)[4:]
		return true
	}
	return false
}

func (T *TPKT) verify(m []byte) (isTPKT bool) {
	if T.Version == 0x03 && T.Reserved == 0x00 && int(T.Length)-len(m) == 0 {
		return true
	}
	return false

}

type TPKT struct {
	Version  uint8
	Reserved uint8
	Length   uint16
}
