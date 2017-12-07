package services

import (
	"bytes"
	"errors"
)

const (
	// Delimiter tag values
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
	valCharSet       byte = 0x47
	naturelLang      byte = 0x48
	mimeMediaType    byte = 0x49
	memberAttribName byte = 0x4a

	// Operation ids
	opPrintJob         int16 = 0x0002
	opValidateJob      int16 = 0x0004
	opCreateJob        int16 = 0x0005
	opGetJobAttrib     int16 = 0x0009
	opGetPrinterAttrib int16 = 0x000b
)

type ippMessage struct {
	versionMajor int8
	versionMinor int8
	statusCode   int16 // is operation-id in request
	requestId    int32
	attributes   []attribGroup
	endTag       byte   // is always endAttribTag (3)
	data         []byte // if there is data otherwise nil
}

type attribGroup struct {
	attribGroupTag int8 //begin-attribute-tag
	val            attribOneValue
	additionalVal  []additionalValue
}

type attribOneValue struct { //Atrribute-with-one-value
	valueTag int8  //value-tag
	nameLen  int16 //name-length
	name     []byte
	valueLen int16 //value-length
	value    []byte
}

type additionalValue struct { //additional-value
	valueTag int8  //value-tag
	nameLen  int16 //name-length should always be 0x0
	valueLen int16 //value-length
	value    []byte
}

// Returns a IPP response based on the IPP request
func ippHandler(ippBody []byte) *bytes.Buffer {
	msg := &ippMessage{}
	if err := msg.Read(ippBody); err != nil {
		return err
	}
}

func (msg *ippMessage) Read(buf *bytes.Buffer) error {

	if msg.versionMajor, err = buf.ReadByte(); err != nil {
		return err
	}
	if msg.versionMinor, err = buf.ReadByte(); err != nil {
		return err
	}
	if buf.Len() == 0 {
		return errors.New("EOF")
	}
	msg.statusCode = int16(buf.Next(2))
	if buf.Len() == 0 {
		return errors.New("EOF")
	}
	msg.requestId = int32(buf.Next(4))
	if buf.Len() == 0 {
		return errors.New("EOF")
	}

	// Check end-of-attribute-tag
	if buf.ReadByte() != 0x03 {
		return errors.New("IPP Request malformed")
	}
	// Read rest of the data if any
	if buf.Len() > 0 {
		msg.data = buf.Next(buf.Len())
	}
	return nil
}
