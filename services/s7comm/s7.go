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
	"net"
	"strconv"
	"strings"

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
	Hardware  string `toml:"basic_hardware"`
	SysName   string `toml:"system_name"`
	Copyright string `toml:"copyright"`
	Version   string `toml:"version"`
	ModType   string `toml:"module_type"`
	Mod       string `toml:"module"`
	SerialNum string `toml:"serial_number"`
	PlantID   string `toml:"plant_identification"`
	CPUType   string `toml:"cpu_type"`
}

type s7commService struct {
	s7commServiceConfig
	c pushers.Channel
}

func (s *s7commService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *s7commService) Handle(ctx context.Context, conn net.Conn) error {

	var cotp, s7 bool = false, false
	var err error
	var C COTP
	for {
		b := make([]byte, 4096)
		bl, err := conn.Read(b)
		b = b[:bl]
		if errCk(err) {
			break
		}
		if len(b) < 1 {
			break
		}
		if !cotp {

			response := C.connect(b)
			if response != nil {
				len, err := conn.Write(response)
				if err != nil || len < 1 {
					break
				}
				cotp = true
			}
		}
		if cotp {
			P, isS7 := unpackS7(b)

			if isS7 && !s7 {

				if P.S7.Parameter.SetupCom.Function == com.S7ConReq {
					response := S7ConResp(P)
					len, err := conn.Write(response)
					if err != nil || len < 1 {
						break
					}
					s7 = true
				}
			}
			if isS7 && s7 {
				if P.S7.Data.SZLID == 0x0011 {
					len, err := conn.Write(s.primRes(P))
					if err != nil || len < 1 {
						break
					}
				}
				if P.S7.Data.SZLID == 0x001c {
					len, err := conn.Write(s.secRes(P))
					if err != nil || len < 1 {
						break
					}
					err = conn.Close()
					if err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func S7ConResp(P com.Packet) (response []byte) {
	var T TPKT
	var C COTP
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
	buf := &bytes.Buffer{}
	S7HeadErr := binary.Write(buf, binary.BigEndian, resHead)
	S7ParaErr := binary.Write(buf, binary.BigEndian, resPara)

	if S7HeadErr == nil && S7ParaErr == nil {
		return T.serialize(C.serialize(buf.Bytes()))
	}
	return nil
}

func unpackS7(m []byte) (P com.Packet, isS7 bool) {
	if len(m) >= 4 {

		var T TPKT
		var C COTP

		chk := T.deserialize(&m)

		if chk {
			C.deserialize(&m)
			if C.PDUType == com.COTPData {
				P.S7.Header = createS7Header(&m)
				P.S7.Parameter = createS7Parameters(&m, P.S7.Header)
				P.S7.Data = createS7Data(&m, P.S7.Header, P.S7.Parameter)
			}
			return P, true
		}
	}
	return P, false
}
func (s *s7commService) secRes(P com.Packet) (response []byte) {
	partialResp, maxVal := s.generateSZL1()

	var T TPKT
	var C COTP
	var Data = com.S7DataNoSZL{
		ReturnCode:    0xff,
		TransportSize: 0x09,                         //no idea
		Length:        uint16(len(partialResp) + 8), //could be six, could be something else
		SZLID:         0x001c,
		SZLIndex:      0x0000,
		SZLListLength: uint16(maxVal),
		SZLListCount:  0x0a,
	}
	var Param = com.UserDataSmallHead{
		ParamHead:    0x112,
		ParamLength:  0x08,
		Method:       0x12,
		MethodType:   0x84,
		SubFunction:  0x01,
		SequenceNum:  P.S7.Parameter.UserData.SequenceNum + 1,
		DataRefNum:   0x03,
		LastDataUnit: 0x01,
		ErrorCode:    0x0000,
	}
	var Head = com.S7CustomHead{
		ProtocolID:  0x32,
		MessageType: 0x07,
		Reserved:    0x0000,
		PDURef:      0x00,
		ParamLength: 0x000c,
		DataLength:  Data.Length + 4,
	}

	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, Head)
	_ = binary.Write(buf, binary.BigEndian, Param)
	_ = binary.Write(buf, binary.BigEndian, Data)
	_ = binary.Write(buf, binary.BigEndian, partialResp)

	return T.serialize(C.serialize(buf.Bytes()))
}

func (s *s7commService) primRes(P com.Packet) (response []byte) {
	vA := strings.Split(s.Version, ".")
	vS := make([]byte, 3)
	for i := 0; i < len(vA); i++ {
		val, _ := strconv.Atoi(vA[i])
		vS[i] = byte(val)
	}

	SZL1 := append([]byte{0x00, 0x01}, []byte(s.Mod)...)
	SZL1 = append(SZL1, []byte{0x20, 0x20, 0x20, 0x00, 0x01, 0x20, 0x20}...)

	SZL2 := append([]byte{0x00, 0x06}, []byte(s.Mod)...)
	SZL2 = append(SZL2, []byte{0x20, 0x20, 0x20, 0x00, 0x01, 0x20, 0x20}...)

	SZL3 := append([]byte{0x00, 0x07}, []byte(s.Mod)...)
	SZL3 = append(SZL3, []byte{0x20, 0x20, 0x20, 0x56}...)
	SZL3 = append(SZL3, vS...)

	tb := &bytes.Buffer{}
	_ = binary.Write(tb, binary.BigEndian, SZL1)
	_ = binary.Write(tb, binary.BigEndian, SZL2)
	_ = binary.Write(tb, binary.BigEndian, SZL3)

	masterbuf := tb.Bytes()

	var Data = com.S7DataNoSZL{
		ReturnCode:    0xff,
		TransportSize: 0x09,                       //no idea
		Length:        uint16(len(masterbuf) + 8), //could be six, could be something else
		SZLID:         0x0011,
		SZLIndex:      0x0001,
		SZLListLength: 0x005c,
		SZLListCount:  0x03,
	}
	var Param = com.UserDataSmallHead{
		ParamHead:    0x112,
		ParamLength:  0x08,
		Method:       0x12,
		MethodType:   0x84,
		SubFunction:  0x01,
		SequenceNum:  P.S7.Parameter.UserData.SequenceNum + 1,
		DataRefNum:   0x00,
		LastDataUnit: 0x00,
		ErrorCode:    0x0000,
	}
	var Head = com.S7CustomHead{
		ProtocolID:  0x32,
		MessageType: 0x07,
		Reserved:    0x0000,
		PDURef:      0x00,
		ParamLength: 0x000c,
		DataLength:  Data.Length + 4,
	}
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, Head)
	_ = binary.Write(buf, binary.BigEndian, Param)
	_ = binary.Write(buf, binary.BigEndian, Data)
	_ = binary.Write(buf, binary.BigEndian, masterbuf)
	var C COTP
	var T TPKT

	return T.serialize(C.serialize(buf.Bytes()))
}

func (s *s7commService) generateSZL1() (partialresp []byte, SZLLen int) {

	InputLen := []string{s.SysName, s.ModType, s.PlantID, s.Copyright, s.SerialNum, s.CPUType}

	var lo string

	for _, cu := range InputLen {
		if len(cu) > len(lo) {
			lo = cu
		}
	}

	maxVal := 0x22
	Sys := append([]byte{0x00, 0x01}, s.SysName...)
	bufferFiller(&Sys, maxVal)
	Mtp := append([]byte{0x00, 0x02}, s.ModType...)
	bufferFiller(&Mtp, maxVal)
	PID := append([]byte{0x00, 0x03}, s.PlantID...)
	bufferFiller(&PID, maxVal)
	Cpr := append([]byte{0x00, 0x04}, s.Copyright...)
	bufferFiller(&Cpr, maxVal)
	Snr := append([]byte{0x00, 0x05}, s.SerialNum...)
	bufferFiller(&Snr, maxVal)
	CPU := append([]byte{0x00, 0x07}, s.CPUType...)
	bufferFiller(&CPU, maxVal)

	var masterbuf []byte

	masterbuf = append(Sys, Mtp...)
	masterbuf = append(masterbuf, PID...)
	masterbuf = append(masterbuf, Cpr...)
	masterbuf = append(masterbuf, Snr...)
	masterbuf = append(masterbuf, CPU...)

	return masterbuf, maxVal

}

func bufferFiller(m *[]byte, tl int) {
	tl = tl - len((*m))
	a := make([]byte, tl)
	(*m) = append((*m), a...)
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

func errCk(err error) bool {
	if err != nil {
		if err.Error() == "EOF" {
			return true
		}
	}

	return false
}
