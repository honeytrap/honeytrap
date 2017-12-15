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
	val            attribOneValue
	additionalVal  []additionalValue
}

type attribOneValue struct { //Atrribute-with-one-value
	valueTag byte  //value-tag
	nameLen  int16 //name-length
	name     []byte
	valueLen int16 //value-length
	value    []byte
}

type additionalValue struct { //additional-value
	valueTag byte  //value-tag
	nameLen  int16 //name-length should always be 0x0
	valueLen int16 //value-length
	value    []byte
}

// Returns a IPP response based on the IPP request
func ippHandler(ippBody []byte) (*bytes.Buffer, []byte) {
	body := &ippMessage{}

	err := body.Read(ippBody)
	if err != nil {
		return nil
	}

	rbody := &ippMessage{}
	rbody.versionMajor = body.versionMajor
	rbody.versionMinor = body.versionMinor
	rbody.requestId = body.requestId
	rbody.statusCode = sOk //We have the ultimate printer

	var print []byte //nil slice, the print data if any

	switch body.statusCode { //operation-id
	case opPrintJob:
		print := body.data
	case opValidateJob:
	case opCreateJob:
	case opGetJobAttrib:
	case opGetPrinterAttrib:
	default:
	}

	return rbody.Response(), print
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

	//Find the end tag
	for end, err := dec.Byte(); end == endAttribTag; end, err = dec.Byte() {
		if err != nil {
			break
		}
	}
	m.data = dec.Copy(dec.Available())

	return nil
}

func (m *ippMessage) Response() *bytes.Buffer {
	buf := new(bytes.Buffer)
	_ := binary.Write(buf, binary.BigEndian, m.versionMajor)
	_ := binary.Write(buf, binary.BigEndian, m.versionMinor)
	_ := binary.Write(buf, binary.BigEndian, m.statusCode)
	_ := binary.Write(buf, binary.BigEndian, m.requestId)

	return buf
}
