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
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/s7comm/com"
	logging "github.com/op/go-logging"
)

var (
	_ = services.Register("s7comm", S7)
)

var log = logging.MustGetLogger("services")

func S7(options ...services.ServicerFunc) services.Servicer {
	s := &s7commService{
		s7commServiceConfig: s7commServiceConfig{},
	}

	for _, o := range options {
		_ = o(s)
	}
	return s
}

type s7commServiceConfig struct {
	HardWare  string `toml:"basic_hardware"`
	SysName   string `toml:"system_name"`
	Copyright string `toml:"copyright"`
	Version   string `toml:"version"`
	ModType   string `toml:"module_type"`
	Mod       string `toml:"module"`
	SerialNum string `toml:"serial_number"`
	PlantID   string `toml:"plant_identification"`
}

type s7commService struct {
	s7commServiceConfig
	c pushers.Channel
}

func (s *s7commService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *s7commService) Handle(ctx context.Context, conn net.Conn) error {
	var err error
	COTPCONNECTED := false
	S7CONNECTED := false
	for {
		/* reading TCP buffer and storing into a local one */
		b := make([]byte, 4096)
		bl, err := conn.Read(b)
		b = b[:bl]

		/* Handing unknown input */
		if handleError(err) {
			return err
		}

		if len(b) < 1 {
			break
		}

		/* Creating a COTP connection with client */
		if !COTPCONNECTED {
			response := generateCOTPConResp(b)

			if response != nil {
				_, _ = conn.Write(response)
				COTPCONNECTED = true
				log.Info("COTP handshake confirmed")
			}
		}

		/* Handling S7 packets */
		if COTPCONNECTED {
			P, isS7 := unpackS7(b)

			if isS7 && !S7CONNECTED {
				if P.S7.Parameter.SetupCom.Function == com.S7ConReq {

					response := generateS7ConResp(P)
					_, _ = conn.Write(response)
					S7CONNECTED = true
					log.Info("S7 Job confirmed")
				}
			}

			if isS7 && S7CONNECTED {

				if P.S7.Data.SZLID == 0x0011 {
					_, _ = conn.Write(com.Scan.PrimaryBasicResp)
				}

				if P.S7.Data.SZLID == 0x001c {
					_, _ = conn.Write(com.Scan.SecondaryBasicResp)
					t, _ := unpackS7(com.Scan.SecondaryBasicResp)
					log.Info("Version number: " + string(t.S7.Data.SZLDataTree[1].MlfB))
					_ = conn.Close()
				}
			}
		}

	}
	return err
}

func generateS7ConResp(P com.Packet) (response []byte) {

	var resPara = com.S7SetupCom{
		Function:      P.S7.Parameter.SetupCom.Function,
		Reserved:      P.S7.Parameter.SetupCom.Reserved,
		MaxAmQCalling: P.S7.Parameter.SetupCom.MaxAmQCalling,
		MaxAmQCalled:  P.S7.Parameter.SetupCom.MaxAmQCalled,
		PDULength:     0xf0,
	}
	var resHead = com.S7Header{
		ProtocolID:  0x32,
		MessageType: com.AckData,
		Reserved:    P.S7.Header.Reserved,
		PDURef:      P.S7.Header.PDURef,
		ParamLength: P.S7.Header.ParamLength,
		DataLength:  P.S7.Header.DataLength,
		ErrorClass:  0x00,
		ErrorCode:   0x00,
	}
	/* No need for rebuilding COTP */

	var TPKT = com.TPKTPacket{
		Version:  P.TPKT.Version,
		Reserved: P.TPKT.Reserved,
		Length:   P.TPKT.Length + 0x02,
	}

	/* Building response buffer */
	buf := &bytes.Buffer{}
	TPKTErr := binary.Write(buf, binary.BigEndian, TPKT)
	COTPErr := binary.Write(buf, binary.BigEndian, P.COTP)
	S7HeadErr := binary.Write(buf, binary.BigEndian, resHead)
	S7ParaErr := binary.Write(buf, binary.BigEndian, resPara)

	if TPKTErr == nil && COTPErr == nil && S7HeadErr == nil && S7ParaErr == nil {
		return buf.Bytes()
	}
	return nil
}

