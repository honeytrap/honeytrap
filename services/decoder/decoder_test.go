package decoder

import (
	"encoding/binary"
	"testing"
)

func TestHasBytes(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if err := dec.HasBytes(1); err != nil {
		t.Error(err)
	}
	if err := dec.HasBytes(len(bs)); err != nil {
		t.Error(err)
	}
	if err := dec.HasBytes(len(bs) + 1); err == nil {
		t.Error("HasBytes(max+1) returns no error while out of index")
	}
	if err := dec.HasBytes(0); err == nil {
		t.Error("HasBytes(0) returns no error with zero value")
	}
}

func TestByte(t *testing.T) {
	bs := []byte{0xff, 0x01, 0, 0, 0, 0, 0, 0, 0x01}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if b, _ := dec.PeekByte(); b != 0xff {
		t.Errorf("PeekByte: expected 0xff, got %d", b)
	}
	if b, _ := dec.Byte(); b != 0xff { // Read the fist byte again
		t.Errorf("Byte: expected 0xff, got %d", b)
	}
	if b, _ := dec.Uint8(); b != 1 { // Second byte
		t.Errorf("Byte: expected 1, got %d", b)
	}
}

func TestInt16(t *testing.T) {
	bs := []byte{0, 0x01, 0, 0x02, 0, 0x03}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if b, _ := dec.Int16(); b != 1 {
		t.Errorf("int16: expected 1, got %d", b)
	}
	if b, _ := dec.Uint16(); b != 2 {
		t.Errorf("int16: expected 2, got %d", b)
	}
}

func TestInt32(t *testing.T) {
	bs := []byte{0, 0, 0, 0x01, 0, 0, 0, 0x02}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if b, _ := dec.Int32(); b != 1 {
		t.Errorf("expect 1, got %v", b)
	}
	if b, _ := dec.Uint32(); b != 2 {
		t.Errorf("expect 2, got %v", b)
	}
}

func TestInt64(t *testing.T) {
	bs := []byte{0, 0, 0, 0, 0, 0, 0, 0x01, 0, 0, 0, 0, 0, 0, 0, 0x02}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if b, _ := dec.Int64(); b != 1 {
		t.Errorf("expect 1, got %v", b)
	}
	if b, _ := dec.Uint64(); b != 2 {
		t.Errorf("expect 2, got %v", b)
	}
}

func TestNew(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	if dec.Offset() != 0 || dec.StartOffset() != 0 {
		t.Errorf("NewDefaultDecoder: offset: %d, startoffset: %d, expected 0 for both!", dec.Offset(), dec.StartOffset())
	}

	if avail := dec.Available(); avail != len(bs) {
		t.Errorf("Available gives wrong size got %d, expected %d", avail, len(bs))
	}

}

func TestSeek(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	_, _ = dec.Seek(2)
	want := byte(0x03)
	if got, _ := dec.PeekByte(); got != want {
		t.Errorf("Seek(2): got %d, expected %d", got, want)
	}

	_, _ = dec.Seek(len(bs) - 1) // Seek to the end of data
	want = 0x05
	if got, _ := dec.PeekByte(); got != want {
		t.Errorf("Seek(max): got %d, expected %d", got, want)
	}

	if _, err := dec.Seek(len(bs)); err == nil {
		t.Errorf("Seek(max): No error return! Decoder: %v", dec)
	}

	if _, err := dec.Seek(-1); err == nil {
		t.Errorf("Seek(-1): No error return! Decoder: %v", dec)
	}
}

func TestSkip(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	_ = dec.Skip(2)
	want := byte(0x03)
	if got, _ := dec.PeekByte(); got != want {
		t.Errorf("Skip(2): got %d, expected %d", got, want)
	}

	if err := dec.Skip(len(bs)); err == nil {
		t.Errorf("Skip(max): No error return! Decoder: %v", err)
	}

	if err := dec.Skip(-100); err == nil {
		t.Errorf("Skip(-100): No error Return! Decoder: %v", dec)
	}
}

func TestAlign(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	_, _ = dec.Seek(2)
	dec.Align(4)
	want := byte(0x04)
	if got, _ := dec.PeekByte(); got != want {
		t.Errorf("Align(4): got %d, expected %d", got, want)
	}
}

func TestAString(t *testing.T) {
	bs := []byte{74, 69, 82, 82, 89, 65}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	want := "JERRY"
	if got, _ := dec.AString(5); got != want {
		t.Errorf("AString(5): got %v, want %v", got, want)
	}
	want = "A"
	if got, _ := dec.AString(1); got != want {
		t.Errorf("AString(1): got %v, want %v", got, want)
	}
}

func TestCString(t *testing.T) {
	bs := []byte{74, 69, 82, 82, 89, 0, 65, 0}
	dec := NewDefaultDecoder(bs, binary.BigEndian)

	want := "JERRY"
	if got := dec.CString(); got != want {
		t.Errorf("CString: got %v, want %v", got, want)
	}
	want = "A"
	if got := dec.CString(); got != want {
		t.Errorf("CString: got %v, want %v", got, want)
	}
}
