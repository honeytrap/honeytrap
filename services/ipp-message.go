package services

import (
	"bytes"
	"encoding/binary"

	"github.com/honeytrap/honeytrap/services/decoder"
)

const (
	// Delimiter tags , begin-attribute-group-tag
	opAttribTag      byte = 0x01 //operation-attributes-tag
	jobAttribTag     byte = 0x02 //job-attributes-tag
	endAttribTag     byte = 0x03 //end-of-attributes-tag
	printerAttribTag byte = 0x04 //printer-attributes-tag
	unsupAttribTag   byte = 0x05 //unsupported-attributes-tag

	// Value tags
	valInteger       byte = 0x21
	valBoolean       byte = 0x22
	valEnum          byte = 0x23
	valOctetStr      byte = 0x30
	valDateTime      byte = 0x31
	valResolution    byte = 0x32
	valRangeOfInt    byte = 0x33
	begCollection    byte = 0x34
	textWithLang     byte = 0x35
	nameWithLang     byte = 0x36
	endCollection    byte = 0x37
	textWithoutLang  byte = 0x41
	nameWithoutLang  byte = 0x42
	valKeyword       byte = 0x44
	valUri           byte = 0x45
	valUriScheme     byte = 0x46
	valCharSet       byte = 0x47 //attributes-charset
	naturelLang      byte = 0x48 //attributes-naturel-language
	mimeMediaType    byte = 0x49
	memberAttribName byte = 0x4a

	// Operation ids
	opPrintJob         int16 = 0x0002
	opValidateJob      int16 = 0x0004
	opCreateJob        int16 = 0x0005
	opGetJobAttrib     int16 = 0x0009
	opGetPrinterAttrib int16 = 0x000b

	// Status values
	sOk int16 = 0x0000 //successful-ok
)

type ippMessage struct {
	versionMajor byte
	versionMinor byte
	statusCode   int16 //is operation-id in request
	requestId    int32
	attributes   []attribGroup
	endTag       byte   //is always endAttribTag (3)
	data         []byte //if there is data otherwise nil
}

type attribGroup struct {
	attribGroupTag byte //begin-attribute-group-tag
	val            []attribOneValue
}

type attribOneValue struct { //Atrribute-with-one-value
	valueTag byte  //value-tag
	nameLen  int16 //name-length
	name     []byte
	valueLen int16 //value-length
	value    []byte
	aVal     []additionalValue
}

type additionalValue struct { //additional-value
	valueTag byte  //value-tag
	nameLen  int16 //name-length should always be 0x0
	valueLen int16 //value-length
	value    []byte
}

func (ao attribOneValue) decode(dec decoder.Decoder) error {
	var err error

	vtag, err := dec.Byte()
	nlen, err := dec.Int16()
	if nlen == 0 { //Additional value
		vlen, err := dec.Int16()
		v := dec.Copy(int(vlen))
		a := additionalValue{
			valueTag: vtag,
			nameLen:  nlen,
			valueLen: vlen,
			value:    v,
		}
		ao.aVal = append(ao.aVal, a)
		return err
	} else {
		ao.valueTag = vtag
		ao.nameLen = nlen
		ao.name = dec.Copy(int(ao.nameLen))
		ao.valueLen, err = dec.Int16()
		ao.value = dec.Copy(int(ao.valueLen))
	}
	return err
}

// Returns a IPP response based on the IPP request
func ippHandler(ippBody []byte) (*bytes.Buffer, []byte) {
	body := &ippMessage{}

	err := body.Read(ippBody)
	if err != nil {
		return nil, nil
	}

	rbody := &ippMessage{}
	rbody.versionMajor = body.versionMajor
	rbody.versionMinor = body.versionMinor
	rbody.requestId = body.requestId
	rbody.statusCode = sOk //We have the ultimate printer

	print := []byte{}

	switch body.statusCode { //operation-id
	case opPrintJob:
		print = body.data
	case opValidateJob:
	case opCreateJob:
	case opGetJobAttrib:
	case opGetPrinterAttrib:
	default:
	}

	if len(print) > 0 {
		return rbody.Response(), print
	}
	return rbody.Response(), nil
}

func (m *ippMessage) Read(raw []byte) error {
	dec := decoder.NewDefaultDecoder(raw, binary.BigEndian)
	if err := dec.HasBytes(8); err != nil {
		return err
	}
	m.versionMajor, _ = dec.Byte()
	m.versionMinor, _ = dec.Byte()
	m.statusCode, _ = dec.Int16()
	m.requestId, _ = dec.Int32()

	for dtag, err := dec.Byte(); dtag != endAttribTag; dtag, err = dec.Byte() {
		if err != nil {
			return err
		}

		group := attribGroup{}
		group.attribGroupTag = dtag
		for dp, err := dec.PeekByte(); dp > unsupAttribTag; dp, err = dec.PeekByte() { //Not a delimiter tag
			if err != nil {
				return err
			}

			aov := attribOneValue{}
			if err := aov.decode(dec); err != nil {
				return err
			}
			group.val = append(group.val, aov)
		}

		m.attributes = append(m.attributes, group)
	}

	m.endTag = endAttribTag
	m.data = dec.Copy(dec.Available())

	return nil
}

func (m *ippMessage) Response() *bytes.Buffer {
	buf := new(bytes.Buffer)
	var err error
	err = binary.Write(buf, binary.BigEndian, m.versionMajor)
	err = binary.Write(buf, binary.BigEndian, m.versionMinor)
	err = binary.Write(buf, binary.BigEndian, m.statusCode)
	err = binary.Write(buf, binary.BigEndian, m.requestId)
	if m.attributes[0].attribGroupTag == opAttribTag {
		err = binary.Write(buf, binary.BigEndian, opAttribTag)
		//write charset
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[0].valueTag)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[0].nameLen)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[0].name)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[0].valueLen)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[0].value)

		//write language
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[1].valueTag)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[1].nameLen)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[1].name)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[1].valueLen)
		err = binary.Write(buf, binary.BigEndian, m.attributes[0].val[1].value)
	}

	err = binary.Write(buf, binary.BigEndian, endAttribTag)
	if err != nil {
		return nil
	}
	return buf
}