/* unpacking incoming packet, checking if it contains a COTP Connection Request and create a response */
func generateCOTPConResp(m []byte) (response []byte) {
	if len(m) > 0 {
		var P com.Packet
		var TPKTcheck bool
		P.TPKT, TPKTcheck = createTPTKpacket(&m)
		fmt.Printf("\nTPKT check: \t %v \n", TPKTcheck)
		if TPKTcheck && P.TPKT.Length == 0x16 {
			response := createCOTPCon(m, P)

			if response != nil {
				return response
			}
		}
	}

	return nil
}

/* unpacking incoming packet and loading it into a Packet variable */
func unpackS7(m []byte) (P com.Packet, isS7 bool) {
	if len(m) >= 4 {
		var TPKTcheck bool
		P.TPKT, TPKTcheck = createTPTKpacket(&m)

		if TPKTcheck {
			P.COTP = createCOTPpacket(&m)
			if P.COTP.PDUType == com.COTPData {
				P.S7.Header = createS7Header(&m)
				P.S7.Parameter = createS7Parameters(&m, P.S7.Header)
				P.S7.Data = createS7Data(&m, P.S7.Header, P.S7.Parameter)
			}
			return P, true
		}
	}
	return P, false

}

func createTPTKpacket(m *[]byte) (TPKT com.TPKTPacket, verify bool) {
	Reserved := binary.BigEndian.Uint16((*m)[1:3])
	TPKT = com.TPKTPacket{
		Version:  (*m)[0],
		Reserved: Reserved, //d.Int16((*m)[1:3]),
		Length:   (*m)[3],
	}

	if TPKT.Version == 0x03 && TPKT.Reserved == 0x00 && int(TPKT.Length)-len(*m) == 0 {
		(*m) = (*m)[4:]
		return TPKT, true
	}
	return TPKT, false
}

func createCOTPpacket(m *[]byte) (COTP com.COTPPacket) {
	if (*m)[0] == 0x02 {
		COTP = com.COTPPacket{
			Length:  (*m)[0],
			PDUType: (*m)[1],
			DestRef: (*m)[2],
		}
		(*m) = (*m)[3:]
		return
	}
	return
}

func createS7Header(mp *[]byte) (H com.S7Header) {
	var m = (*mp)
	if m[0] == 0x32 {
		Reserved := binary.BigEndian.Uint16(m[2:4])
		PDURef := binary.LittleEndian.Uint16(m[4:6])
		ParamLength := binary.BigEndian.Uint16(m[6:8])
		DataLength := binary.BigEndian.Uint16(m[8:10])

		H = com.S7Header{
			ProtocolID:  m[0],
			MessageType: m[1],
			Reserved:    Reserved,
			PDURef:      PDURef,
			ParamLength: ParamLength,
			DataLength:  DataLength,
		}

		if H.MessageType == com.AckData {
			H.ErrorClass = m[10]
			H.ErrorCode = m[11]
			(*mp) = (*mp)[12:]
		} else {
			(*mp) = (*mp)[10:]
		}

		return
	}
	return
}

