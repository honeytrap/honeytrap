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

func NewDefaultDecoder(data []byte, byteOrder binary.ByteOrder) *DefaultDecoder {
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

func (d *DefaultDecoder) Data() (string, error) {
	l, err := d.Int16()
	return string(d.Copy(int(l))), err
}

func (d *DefaultDecoder) PeekByte() (byte, error) {
	if err := d.HasBytes(1); err != nil {
		return 0, err
	}

	return d.data[d.offset], nil
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
