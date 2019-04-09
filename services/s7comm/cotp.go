/* Copyright 2016-2019 DutchSec (https://dutchsec.com/)
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
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
	if !T.deserialize(&m) {
		return nil
	} else if r := createCOTPCon(m); r == nil {
		return nil
	} else {
		return r
	}
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
