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
	data         []byte //if there is data otherwise nil
	username     string
}

type attribGroup struct {
	attribGroupTag byte //begin-attribute-group-tag
	val            []attribOneValue
}

type attribOneValue struct { //Atrribute-with-one-value
	valueTag byte //value-tag
	name     string
	value    string
	aVal     []additionalValue
}

type additionalValue struct { //additional-value
	valueTag byte //value-tag
	value    string
}

func (ao *attribOneValue) decode(dec *decoder.DefaultDecoder) error {

	var err error
	vtag, err := dec.Byte()
	if l, _ := dec.Int16(); l == 0 { //This is an Additional value
		s, _ := dec.Data()

		a := additionalValue{
			valueTag: vtag,
			value:    s,
		}
		ao.aVal = append(ao.aVal, a)
		return err

	} else {
		ao.valueTag = vtag
		ao.name, _ = dec.Data()
		ao.value, _ = dec.Data()
	}
	return err
}

func (v *additionalValue) encode(buf *decoder.Encoder) {
	buf.WriteUint8(v.valueTag)
	buf.WriteUint16(int16(0))
	buf.WriteData(v.value)
}

func (v *attribOneValue) encode(buf *decoder.Encoder) {
	buf.WriteUint8(v.valueTag)
	buf.WriteData(v.name)
	buf.WriteData(v.value)
	if v.aVal != nil {
		for _, av := range v.aVal {
			av.encode(buf)
		}
	}
}

func (v *attribGroup) encode(buf *decoder.Encoder) {
	buf.WriteUint8(v.attribGroupTag)
	if v.val != nil {
		for _, aov := range v.val {
			aov.encode(buf)
		}
	}
}

func (v *ippMessage) encode(buf *decoder.Encoder) {
	buf.WriteUint8(v.versionMajor)
	buf.WriteUint8(v.versionMinor)
	buf.WriteUint16(v.statusCode)
	buf.WriteUint32(v.requestId)
	if v.attributes != nil {
		for _, im := range v.attributes {
			im.encode(buf)
		}
	}
}

// Returns a IPP response based on the IPP request
func IPPHandler(ippBody []byte) (*ippMessage, error) {
	body := &ippMessage{}

	err := body.Read(ippBody)
	if err != nil {
		return nil, err
	}

	//Response structure
	rbody := &ippMessage{
		versionMajor: body.versionMajor,
		versionMinor: body.versionMinor,
		statusCode:   sOk,
		requestId:    body.requestId,
	}

	switch body.statusCode { //operation-id
	case opPrintJob:
	case opValidateJob:
	case opCreateJob:
	case opGetJobAttrib:
	case opGetPrinterAttrib:
	default:
	}
	return rbody, nil
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
	// Append required end tag
	m.attributes = append(m.attributes, attribGroup{attribGroupTag: endAttribTag})

	// Copy remaining data (printdata)
	m.data = dec.Copy(dec.Available())

	return nil
}

func (m *ippMessage) Response() *bytes.Buffer {
	buf := &decoder.Encoder{}
	m.encode(buf)
	return &buf.Buffer
}
