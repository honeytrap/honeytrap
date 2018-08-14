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
	"io"
)

// func decodeMsgHeader(bb *bytes.Buffer) *MsgHeader {
func decodeMsgHeader(bb *bytes.Buffer, md *MsgData) {
	md.mh = MsgHeader{
		MessageLength: decodeInt32(bb),
		RequestID:     decodeInt32(bb),
		ResponseTo:    decodeInt32(bb),
		OpCode:        decodeInt32(bb),
	}
}

func decodeInt32(bb *bytes.Buffer) int32 {
	i := make([]byte, 4)
	_, err := bb.Read(i[:4])
	if err != nil {
		log.Error("Error decoding int32: %s", err.Error())
	}
	return int32(binary.LittleEndian.Uint32(i))
}

func decodeLengthDocument(bb *bytes.Buffer) (int32, error) {
	i := make([]byte, 4)
	_, err := bb.Read(i[:4])

	if err == io.EOF {
		return 0, err
	} else if err != nil {
		log.Error("Error decoding length document: %s", err.Error())
	}

	return int32(binary.LittleEndian.Uint32(i)), nil
}

func decodeInt64(bb *bytes.Buffer) int64 {
	i := make([]byte, 8)
	_, err := bb.Read(i[:8])
	if err != nil {
		log.Error("Error decoding int64: %s", err.Error())
	}
	return int64(binary.LittleEndian.Uint64(i))
}

func parseOpCommand(bb *bytes.Buffer, ev map[string]interface{}) *OpCommandMsg {
	ocm := &OpCommandMsg{}
	ocm.database = decodeString(bb)
	ev["database.name"] = ocm.database
	ocm.commandName = decodeString(bb)
	ocm.metadata = *decodeDocument(bb, ev, "")
	ocm.commandArgs = *decodeDocument(bb, ev, "")
	// inputDocs   []document // zero or more documents
	return ocm
}

func parseOpQuery(bb *bytes.Buffer, ev map[string]interface{}) *OpQueryMsg {
	oqm := &OpQueryMsg{}
	oqm.flags = decodeInt32(bb)
	oqm.fullCollectionName = decodefullCollectionName(bb, ev)
	oqm.numberToSkip = decodeInt32(bb)
	oqm.numberToReturn = decodeInt32(bb)
	oqm.query = *decodeDocument(bb, ev, "")
	oqm.returnFieldsSelector = *decodeDocument(bb, ev, "")
	return oqm
}

func decodeDocument(bb *bytes.Buffer, ev map[string]interface{}, documentName string) *document {
	d := &document{}
	var err error
	d.length, err = decodeLengthDocument(bb)
	if err == io.EOF { // No document  ex: returnFieldsSelector is optional
		return d
	}
	d.elements = decodeElements(bb, ev, documentName)
	return d
}

func decodeElements(bb *bytes.Buffer, ev map[string]interface{}, documentName string) []elem {
	elements := []elem{}
	for {
		by, err := bb.ReadByte()

		if err != nil {
			log.Error("Error decoding Elements: %s", err.Error())
		}
		if by == 0x00 {
			break
		}
		elem := decodeElem(bb, by, ev, documentName)
		elements = append(elements, elem)

	}
	return elements
}

func decodeElem(bb *bytes.Buffer, by byte, ev map[string]interface{}, documentName string) interface{} {

	var elem struct{}

	switch by {

	case 0x01:
		return doubleStruct{
			0x01,
			decodeString(bb),
			decodeInt64(bb),
		}

	case 0x02:
		key := decodeString(bb)
		a := decodeInt32(bb)
		value := decodeString(bb)

		key = documentName + "." + key
		ev[key] = value
		return stringStruct{
			0x02,
			key,
			a,
			value,
		}

	case 0x03:
		documentName := decodeString(bb)
		return documentStruct{
			0x03,
			documentName,
			*decodeDocument(bb, ev, documentName),
		}
	case 0x04:
		return arrayStruct{
			0x04,
			decodeString(bb),
			*decodeDocument(bb, ev, documentName),
		}

	case 0x05:
		str := decodeString(bb)
		l := decodeInt32(bb)
		subtype, _ := bb.ReadByte()
		return binaryStruct{
			0x05,
			str,
			l,
			subtype,
			decodePayload(bb, l, ev),
		}

	case 0x08:
		return booleanStruct{
			0x08,
			decodeString(bb),
			decodeBool(bb),
		}

	case 0x10:
		return int32Struct{
			0x10,
			decodeString(bb),
			decodeInt32(bb),
		}

	default:
		log.Error("Error decoding new Elem: type unknown: ", by)
	}
	return elem
}

func decodePayload(bb *bytes.Buffer, len int32, ev map[string]interface{}) []byte {
	pay := make([]byte, len)

	_, err := bb.Read(pay[:len])
	if err != nil {
		log.Error("Error decoding int64: %s", err.Error())
	}
	ev["binaryPayload"] = pay
	return pay
}

func decodefullCollectionName(bb *bytes.Buffer, ev map[string]interface{}) fullCollectionName {
	f := fullCollectionName{}
	tmp, err := bb.ReadBytes(0x00)
	if err != nil {
		log.Error("Error decoding fullCollectionName: %s", err.Error())
	}
	tmp2 := bytes.Split(tmp[:len(tmp)-1], []byte("."))
	f.databaseName = string(tmp2[0])
	f.collectionName = string(tmp2[1])

	ev["database.name"] = f.databaseName
	ev["collection.name"] = f.collectionName

	return f
}

func decodeString(bb *bytes.Buffer) string {
	s, err := bb.ReadBytes(0x00)

	if err == io.EOF {
		er := bb.WriteByte(0x00) // TODO
		if er != nil {
			log.Error("Error when adding 0x00: %s", er.Error())
		}

	} else if err != nil {
		log.Error("Error decoding string: %s", err.Error())
	}

	return string(s[:len(s)-1])
}

func decodeBool(bb *bytes.Buffer) bool {
	by, err := bb.ReadByte()
	if err != nil {
		log.Error("Error decoding bool: %s", err.Error())
	}
	if by == 0x01 {
		return true
	}
	if by != 0x00 {
		log.Error("Error Wrong value when decoding bool: get ", by)
	}
	return false
}