func createS7Parameters(mp *[]byte, H com.S7Header) (P com.S7Parameter) {
	var m = (*mp)
	if H.MessageType == com.Request {
		AmQCalling := binary.BigEndian.Uint16(m[2:4])
		AmQCalled := binary.BigEndian.Uint16(m[4:6])
		PDULen := binary.BigEndian.Uint16(m[6:8])
		P = com.S7Parameter{
			SetupCom: com.S7SetupCom{
				Function:      m[0],
				Reserved:      m[1],
				MaxAmQCalling: AmQCalling,
				MaxAmQCalled:  AmQCalled,
				PDULength:     PDULen,
			},
		}

	} else if H.MessageType == com.UserData {

		ParamHead := binary.BigEndian.Uint32(m[0:4]) >> 8
		P = com.S7Parameter{
			UserData: com.S7UserData{
				ParamHead:      ParamHead,
				ParamLength:    m[3],
				Method:         m[4],
				MethodType:     m[5] >> 4,
				MethodFunction: m[5] << 4 >> 4,
				SubFunction:    m[6],
				SequenceNum:    m[7],
			},
		}
		if P.UserData.MethodType == com.S7DataResponse {
			P.UserData.DataRefNum = m[8]
			P.UserData.LastDataUnit = m[9]
			P.UserData.ErrorCode = binary.BigEndian.Uint16(m[10:12])
			(*mp) = (*mp)[12:]
			return
		}

	}

	(*mp) = (*mp)[8:]
	return
}

func createS7Data(mp *[]byte, H com.S7Header, P com.S7Parameter) (D com.S7Data) {
	var m = (*mp)
	if H.MessageType == com.UserData {

		if P.UserData.MethodType == com.S7DataRequest {
			Length := binary.BigEndian.Uint16(m[2:4])
			SZLID := binary.BigEndian.Uint16(m[4:6])
			SZLIndex := binary.BigEndian.Uint16(m[6:8])

			D = com.S7Data{
				ReturnCode:    m[0],
				TransportSize: m[1],
				Length:        Length,
				SZLID:         SZLID,
				SZLIndex:      SZLIndex,
			}

		} else if P.UserData.MethodType == com.S7DataResponse {

			Length := binary.BigEndian.Uint16(m[2:4])
			SZLID := binary.BigEndian.Uint16(m[4:6])
			SZLIndex := binary.BigEndian.Uint16(m[6:8])
			SZLListLength := binary.BigEndian.Uint16(m[8:10])
			SZLListCount := binary.BigEndian.Uint16(m[10:12])

			D = com.S7Data{
				ReturnCode:    m[0],
				TransportSize: m[1],
				Length:        Length,
				SZLID:         SZLID,
				SZLIndex:      SZLIndex,
				SZLListLength: SZLListLength,
				SZLListCount:  SZLListCount,
			}
			offset := int(D.SZLListLength)

			for i := 0; i < int(D.SZLListCount); i++ {
				Index := binary.BigEndian.Uint16(m[12+offset*i : 14+offset*i])
				MlfB := m[14+offset*i : 34+offset*i]
				BGType := binary.BigEndian.Uint16(m[34+offset*i : 36+offset*i])
				Ausbg := binary.BigEndian.Uint16(m[36+offset*i : 38+offset*i])
				Ausbe := binary.BigEndian.Uint16(m[38+offset*i : 40+offset*i])

				var DT = com.SLZDataTree{
					Index:  Index,
					MlfB:   MlfB,
					BGType: BGType,
					Ausbg:  Ausbg,
					Ausbe:  Ausbe,
				}
				D.AddSLZDataTree(DT)
			}
		}
	}
	return
}

func handleError(err error) bool {
	if err != nil {
		if err.Error() == "EOF" {
			return true
		}
	}

	return false
}

func createCOTPCon(m []byte, P com.Packet) (response []byte) {

	if len(m) == 0x12 {
		DestRef := binary.BigEndian.Uint16(m[2:4])
		SourceRef := binary.BigEndian.Uint16(m[4:6])
		SourceTSAP := binary.BigEndian.Uint16(m[9:11])
		DestTSAP := binary.BigEndian.Uint16(m[13:15])

		var COTPRequest = com.COTPConnectRequest{
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

		var COTPResponse = com.COTPConnectConfirm{
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
		TPKTerr := binary.Write(buf, binary.BigEndian, P.TPKT)
		COTPerr := binary.Write(buf, binary.BigEndian, COTPResponse)

		if TPKTerr == nil && COTPerr == nil {
			return buf.Bytes()
		}
	}
	return nil
}
