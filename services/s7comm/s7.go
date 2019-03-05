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
	"context"
	"encoding/binary"
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
		o(s)
	}
	return s
}

type s7commServiceConfig struct {
	HardWare  string `toml:"basic hardware"`
	SysName   string `toml:"system name"`
	Copyright string `toml:"copyright"`
	Version   string `toml:"version"`
	ModType   string `toml:"module type"`
	Mod       string `toml:"module"`
	SerialNum string `toml:"serial number"`
	PlantID   string `toml:"plant identification"`
}

type s7commService struct {
	s7commServiceConfig
	c pushers.Channel
}

func (s *s7commService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *s7commService) Handle(ctx context.Context, conn net.Conn) error {

	// Why this for-loop you ask? Well, ignoring RESET packets to keep a connection alive
	for {
		b := make([]byte, 4096)
		var bl int

		bl, err := conn.Read(b)

		if handleError(err) {
			conn.Close()
			return nil
		}
		P := unpack(b[:bl])

		if P.COTP.PDUType == com.CR {
			conn.Write(com.Cotp.ConnConfirm)
		}
		if P.S7.Parameter.SetupCom.Function == com.S7ConReq {
			conn.Write(com.S7comm.SetupComConf)
		}

		if P.S7.Data.SZLID == 0x0011 {
			conn.Write(com.Scan.PrimaryBasicResp)
		}

		if P.S7.Data.SZLID == 0x001c {
			conn.Write(com.Scan.SecondaryBasicResp)
			t := unpack(com.Scan.SecondaryBasicResp)
			log.Info("Version number: " + string(t.S7.Data.SZLDataTree[1].MlfB))
		}
	}
}

/*
===============================================================================
===============================================================================
*/

func unpack(m []byte) (P com.Packet) {

	P.TPKT = createTPTKpacket(&m)

	if P.TPKT.Version == 0x03 && P.TPKT.Reserved == 0x00 {
		P.COTP = createCOTPpacket(&m)
		if P.COTP.PDUType == com.COTPData {
			P.S7.Header = createS7Header(&m)
			P.S7.Parameter = createS7Parameters(&m, P.S7.Header)
			P.S7.Data = createS7Data(&m, P.S7.Header, P.S7.Parameter)

		}
	}
	return
}

func createTPTKpacket(m *[]byte) (TPKT com.TPKTPacket) {
	Reserved := binary.BigEndian.Uint16((*m)[1:3])
	TPKT = com.TPKTPacket{
		Version:  (*m)[0],
		Reserved: Reserved, //d.Int16((*m)[1:3]),
		Length:   (*m)[3],
	}
	(*m) = (*m)[4:]
	return
}

func createCOTPpacket(m *[]byte) (COTP com.COTPPacket) {
	COTP = com.COTPPacket{
		Length:  (*m)[0],
		PDUType: (*m)[1],
		DestRef: (*m)[3],
	}
	(*m) = (*m)[3:]
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
