package services

import "github.com/honeytrap/honeytrap/services/decoder"

type ippValueType interface {
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
	log.Debug("START valInt.encode(decoder.EncoderType)")
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
	log.Debug("START valInt.decode(decoder.Decoder)")

	v.name = dec.Data()
	//Read value lenght field away, is always 4
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
	log.Debug("START valStr.encode(decoder.EncoderType)")
	z := false //zero length string

	for _, val := range v.val {
		buf.WriteUint8(v.tag)
		buf.WriteData(v.name, z)
		buf.WriteData(val, false)
		z = true
	}
}

func (v *valStr) decode(dec decoder.Decoder) error {
	log.Debug("START valStr.decode(decoder.Decoder)")

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

	if err := dec.LastError(); err != nil {
		return err
	}

	return nil
}

func (v *valBool) encode(buf decoder.EncoderType) {
	log.Debug("START valBool.encode(decoder.EncoderType)")
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
	log.Debug("START valBool.decode(decoder.Decoder)")
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

	if err := dec.LastError(); err != nil {
		return err
	}

	return nil
}

func (v *valRangeInt) encode(buf decoder.EncoderType) {
	log.Debug("START valRangeInt.encode(decoder.EncoderType)")
	buf.WriteUint8(v.tag)
	buf.WriteData(v.name, false)

	buf.WriteUint16(8) //value length

	buf.WriteUint32(v.low)
	buf.WriteUint32(v.high)
}

func (v *valRangeInt) decode(dec decoder.Decoder) error {
	log.Debug("START valangeInt.decode(decoderr.Decoder)")
	v.name = dec.Data()
	_ = dec.Int16() //Read away, len is always 8

	v.low = dec.Int32()
	v.high = dec.Int32()

	if err := dec.LastError(); err != nil {
		return err
	}

	return nil
}

type attribGroup struct {
	tag byte //begin-attribute-group-tag
	val []ippValueType
}

func (ag *attribGroup) decode(dec decoder.Decoder) error {
	log.Debug("START attribGroup.decode(decoder.Decoder)")

	for vtag := dec.Byte(); vtag > unsupAttribTag; vtag = dec.Byte() { //Not a delimiter tag
		if err := dec.LastError(); err != nil {
			return err
		}

		var v ippValueType

		switch vtag {
		case valInteger:
			v = &valInt{tag: vtag}
		case valBoolean:
			v = &valBool{tag: vtag}
		case valKeyword:
			v = &valStr{tag: vtag}
		case valCharSet:
			v = &valStr{tag: vtag}
		case valUri:
			v = &valStr{tag: vtag}
		case valRangeOfInt:
			v = &valInt{tag: vtag}
		case naturelLang:
			v = &valStr{tag: vtag}
		case mimeMediaType:
			v = &valStr{tag: vtag}
		case textWithoutLang:
			v = &valStr{tag: vtag}
		case valEnum:
			v = &valInt{tag: vtag}
		case nameWithoutLang:
			v = &valStr{tag: vtag}
		}

		v.decode(dec)
		ag.val = append(ag.val, v)
	}

	// Put back last read byte, because it is a delimiter tag
	dec.Seek(-1)

	return nil
}

func (ag *attribGroup) encode(buf decoder.EncoderType) error {
	log.Debug("START attribGroup.encode(decoder.EncoderType)")
	buf.WriteUint8(ag.tag)

	if ag.val != nil {
		for _, vals := range ag.val {
			vals.encode(buf)
		}
	}
	return nil
}

var model *attribGroup = &attribGroup{
	tag: printerAttribTag,
	val: []ippValueType{
		&valStr{valKeyword, "compression-supported", []string{"none"}},
		&valRangeInt{valRangeOfInt, "copies-supported", int32(1), int32(1)},
		&valStr{mimeMediaType, "document-format-supported", []string{"application/octet-stream", "image/pwg-raster"}},
		&valStr{nameWithoutLang, "marker-colors", []string{"black", "cyan", "magenta", "yellow"}},
		&valInt{valInteger, "marker-high-levels", []int32{100, 100, 100, 100}},
		&valInt{valInteger, "marker-levels", []int32{80, 100, 100, 100}},
		&valInt{valInteger, "marker-low-levels", []int32{10, 10, 10, 10}},
		&valStr{valKeyword, "media-cols-supported", []string{"media-type", "media-size"}},
		&valInt{valEnum, "operations-supported", []int32{2, 4, 11}},
		&valStr{valKeyword, "print-color-mode-supported", []string{"auto", "color", "monochrome"}},
		&valBool{valBoolean, "printer-is-accepting-jobs", []bool{true}},
		&valInt{valEnum, "printer-state", []int32{3}},
		&valStr{valKeyword, "printer-state-reasons", []string{"none"}},
	},
}
