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
	"github.com/honeytrap/honeytrap/services/decoder"
)

func (C *COTP) serialize(m []byte) (r []byte) {
	C.Length = 0x02
	C.PDUType = 0xf0
	C.DestRef = 0x80

	rb := &bytes.Buffer{}
	var eh errHandler

	eh.serializer(rb, C)
	eh.serializer(rb, m)

	if eh.err == nil {
		return rb.Bytes()
	}
	return nil
}

func (C *COTP) deserialize(m *[]byte) (verified bool) {

	dec := decoder.NewDecoder(*m)

	C.Length = dec.Byte()
	C.PDUType = dec.Byte()
	C.DestRef = dec.Byte()
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


	if len(m) > 0x11 && m[1] == 0xe0 && m[7] == 0xc1 && m[11] == 0xc2{
		dec := decoder.NewDecoder(m)

		var COTPRequest = COTPConnectRequest{
			Length:        dec.Byte(),
			PDUType:       dec.Byte(),
			DestRef:       dec.Uint16(),
			SourceRef:     dec.Uint16(),
			Reserved:      dec.Byte(),
			ParamSrcTSAP:  dec.Byte(),
			SourceTSAP:    dec.Copy(int(dec.Byte())),
			ParamDstTSAP:  dec.Byte(),
			DestTSAP:      dec.Copy(int(dec.Byte())),
			ParamTPDUSize: dec.Byte(),
			ParamTPDULen:  dec.Byte(),
			TPDUSize:      dec.Byte(),
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
			ParamSrcLen:   uint8(len(COTPRequest.SourceTSAP)),
			SourceTSAP:    COTPRequest.SourceTSAP,
			ParamDstTSAP:  COTPRequest.ParamDstTSAP,
			ParamDstLen:   uint8(len(COTPRequest.DestTSAP)),
			DestTSAP:      COTPRequest.DestTSAP,
		}

		buf := &bytes.Buffer{}

		var eh errHandler

		eh.serializer(buf, COTPResponse.Length)
		eh.serializer(buf, COTPResponse.PDUType)
		eh.serializer(buf, COTPResponse.DestRef)
		eh.serializer(buf, COTPResponse.SourceRef)
		eh.serializer(buf, COTPResponse.Reserved)
		eh.serializer(buf, COTPResponse.ParamTPDUSize)
		eh.serializer(buf, COTPResponse.ParamTPDULen)
		eh.serializer(buf, COTPResponse.TPDUSize)
		eh.serializer(buf, COTPResponse.ParamSrcTSAP)
		eh.serializer(buf, COTPResponse.ParamSrcLen)
		eh.serializer(buf, COTPResponse.SourceTSAP)
		eh.serializer(buf, COTPResponse.ParamDstTSAP)
		eh.serializer(buf, COTPResponse.ParamDstLen)
		eh.serializer(buf, COTPResponse.DestTSAP)

		if eh.err == nil {
			return T.serialize(buf.Bytes())
		}
	}
	return nil
}
