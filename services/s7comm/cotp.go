package s7comm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/honeytrap/honeytrap/services/s7comm/com"
)

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
	SourceTSAP    uint16
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      uint16
}

func (C *COTP) serialize(m []byte) (r []byte) {
	fmt.Println("Serialize")
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
	fmt.Println("Deserialize")
	C.Length = (*m)[0]
	C.PDUType = (*m)[1]
	C.DestRef = (*m)[2]
	fmt.Printf("COTP VERIFY: %v\n", C.verify())
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
	if len(m) == 0x12 {
		DestRef := binary.BigEndian.Uint16(m[2:4])
		SourceRef := binary.BigEndian.Uint16(m[4:6])
		SourceTSAP := binary.BigEndian.Uint16(m[9:11])
		DestTSAP := binary.BigEndian.Uint16(m[13:15])

		var COTPRequest = COTPConnectRequest{
			Length:        m[0],
			PDUType:       m[1],
			DestRef:       DestRef,
			SourceRef:     SourceRef,
			Reserved:      m[6],
			ParamSrcTSAP:  m[7],
			ParamSrcLen:   m[8],
			SourceTSAP:    SourceTSAP,
			ParamDstTSAP:  m[11],
			ParamDstLen:   m[12],
			DestTSAP:      DestTSAP,
			ParamTPDUSize: m[15],
			ParamTPDULen:  m[16],
			TPDUSize:      m[17],
		}

		var COTPResponse = COTPConnectConfirm{
			Length:        COTPRequest.Length,
			PDUType:       com.CC,
			DestRef:       COTPRequest.SourceRef,
			SourceRef:     0x02,
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

		/* Building response buffer */
		buf := &bytes.Buffer{}
		Cerr := binary.Write(buf, binary.BigEndian, COTPResponse)
		if Cerr == nil {
			return T.serialize(buf.Bytes())
		}
		//COTPerr := binary.Write(buf, binary.BigEndian, COTPResponse)

		//if TPKTerr == nil && COTPerr == nil {
		//	return buf.Bytes()
		//}
	}
	return nil
}

type COTPConnectRequest struct {
	Length        uint8
	PDUType       uint8
	DestRef       uint16
	SourceRef     uint16
	Reserved      uint8
	ParamSrcTSAP  uint8
	ParamSrcLen   uint8
	SourceTSAP    uint16
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      uint16
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
	SourceTSAP    uint16
	ParamDstTSAP  uint8
	ParamDstLen   uint8
	DestTSAP      uint16
}
