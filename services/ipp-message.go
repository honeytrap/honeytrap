package services

import (
	"bytes"

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

// Wraps a complete ipp object
type ippMsg struct {
	versionMajor byte
	versionMinor byte
	statusCode   int16 //is operation-id in request
	requestId    int32
	attributes   []*attribGroup
	data         []byte //if there is data otherwise nil
	username     string
}

type attribGroup struct {
	tag byte //begin-attribute-group-tag
	val []ippValueType
}

func (ag *attribGroup) decode(dec decoder.Decoder) error {

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
		}

		v.decode(dec)
		ag.val = append(ag.val, v)
	}

	// Put back last read byte, because it is a delimiter tag
	dec.Seek(-1)

	return nil
}

func (ag *attribGroup) encode(buf decoder.EncoderType) error {
	buf.WriteUint8(ag.tag)

	if ag.val != nil {
		for _, vals := range ag.val {
			vals.encode(buf)
		}
	}
	return nil
}

func (m *ippMsg) read(raw []byte) error {
	dec := decoder.NewDecoder(raw)

	m.versionMajor = dec.Byte()
	m.versionMinor = dec.Byte()
	m.statusCode = dec.Int16()
	m.requestId = dec.Int32()

	// Groups, dtag is a delimiter(group) tag
	for dtag := dec.Byte(); dtag != endAttribTag; dtag = dec.Byte() {

		group := &attribGroup{tag: dtag}
		if err := group.decode(dec); err != nil {
			return err
		}

		m.attributes = append(m.attributes, group)
	}

	// Append required endtag
	m.attributes = append(m.attributes, &attribGroup{tag: endAttribTag})

	// Copy remaining data (printdata)
	m.data = dec.Copy(dec.Available())

	return nil
}

// Encodes the ipp response message suitable for http transport
func (m *ippMsg) encode() *bytes.Buffer {
	buf := decoder.NewEncoder()

	m.encode(buf)
	return &buf.Buffer
}

func (v *ippMsg) encode(buf decoder.EncoderType) {

	// Header
	buf.WriteUint8(v.versionMajor)
	buf.WriteUint8(v.versionMinor)
	buf.WriteUint16(v.statusCode)
	buf.WriteUint32(v.requestId)

	if v.attributes != nil {
		for _, group := range v.attributes {
			group.encode(buf)
		}
	}
}

func (gout *ippMsg) setGroupResponse(gin *attribGroup) {
	grp := &attribGroup{attribGroupTag: gin.tag}

	switch gin.tag {
	case opAttribTag:
		for _, v := range gin.val {
			switch v.Tag() {
			case valCharSet:
				grp.val = append(grp.val, v)
			case naturelLang:
				grp.val = append(grp.val, v)
			case nameWithoutLang:
				if v.name == "requesting-user-name" {
					gout.username = v.value
				}
			}
		}
	case printerAttribTag:
	case jobAttribTag:
	case endAttribTag:
	}

	gout.attributes = append(gout.attributes, grp)
}

// Returns a IPP response based on the IPP request
func IPPHandler(ippBody []byte) (*ippMsg, error) {
	body := &ippMsg{}

	err := body.read(ippBody)
	if err != nil {
		return nil, err
	}

	//Response structure
	rbody := &ippMsg{
		versionMajor: body.versionMajor,
		versionMinor: body.versionMinor,
		statusCode:   sOk,
		requestId:    body.requestId,
		data:         body.data,
	}

	switch body.statusCode { //operation-id
	case opPrintJob:
		for _, g := range body.attributes {
			rbody.setGroupResponse(g)
		}
	case opValidateJob:
	case opGetJobAttrib:
	case opGetPrinterAttrib:
	}

	return rbody, nil
}
