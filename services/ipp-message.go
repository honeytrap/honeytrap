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

	data     []byte //if there is data otherwise nil
	username string
	uri      string
	format   string
}

func (m *ippMsg) read(raw []byte) error {
	log.Debug("START ippMsg.read([]byte)")
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
func (v *ippMsg) encode() *bytes.Buffer {
	log.Debug("START ippMsg.encode()")
	buf := decoder.NewEncoder()

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
	return &buf.Buffer
}

func (r *ippMsg) setOpAttribResponse(gin *attribGroup) {
	log.Debug("START ippMsg.setOpAttribResponse(*attribGroup)")
	grp := &attribGroup{tag: gin.tag}

	for _, v := range gin.val {
		switch v.Tag() {
		case valCharSet:
			grp.val = append(grp.val, v)
		case naturelLang:
			grp.val = append(grp.val, v)
		}
	}
	r.attributes = append(r.attributes, grp)
}

func (r *ippMsg) setGetPrinterResponse() {
	log.Debug("START ippMsg.setGetPrinterResponse()")
	//Append a printer profile
	r.attributes = append(r.attributes, model)
}

func (r *ippMsg) setPrintJobResponse(b *ippMsg) {
	log.Debug("START ippMsg.setPrintJobResponse(*ippMsg)")
	for _, g := range b.attributes {
		if g.tag == opAttribTag {
			for _, val := range g.val {
				v, _ := val.(*valStr)
				if v.name == "printer-uri" {
					r.uri = v.val[0]
				} else if v.name == "requesting-user-name" {
					r.uri = v.val[0]
				} else if v.name == "document-format" {
					r.format = v.val[0]
				}
			}
			break
		}
	}
	r.data = b.data
}

// Returns a IPP response based on the IPP request
func IPPHandler(ippBody []byte) (*ippMsg, error) {
	log.Debug("IPP handler started")

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

	//Set operation attributes
	log.Debug("SET operation attributes ID:", body.requestId)
	for _, g := range body.attributes {
		if g.tag == opAttribTag {
			rbody.setOpAttribResponse(g)
			break
		}
	}

	switch body.statusCode { //operation-id
	case opGetPrinterAttrib:
		log.Debug("Get Printer Attributes")
		rbody.setGetPrinterResponse()
	case opPrintJob:
		log.Debug("Print Job")
		rbody.setPrintJobResponse(body)
	case opValidateJob:
		log.Debug("Validate Job")
	case opGetJobAttrib:
		log.Debug("Get Job Attributes")
	}

	//Set end tag
	rbody.attributes = append(rbody.attributes, &attribGroup{tag: endAttribTag})

	return rbody, nil
}
