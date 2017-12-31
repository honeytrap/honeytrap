package decoder

import (
	"testing"
)

func TestHasBytes(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

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
	dec := NewDecoder(bs)

	if b := dec.PeekByte(); b != 0xff {
		t.Errorf("PeekByte: expected 0xff, got %d", b)
	}
	if b := dec.Byte(); b != 0xff { // Read the fist byte again
		t.Errorf("Byte: expected 0xff, got %d", b)
	}
}

func TestInt16(t *testing.T) {
	bs := []byte{0, 0x01, 0, 0x02, 0, 0x03}
	dec := NewDecoder(bs)

	if b := dec.Int16(); b != 1 {
		t.Errorf("int16: expected 1, got %d", b)
	}
}

func TestInt32(t *testing.T) {
	bs := []byte{0, 0, 0, 0x01, 0, 0, 0, 0x02}
	dec := NewDecoder(bs)

	if b := dec.Int32(); b != 1 {
		t.Errorf("expect 1, got %v", b)
	}
}

func TestNew(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	if dec.offset != 0 {
		t.Errorf("NewDecoder: offset: %d, expected 0!", dec.offset)
	}

	if avail := dec.Available(); avail != len(bs) {
		t.Errorf("Available gives wrong size got %d, expected %d", avail, len(bs))
	}

}

func TestSeek(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	dec.Seek(2)
	want := byte(0x03)
	if got := dec.PeekByte(); got != want {
		t.Errorf("Seek(2): got %d, expected %d", got, want)
	}

	dec.Seek(len(bs)) //Seek past the end
	if err := dec.LastError(); err == nil {
		t.Errorf("Seek(past the end): No error return! Decoder: %v", dec)
	}

	dec.Seek(-100) //Seek before start
	if err := dec.LastError(); err == nil {
		t.Errorf("Seek(before start): No error return! Decoder: %v", dec)
	}
}

func TestCopyAll(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	c := dec.Copy(dec.Available())
	if len(c) != len(bs) {
		t.Errorf("Copy: len copy: %v != len orig: %v", len(c), len(bs))
	}
}

func TestCopy2(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	sz := 2
	c := dec.Copy(sz)
	if len(c) != sz {
		t.Errorf("Copy: len copy: %v != len orig: %v", len(c), sz)
	}
	if c[1] != 0x02 {
		t.Errorf("Copy2: copied wrong bytes, got %v want 2", c[1])
	}
}

func TestCopyTooMuch(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	c := dec.Copy(len(bs) + 1)
	if c != nil {
		t.Errorf("Copy: is not nil after asking for too much bytes")
	}
}

func TestCopyOffset(t *testing.T) {
	bs := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	dec := NewDecoder(bs)

	_ = dec.Copy(2)
	cc := dec.Copy(2)
	if cc[0] != 0x03 {
		t.Errorf("CopyOffset: Offset wrong with second copy. Got %v want 0x03", cc[0])
	}
}
