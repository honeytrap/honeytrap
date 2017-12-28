package decoder

import (
	"encoding/binary"
	"fmt"
)

type ErrOutOfBounds struct {
	Min int
	Max int
	Got int
}

func (e ErrOutOfBounds) Error() string {
	return fmt.Sprintf("Index out of bounds! min: %v, max: %v got: %v", e.Min, e.Max, e.Got)
}

type Decoder interface {
	Available() int
	HasBytes(size int) error

	Byte() byte
	Int16() int16
	Int32() int32

	PeekByte() byte
	PeekInt16() int16

	Data() string
	Copy(size int) []byte
	Seek(pos int)

	LastError() error
}

type Decode struct {
	offset    int
	data      []byte
	lasterror error
}

func NewDecoder(data []byte) *Decode {
	return &Decode{
		offset: 0,
		data:   data,
	}
}

func (d *Decode) LastError() error {
	return d.lasterror

}

func (d *Decode) Available() int {
	return len(d.data) - d.offset

}

func (d *Decode) HasBytes(size int) error {
	if size > 0 && len(d.data) >= d.offset+size {
		return nil
	}

	return ErrOutOfBounds{
		Min: 0,
		Max: len(d.data),
		Got: size,
	}
}

func (d *Decode) Byte() byte {
	if err := d.HasBytes(1); err != nil {
		d.lasterror = err
		return 0
	}

	defer func() {
		d.offset += 1
	}()

	return d.data[d.offset]
}

func (d *Decode) Int16() int16 {
	if err := d.HasBytes(2); err != nil {
		d.lasterror = err
		return 0
	}

	defer func() {
		d.offset += 2
	}()

	return int16(binary.BigEndian.Uint16(d.data[d.offset : d.offset+2]))
}

func (d *Decode) Int32() int32 {
	if err := d.HasBytes(4); err != nil {
		d.lasterror = err
		return 0
	}

	defer func() {
		d.offset += 4
	}()

	return int32(binary.BigEndian.Uint32(d.data[d.offset : d.offset+4]))
}

func (d *Decode) Data() string {
	l := d.Int16()
	return string(d.Copy(int(l)))
}

func (d *Decode) PeekByte() byte {
	if err := d.HasBytes(1); err != nil {
		d.lasterror = err
		return 0
	}

	return d.data[d.offset]
}

func (d *Decode) PeekInt16() int16 {
	if err := d.HasBytes(2); err != nil {
		d.lasterror = err
		return 0
	}

	return int16(binary.BigEndian.Uint16(d.data[d.offset : d.offset+2]))
}

func (d *Decode) Copy(size int) []byte {
	if err := d.HasBytes(size); err != nil {
		d.lasterror = err
		return nil
	}

	c := make([]byte, size)
	copy(c, d.data[d.offset:d.offset+size])
	d.offset += size
	return c
}

// Seeking relative to current offset
func (d *Decode) Seek(pos int) {
	d.offset += pos
	if d.offset < 0 || d.offset > len(d.data) {
		d.lasterror = ErrOutOfBounds{
			Min: 0,
			Max: len(d.data),
			Got: d.offset,
		}
	}
}
