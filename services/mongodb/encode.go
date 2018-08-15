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
package mongodb

import (
	"bytes"
	"encoding/binary"
	"time"
)

func (s *mongodbService) encodeResponse(md *MsgData, msg interface{}) []byte {
	switch md.mh.OpCode {
	case 2004:
		return s.encodeOpReplyMsg(md, msg)
	case 2010:
		return s.encodeOpCommandReplyMsg(md, msg)
	case 2013:
		return s.encodeOpMsgMsg(md, msg)
	}
	log.Error("Error when encoding response: opCode unknown: %d", md.mh.OpCode)
	return []byte("") //TODO
}

func (s *mongodbService) encodeOpReplyMsg(md *MsgData, msg interface{}) []byte {
	orm := &OpReplyMsg{
		responseFlags:  8,
		cursorID:       0,
		startingFrom:   0,
		numberReturned: 1}

	doc := document{}
	doc.elements = msg.([]elem)
	orm.documents = []document{doc}

	aw := bytes.NewBuffer([]byte(""))
	encodeDocuments(aw, orm.documents)
	str := aw.String()
	aw.Reset()
	aw.Write(encodeIntLittleEndian(orm.responseFlags, 4))
	aw.Write(encodeIntLittleEndian(orm.cursorID, 8))
	aw.Write(encodeIntLittleEndian(orm.startingFrom, 4))
	aw.Write(encodeIntLittleEndian(orm.numberReturned, 4))
	aw.WriteString(str)
	return s.encodeHeader(md, aw, orm)
}

func (s *mongodbService) encodeOpCommandReplyMsg(md *MsgData, msg interface{}) []byte {
	ocrm := &OpCommandReplyMsg{}
	ocrm.metadata.elements = msg.([]elem)
	ocrm.commandReply = document{}
	aw := bytes.NewBuffer([]byte(""))
	aw.Write(encodeDoc(ocrm.metadata))
	aw.Write(encodeDoc(ocrm.commandReply))
	return s.encodeHeader(md, aw, ocrm)
}

func (s *mongodbService) encodeOpMsgMsg(md *MsgData, msg interface{}) []byte {
	omm := &OpMsgMsg{}
	omm.flagBits = 0
	sections := document{}
	sections.elements = msg.([]elem)
	aw := bytes.NewBuffer([]byte(""))
	aw.Write(encodeDoc(sections))
	str := aw.String()
	aw.Reset()
	aw.Write(encodeIntLittleEndian(omm.flagBits, 4))
	aw.WriteByte(0x00)
	aw.WriteString(str)
	return s.encodeHeader(md, aw, omm)
}

func encodelocalTime() int64 {
	t := time.Now().UnixNano() / 1000000
	return t
}

func encodeIntLittleEndian(t interface{}, i int) []byte {
	src := make([]byte, i)
	switch i {
	case 4:
		binary.LittleEndian.PutUint32(src, uint32(t.(int32)))
	case 8:
		binary.LittleEndian.PutUint64(src, uint64(t.(int64)))
	}
	return src
}

func encodeIntBigEndian(t interface{}, i int) []byte {
	src := make([]byte, i)
	switch i {
	case 4:
		binary.BigEndian.PutUint32(src, uint32(t.(int32)))
	case 8:
		binary.BigEndian.PutUint64(src, uint64(t.(int64)))
	}
	return src
}

func encodeElem(aw *bytes.Buffer, e elem) {
	switch e.(type) {
	case doubleStruct:
		encodeDoubleStruct(aw, e.(doubleStruct))
	case stringStruct:
		encodeStringStruct(aw, e.(stringStruct))
	case documentStruct:
		encodeDocumentStruct(aw, e.(documentStruct))
	case arrayStruct:
		encodeArrayStruct(aw, e.(arrayStruct))
	case binaryStruct:
		encodeBinaryStruct(aw, e.(binaryStruct))
	case booleanStruct:
		encodeBooleanStruct(aw, e.(booleanStruct))
	case datetimeStruct:
		encodeDatetimeStruct(aw, e.(datetimeStruct))
	case int32Struct:
		encodeInt32Struct(aw, e.(int32Struct))
	case dbSizeStruct:
		encodedbSizeStruct(aw, e.(dbSizeStruct))
	}
}

func encodeBinaryStruct(aw *bytes.Buffer, e binaryStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeIntLittleEndian(int32(len(e.payload)), 4))
	aw.WriteByte(e.subtype)
	aw.Write(e.payload)
}

func encodeDocumentStruct(aw *bytes.Buffer, e documentStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeDoc(e.query))
}

func encodeArrayStruct(aw *bytes.Buffer, e arrayStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeDoc(e.doc))
}

func encodeStringStruct(aw *bytes.Buffer, e stringStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeIntLittleEndian(int32(len(e.value)+1), 4))
	aw.WriteString(e.value)
	aw.WriteByte(0x00)
}

func encodeBooleanStruct(aw *bytes.Buffer, e booleanStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	if e.value {
		aw.WriteByte(0x01)
	} else {
		aw.WriteByte(0x00)
	}
}

func encodeInt32Struct(aw *bytes.Buffer, e int32Struct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeIntLittleEndian(e.value, 4))
}

func encodeDatetimeStruct(aw *bytes.Buffer, e datetimeStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeIntLittleEndian(int64(e.value), 8))
}

func encodeDoubleStruct(aw *bytes.Buffer, e doubleStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(encodeIntBigEndian(e.value, 8))
}

func encodedbSizeStruct(aw *bytes.Buffer, e dbSizeStruct) {
	aw.WriteByte(e.elemType)
	aw.WriteString(e.name)
	aw.WriteByte(0x00)
	aw.Write(e.size)
}

func encodeLen(aw *bytes.Buffer) {
	str := aw.String()
	l := encodeIntLittleEndian(int32(4+aw.Len()), 4)
	aw.Reset()
	aw.Write(l)
	aw.WriteString(str)
}

func encodeDocuments(aw *bytes.Buffer, docs []document) {
	for _, doc := range docs {
		aw.Write(encodeDoc(doc))
	}
}

func encodeDoc(doc document) []byte {
	aw := bytes.NewBuffer([]byte(""))
	encodeElems(aw, doc.elements)
	aw.WriteByte(0x00) // = end of document
	encodeLen(aw)
	return aw.Bytes()
}

func encodeElems(aw *bytes.Buffer, es []elem) {
	for _, e := range es {
		encodeElem(aw, e)
	}
}

func startupTime() string {
	return string(time.Now().Format("2006-01-02T15:04:05.999-0700"))
}

func (s *mongodbService) encodeHeader(md *MsgData, aw *bytes.Buffer, msg interface{}) []byte {

	str := aw.String()
	aw.Reset()
	id := encodeIntLittleEndian(md.mh.RequestID, 4)
	aw.Write(encodeIntLittleEndian(int32(s.responseTo), 4))
	aw.Write(id)

	if md.mh.OpCode == 2004 {
		aw.Write(encodeIntLittleEndian(int32(1), 4))
	} else {
		switch msg.(type) {
		case *OpReplyMsg:
			aw.Write(encodeIntLittleEndian(int32(1), 4))
		case *OpCommandReplyMsg:
			aw.Write(encodeIntLittleEndian(int32(2011), 4))
		case *OpMsgMsg:
			aw.Write(encodeIntLittleEndian(int32(2013), 4))
		}
	}

	aw.WriteString(str)
	encodeLen(aw)
	return aw.Bytes()
}
