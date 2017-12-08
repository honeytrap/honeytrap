package services

import (
	"bytes"
)

const (
	// Delimiter tag values / begin-attribute-group-tag
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
	attribGroupTag int8 //begin-attribute-group-tag
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
func ippHandler(ippBody *[]byte) *bytes.Buffer {
	body := &ippMessage{
		versionMajor: ippBody[0],
		versionMinor: ippBody[1],
		statusCode:   0, //Ok code
		requestId:    ippBody[4:7],
		endTag:       endAttribTag,
	}
	opId := int16(ippBody[2:3])
	switch opId {
	case opPrintJob:
	case opValidateJob:
	case opCreateJob:
	case opGetJobAttrib:
	case opGetPrinterAttrib:
	default:
	}
}
