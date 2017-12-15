package decoder

import (
	"encoding/binary"
	"fmt"
	"math"
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
	HasBytes(size int) error

	Byte() (byte, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)

	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)

	IEEE754_Float32() (float32, error)
	IEEE754_Float64() (float64, error)

	PeekByte() (byte, error)

	AString(size int) (string, error)
	CString() string
	Data() []byte
	Copy(size int) []byte

	Skip(int) error
	Seek(int) (int, error)
	Offset() int
	StartOffset() int
	Available() int

	Dump()

	ByteOrder() binary.ByteOrder
	SetByteOrder(byteOrder binary.ByteOrder) binary.ByteOrder
}

func NewDefaultDecoder(data []byte, byteOrder binary.ByteOrder) Decoder {
	return &DefaultDecoder{
		offset:      0,
		startOffset: 0,
		byteOrder:   byteOrder,
		data:        data,
	}
}

type DefaultDecoder struct {
	offset      int
	startOffset int
	byteOrder   binary.ByteOrder
	data        []byte
}

func (d *DefaultDecoder) SetByteOrder(byteOrder binary.ByteOrder) binary.ByteOrder {
	prevByteOrder := d.byteOrder
	d.byteOrder = byteOrder
	return prevByteOrder
}

func (d *DefaultDecoder) ByteOrder() binary.ByteOrder {
	return d.byteOrder

}

func (d *DefaultDecoder) Offset() int {
	return d.offset

}

func (d *DefaultDecoder) StartOffset() int {
	return d.startOffset

}

func (d *DefaultDecoder) Available() int {
	return len(d.data) - d.offset

}

/* Check if size bytes are available
 * Zero or negative sizes are not allowed
 * nb. Use Seek if you wanna go back
 */
func (d *DefaultDecoder) HasBytes(size int) error {
	if size > 0 && len(d.data) >= d.offset+size {
		return nil
	}

	return ErrOutOfBounds{
		Min: 0,
		Max: len(d.data),
		Got: size,
	}
}

func (d *DefaultDecoder) Byte() (byte, error) {
	if err := d.HasBytes(1); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 1
	}()

	return d.data[d.offset], nil
}

func (d *DefaultDecoder) Int16() (int16, error) {
	if err := d.HasBytes(2); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 2
	}()

	return int16(d.byteOrder.Uint16(d.data[d.offset : d.offset+2])), nil
}

func (d *DefaultDecoder) Int32() (int32, error) {
	if err := d.HasBytes(4); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 4
	}()

	return int32(d.byteOrder.Uint32(d.data[d.offset : d.offset+4])), nil
}

func (d *DefaultDecoder) Int64() (int64, error) {
	if err := d.HasBytes(8); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 8
	}()

	return int64(d.byteOrder.Uint64(d.data[d.offset : d.offset+8])), nil
}

func (d *DefaultDecoder) Uint8() (uint8, error) {
	if err := d.HasBytes(1); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 1
	}()

	return uint8(d.data[d.offset]), nil
}

func (d *DefaultDecoder) Uint16() (uint16, error) {
	if err := d.HasBytes(2); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 2
	}()

	return d.byteOrder.Uint16(d.data[d.offset : d.offset+2]), nil
}

func (d *DefaultDecoder) Uint32() (uint32, error) {
	if err := d.HasBytes(4); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 4
	}()

	return d.byteOrder.Uint32(d.data[d.offset : d.offset+4]), nil
}

func (d *DefaultDecoder) Uint64() (uint64, error) {
	if err := d.HasBytes(8); err != nil {
		return 0, err
	}

	defer func() {
		d.offset += 8
	}()

	return d.byteOrder.Uint64(d.data[d.offset : d.offset+8]), nil
}

func (d *DefaultDecoder) PeekByte() (byte, error) {
	if err := d.HasBytes(1); err != nil {
		return 0, err
	}

	return d.data[d.offset], nil
}

func (d *DefaultDecoder) IEEE754_Float32() (float32, error) {
	if v, err := d.Uint32(); err != nil {
		return 0, err
	} else {
		return math.Float32frombits(v), nil
	}
}

func (d *DefaultDecoder) IEEE754_Float64() (float64, error) {
	if v, err := d.Uint64(); err != nil {
		return 0, err
	} else {
		return math.Float64frombits(v), nil
	}
}

/* Copy size bytes from offset to new byte slice
 * Returns nil if size is 0 or larger then data size.
 * offset is set accordingly
 */
func (d *DefaultDecoder) Copy(size int) []byte {
	if size < 0 || d.offset+size > len(d.data) {
		return nil
	}
	c := make([]byte, size)
	copy(c, d.data[d.offset:d.offset+size])
	d.offset += size
	return c
}

// Read a byte string of given size
func (d *DefaultDecoder) AString(size int) (string, error) {
	if err := d.HasBytes(size); err != nil {
		return "", err
	}

	defer func() {
		d.offset += size
	}()

	return string(d.data[d.offset : d.offset+size]), nil
}

/* Read a null terminated byte string or till end of data
 * This also reads the terminating zero
 */
func (d *DefaultDecoder) CString() string {
	size := 0
	for d.data[d.offset+size] != 0x00 {
		size++
		if d.offset+size == len(d.data) {
			break
		}
	}
	// First char read is a zero, return an empty string
	if size == 0 {
		d.offset++
		return ""
	}

	defer func() {
		// Step over terminating zero
		d.offset += size + 1
	}()

	return string(d.data[d.offset : d.offset+size])
}

/* Seeks to absolute position,
 * Error if n gets out of bounds
 *   offset remains unchanged
 */
func (d *DefaultDecoder) Seek(n int) (int, error) {
	prev := d.offset

	if n < 0 || n >= len(d.data) {
		return prev, ErrOutOfBounds{
			Min: d.startOffset,
			Max: len(d.data) - 1,
			Got: n,
		}
	}
	d.offset = n
	return prev, nil
}

/* Skip to relative position,
 * Error if offset gets out of bounds
 *   offset remains unchanged
 */
func (d *DefaultDecoder) Skip(n int) error {
	if d.offset+n >= len(d.data) || d.offset+n < 0 {
		return ErrOutOfBounds{
			Min: d.startOffset,
			Max: len(d.data) - 1,
			Got: n,
		}
	}

	d.offset += n
	return nil
}

func (d *DefaultDecoder) Data() []byte {
	return d.data[:]
}

func (d *DefaultDecoder) Dump() {

	fmt.Printf("Offset: %d\n% #x \n", d.offset, d.data[d.offset:])
}
