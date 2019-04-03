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

package s7comm

import (
	"bytes"
	"encoding/binary"
)

func (C *COTP) serialize(m []byte) (r []byte) {
	C.Length = 0x02
	C.PDUType = 0xf0
	C.DestRef = 0x80

	rb := &bytes.Buffer{}
	CErr := binary.Write(rb, binary.BigEndian, C)
	mErr := binary.Write(rb, binary.BigEndian, m)

	if CErr != nil || mErr != nil {
		/* Print error message to console */
		return nil
	}
	return rb.Bytes()
}

func (C *COTP) deserialize(m *[]byte) (verified bool) {
	C.Length = (*m)[0]
	C.PDUType = (*m)[1]
	C.DestRef = (*m)[2]
	if C.verify() == 0x01 {
		*m = (*m)[3:]
		return true
	}
	return false
}

func (C *COTP) verify() (isCOTP int) {
	if C.Length == 0x02 && C.PDUType == 0xf0 && C.DestRef == 0x80 {
		return 0x01
	} else if C.PDUType == 0xd0 || C.PDUType == 0xe0 {
		return 0x02
	} else if C.PDUType == 0xf0 {
	}
	return 0x00
}

func (C *COTP) connect(m []byte) (response []byte) {
	var T TPKT
	if T.deserialize(&m) {
		r := createCOTPCon(m)
		if r != nil {
			return r
		}
	}
	return nil
}

func createCOTPCon(m []byte) (response []byte) {
	var T TPKT
	if len(m) > 0x11 {
		DestRef := binary.BigEndian.Uint16(m[2:4])
		SourceRef := binary.BigEndian.Uint16(m[4:6])

		var COTPRequest = COTPConnectRequest{
			Length:        m[0],
			PDUType:       m[1],
			DestRef:       DestRef,
			SourceRef:     SourceRef,
			Reserved:      m[6],
			ParamSrcTSAP:  m[7],
			ParamSrcLen:   m[8],
			SourceTSAP:    m[9 : 9+m[8]],
			ParamDstTSAP:  m[9+m[8]],
			ParamDstLen:   m[10+m[8]],
			DestTSAP:      m[11+m[8] : 11+m[8]+m[10+m[8]]],
			ParamTPDUSize: m[len(m)-3],
			ParamTPDULen:  m[len(m)-2],
			TPDUSize:      m[len(m)-1],
		}

		var COTPResponse = COTPConnectConfirm{
			Length:        COTPRequest.Length,
			PDUType:       CC,
			DestRef:       COTPRequest.SourceRef,
			SourceRef:     COTPRequest.DestRef,
			Reserved:      COTPRequest.Reserved,
			ParamTPDUSize: COTPRequest.ParamTPDUSize,
			ParamTPDULen:  COTPRequest.ParamTPDULen,
			TPDUSize:      COTPRequest.TPDUSize,
			ParamSrcTSAP:  COTPRequest.ParamSrcTSAP,
			ParamSrcLen:   COTPRequest.ParamSrcLen,
			SourceTSAP:    COTPRequest.SourceTSAP,
			ParamDstTSAP:  COTPRequest.ParamDstTSAP,
			ParamDstLen:   COTPRequest.ParamDstLen,
			DestTSAP:      COTPRequest.DestTSAP,
		}
		/* This is a temporary fix for converting the COTP struct to a usable byte slice. This fix is used because the dynamic Dest & Source reference values cannot be written to binary via buf.Bytes()*/
		buf := &bytes.Buffer{}
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.Length)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.PDUType)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.DestRef)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.SourceRef)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.Reserved)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamTPDUSize)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamTPDULen)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.TPDUSize)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamSrcTSAP)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamSrcLen)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.SourceTSAP)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamDstTSAP)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.ParamDstLen)
		_ = binary.Write(buf, binary.BigEndian, COTPResponse.DestTSAP)
		return T.serialize(buf.Bytes())
	}
	return nil
}

type COTP struct {
	Length  uint8
	PDUType uint8
	DestRef uint8
}

type COTPConnect struct {
	Length        uint8
	PDUType       uint8
	DestRef       uint16
	SourceRef     uint16
	Reserved      uint8
	ParamTPDUSize uint8
	ParamTPDULen  uint8
	TPDUSize      uint8
	ParamSrcTSAP  uint8
	ParamSrcLen   uint8
	SourceTSAP    uint16 //should be byte array probably
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      uint16
}

type COTPConnectRequest struct {
	Length        uint8
	PDUType       uint8
	DestRef       uint16
	SourceRef     uint16
	Reserved      uint8
	ParamSrcTSAP  uint8
	ParamSrcLen   uint8
	SourceTSAP    []byte
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      []byte
	ParamTPDUSize uint8
	ParamTPDULen  uint8
	TPDUSize      uint8
}

type COTPConnectConfirm struct {
	Length        uint8
	PDUType       uint8
	DestRef       uint16
	SourceRef     uint16
	Reserved      uint8
	ParamTPDUSize uint8
	ParamTPDULen  uint8
	TPDUSize      uint8
	ParamSrcTSAP  uint8
	ParamSrcLen   uint8
	SourceTSAP    []byte
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      []byte
}
