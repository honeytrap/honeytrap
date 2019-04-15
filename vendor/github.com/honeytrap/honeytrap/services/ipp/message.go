// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ipp

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
	valURI           byte = 0x45
	valURIScheme     byte = 0x46
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
	opCupsGetDevices   int16 = 0x400b

	// Status values
	sOk int16 = 0x0000 //successful-ok
)

type ippMsg struct {

	//IPP message, request and response
	versionMajor byte
	versionMinor byte
	statusCode   int16 //is operation-id in request
	requestID    int32
	attributes   []*attribGroup

	//Extra, for our own use
	data     []byte //if there is data otherwise nil
	username string
	uri      string
	format   string
	jobname  string
}

func (m *ippMsg) decode(raw []byte) error {
	dec := decoder.NewDecoder(raw)

	m.versionMajor = dec.Byte()
	m.versionMinor = dec.Byte()
	m.statusCode = dec.Int16()
	m.requestID = dec.Int32()

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

	return dec.LastError()
}

// Encodes the ipp response message suitable for http transport
func (m *ippMsg) encode() *bytes.Buffer {
	buf := decoder.NewEncoder()

	// Header
	buf.WriteUint8(m.versionMajor)
	buf.WriteUint8(m.versionMinor)
	buf.WriteUint16(m.statusCode)
	buf.WriteUint32(m.requestID)

	if m.attributes != nil {
		for _, group := range m.attributes {
			group.encode(buf)
		}
	}
	return &buf.Buffer
}

func (m *ippMsg) setOpAttribResponse(gin *attribGroup) {
	grp := &attribGroup{tag: gin.tag}

	for _, v := range gin.val {
		switch v.Tag() {
		case valCharSet:
			grp.val = append(grp.val, v)
		case naturelLang:
			grp.val = append(grp.val, v)
		}
	}
	m.attributes = append(m.attributes, grp)
}

func (m *ippMsg) setGetPrinterResponse() {
	//Append a printer profile
	m.attributes = append(m.attributes, model)
}

func (m *ippMsg) setPrintJobResponse(b *ippMsg) {

	for _, g := range b.attributes {
		if g.tag == opAttribTag {
			for _, val := range g.val {
				v, _ := val.(*valStr)
				if v.name == "printer-uri" {
					m.uri = v.val[0]
				} else if v.name == "requesting-user-name" {
					m.username = v.val[0]
				} else if v.name == "document-format" {
					m.format = v.val[0]
				} else if v.name == "job-name" {
					m.jobname = v.val[0]
				}
			}
			break
		}
	}
	m.data = b.data
}

func (m *ippMsg) setGetDevices() {
	grp := &attribGroup{tag: printerAttribTag}
	m.attributes = append(m.attributes, grp)
}

// Returns a IPP response based on the IPP request
func ippHandler(ippBody []byte) (*ippMsg, error) {
	body := &ippMsg{}

	err := body.decode(ippBody)
	if err != nil {
		return nil, err
	}

	//Response structure
	rbody := &ippMsg{
		versionMajor: body.versionMajor,
		versionMinor: body.versionMinor,
		statusCode:   sOk,
		requestID:    body.requestID,
		data:         body.data,
	}

	//Set operation attributes
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
		log.Debug("IPP: Get Job Attributes")
	case opCupsGetDevices:
		log.Debug("IPP: CUPS Get Devices")
		rbody.setGetDevices()
	}

	//Set end tag
	rbody.attributes = append(rbody.attributes, &attribGroup{tag: endAttribTag})

	return rbody, nil
}
