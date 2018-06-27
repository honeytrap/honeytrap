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
	"strconv"
)

type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

type mongoMsg struct {
	mh MsgHeader
	rq string
}

func (s *mongodbService) MongoDBHandler(port int, bb *bytes.Buffer) []byte {
	p := strconv.Itoa(port)
	m := &mongoMsg{}
	m.decodeMsgHeader(bb)
	m.decodeRqData(bb)
	return s.encodeResp(m, p)
}

// decode the first 16 bytes of the header
func (m *mongoMsg) decodeMsgHeader(bb *bytes.Buffer) {
	m.mh.MessageLength = decodeInt32(bb)
	m.mh.RequestID = decodeInt32(bb)
	m.mh.ResponseTo = decodeInt32(bb)
	m.mh.OpCode = decodeInt32(bb)
}

// decode the whole request data
func (m *mongoMsg) decodeRqData(bb *bytes.Buffer) {
	m.rq = bb.String()
}

// decode int32 by reversing the bytes
func decodeInt32(bb *bytes.Buffer) int32 {
	i := make([]byte, 4)
	_, err := bb.Read(i[:4])
	if err != nil {
		log.Error("Error decoding int32: %s", err.Error())
	}
	return int32(binary.LittleEndian.Uint32(i))
}
