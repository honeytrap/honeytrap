/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package ipp

import "github.com/honeytrap/honeytrap/services/decoder"

type ValueType interface {
	Tag() byte
	encode(buf decoder.EncoderType)
	decode(dec decoder.Decoder) error
}

type valInt struct {
	tag  byte
	name string
	val  []int32
}

type valStr struct {
	tag  byte
	name string
	val  []string //also 1setof
}

type valBool struct {
	tag  byte
	name string
	val  []bool
}

type valRangeInt struct {
	tag  byte
	name string
	low  int32
	high int32
}

func (v *valInt) Tag() byte {
	return v.tag
}

func (v *valStr) Tag() byte {
	return v.tag
}

func (v *valBool) Tag() byte {
	return v.tag
}

func (v *valRangeInt) Tag() byte {
	return v.tag
}

func (v *valInt) encode(buf decoder.EncoderType) {
	z := false //Zero length string

	for _, value := range v.val {
		buf.WriteUint8(v.tag)
		buf.WriteData(v.name, z)

		buf.WriteUint16(4) //length of value
		buf.WriteUint32(value)
		z = true
	}
}

func (v *valInt) decode(dec decoder.Decoder) error {

	v.name = dec.Data()
	//Read value length field away, is always 4
	_ = dec.Int16()
	v.val = append(v.val, dec.Int32())

	// Check for additional values
	if vtag := dec.Byte(); vtag == v.tag {
		//check name length
		if l := dec.Int16(); l == 0 {
			_ = dec.Int16()
			v.val = append(v.val, dec.Int32())
		} else {
			//rewind buffer
			dec.Seek(-3)
		}
	} else {
		//rewind buffer
		dec.Seek(-1)
	}
	if err := dec.LastError(); err != nil {
		return err
	}

	return nil
}

func (v *valStr) encode(buf decoder.EncoderType) {
	z := false //zero length string

	for _, val := range v.val {
		buf.WriteUint8(v.tag)
		buf.WriteData(v.name, z)
		buf.WriteData(val, false)
		z = true
	}
}

func (v *valStr) decode(dec decoder.Decoder) error {

	v.name = dec.Data()
	v.val = append(v.val, dec.Data())

	// Check for additional values
	vtag := dec.Byte()
	for vtag == v.tag {
		//check name length
		if l := dec.Int16(); l == 0 {
			v.val = append(v.val, dec.Data())
			vtag = dec.Byte()
		} else {
			dec.Seek(-2) //Rewind name length
			break
		}
	}
	dec.Seek(-1) //Rewind tag

	return dec.LastError()
}

func (v *valBool) encode(buf decoder.EncoderType) {
	z := false //zero length string

	for _, val := range v.val {
		buf.WriteUint8(v.tag)
		buf.WriteData(v.name, z)
		buf.WriteUint16(1) //lenght of bool
		if val {
			buf.WriteUint8(1)
		} else {
			buf.WriteUint8(0)
		}
		z = true
	}
}

func (v *valBool) decode(dec decoder.Decoder) error {
	v.name = dec.Data()
	_ = dec.Byte() // len is 1, read away

	if b := dec.Byte(); b == 1 {
		v.val = append(v.val, true)
	} else {
		v.val = append(v.val, false)
	}

	// Check for additional values
	vtag := dec.Byte()
	for vtag != v.tag {
		//check name length
		if l := dec.Int16(); l == 0 {
			if b := dec.Byte(); b == 1 {
				v.val = append(v.val, true)
			} else {
				v.val = append(v.val, false)
			}
			vtag = dec.Byte()
		} else {
			dec.Seek(-2) //Rewind name length
			break
		}
	}
	dec.Seek(-1) //Rewind tag

	return dec.LastError()
}

func (v *valRangeInt) encode(buf decoder.EncoderType) {
	buf.WriteUint8(v.tag)
	buf.WriteData(v.name, false)

	buf.WriteUint16(8) //value length

	buf.WriteUint32(v.low)
	buf.WriteUint32(v.high)
}

func (v *valRangeInt) decode(dec decoder.Decoder) error {
	v.name = dec.Data()
	_ = dec.Int16() //Read away, len is always 8

	v.low = dec.Int32()
	v.high = dec.Int32()

	return dec.LastError()
}
