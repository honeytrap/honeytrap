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
	"strconv"
	"strings"

	"github.com/honeytrap/honeytrap/services/decoder"
)

func (s7 *S7Packet) deserialize(m []byte) (P Packet, isS7 bool) {
	if len(m) >= 4 {
		if s7.T.deserialize(&m) {
			s7.C.deserialize(&m)
			if s7.C.PDUType == COTPData {
				P.S7.Header = createS7Header(&m)
				P.S7.Parameter = createS7Parameters(&m, P.S7.Header)
				P.S7.Data = createS7Data(&m, P.S7.Header, P.S7.Parameter)
			}
			return P, true
		}
	}
	return P, false
}

func (s7 *S7Packet) connect(P Packet) (response []byte) {
	var resPara = S7SetupCom{
		Function:      P.S7.Parameter.SetupCom.Function,
		Reserved:      P.S7.Parameter.SetupCom.Reserved,
		MaxAmQCalling: P.S7.Parameter.SetupCom.MaxAmQCalling,
		MaxAmQCalled:  P.S7.Parameter.SetupCom.MaxAmQCalled,
		PDULength:     0xf0,
	}
	var resHead = S7Header{
		ProtocolID:  0x32,
		MessageType: AckData,
		Reserved:    P.S7.Header.Reserved,
		PDURef:      P.S7.Header.PDURef,
		ParamLength: P.S7.Header.ParamLength,
		DataLength:  P.S7.Header.DataLength,
		ErrorClass:  0x00,
		ErrorCode:   0x00,
	}

	var eh errHandler
	buf := &bytes.Buffer{}

	eh.serializer(buf, resHead)
	eh.serializer(buf, resPara)

	if eh.err == nil {
		return s7.T.serialize(s7.C.serialize(buf.Bytes()))
	}
	return nil
}

func (s7 *S7Packet) secRes(P Packet) (response []byte) {

	maxLen := 0x22
	Sys := append([]byte{0x00, 0x01}, s7.ui.SysName...)
	bufferFiller(&Sys, maxLen)
	Mtp := append([]byte{0x00, 0x02}, s7.ui.ModType...)
	bufferFiller(&Mtp, maxLen)
	PID := append([]byte{0x00, 0x03}, s7.ui.PlantID...)
	bufferFiller(&PID, maxLen)
	Cpr := append([]byte{0x00, 0x04}, s7.ui.Copyright...)
	bufferFiller(&Cpr, maxLen)
	Snr := append([]byte{0x00, 0x05}, s7.ui.SerialNum...)
	bufferFiller(&Snr, maxLen)
	CPU := append([]byte{0x00, 0x07}, s7.ui.CPUType...)
	bufferFiller(&CPU, maxLen)

	sp := append(Sys, Mtp...)
	sp = append(sp, PID...)
	sp = append(sp, Cpr...)
	sp = append(sp, Snr...)
	sp = append(sp, CPU...)

	var Data = S7DataNoSZL{
		ReturnCode:    0xff,
		TransportSize: 0x09,
		Length:        uint16(len(sp) + 8),
		SZLID:         0x001c,
		SZLIndex:      0x0000,
		SZLListLength: uint16(maxLen),
		SZLListCount:  0x0a,
	}
	var Param = UserDataSmallHead{
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
	var Head = S7CustomHead{
		ProtocolID:  0x32,
		MessageType: 0x07,
		Reserved:    0x0000,
		PDURef:      0x00,
		ParamLength: 0x000c,
		DataLength:  Data.Length + 4,
	}

	buf := &bytes.Buffer{}
	var eh errHandler

	eh.serializer(buf, Head)
	eh.serializer(buf, Param)
	eh.serializer(buf, Data)
	eh.serializer(buf, sp)

	if eh.err == nil {
		return s7.T.serialize(s7.C.serialize(buf.Bytes()))
	}

	return nil
}

func (s7 *S7Packet) primRes(P Packet) (response []byte) {
	vA := strings.Split(s7.ui.Version, ".")
	vS := make([]byte, 3)
	for i := 0; i < len(vA); i++ {
		val, _ := strconv.Atoi(vA[i])
		vS[i] = byte(val)
	}

	SZL1 := append([]byte{0x00, 0x01}, []byte(s7.ui.Mod)...)
	SZL1 = append(SZL1, []byte{0x20, 0x20, 0x20, 0x00, 0x01, 0x20, 0x20}...)

	SZL2 := append([]byte{0x00, 0x06}, []byte(s7.ui.Mod)...)
	SZL2 = append(SZL2, []byte{0x20, 0x20, 0x20, 0x00, 0x01, 0x20, 0x20}...)

	SZL3 := append([]byte{0x00, 0x07}, []byte(s7.ui.Mod)...)
	SZL3 = append(SZL3, []byte{0x20, 0x20, 0x20, 0x56}...)
	SZL3 = append(SZL3, vS...)

	tb := &bytes.Buffer{}
	var eh errHandler
	eh.serializer(tb, SZL1)
	eh.serializer(tb, SZL2)
	eh.serializer(tb, SZL3)

	sp := tb.Bytes()

	var Data = S7DataNoSZL{
		ReturnCode:    0xff,
		TransportSize: 0x09,
		Length:        uint16(len(sp) + 8),
		SZLID:         0x0011,
		SZLIndex:      0x0001,
		SZLListLength: 0x005c,
		SZLListCount:  0x03,
	}
	var Param = UserDataSmallHead{
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
	var Head = S7CustomHead{
		ProtocolID:  0x32,
		MessageType: 0x07,
		Reserved:    0x0000,
		PDURef:      0x00,
		ParamLength: 0x000c,
		DataLength:  Data.Length + 4,
	}
	buf := &bytes.Buffer{}

	eh.serializer(buf, Head)
	eh.serializer(buf, Param)
	eh.serializer(buf, Data)
	eh.serializer(buf, sp)
	if eh.err == nil {
		return s7.T.serialize(s7.C.serialize(buf.Bytes()))
	}
	return nil
}

func bufferFiller(m *[]byte, tl int) {
	tl = tl - len((*m))
	a := make([]byte, tl)
	(*m) = append((*m), a...)
}

func createS7Header(m *[]byte) (H S7Header) {

	if (*m)[0] == 0x32 {

		dec := decoder.NewDecoder(*m)
		H = S7Header{
			ProtocolID:  dec.Byte(),
			MessageType: dec.Byte(),
			Reserved:    dec.Uint16(),
			PDURef:      dec.Uint16Le(),
			ParamLength: dec.Uint16(),
			DataLength:  dec.Uint16(),
		}

		if H.MessageType == AckData {
			H.ErrorClass = dec.Byte()
			H.ErrorCode = dec.Byte()
			(*m) = (*m)[12:]
		} else {
			(*m) = (*m)[10:]
		}
	}
	return
}

func createS7Parameters(m *[]byte, H S7Header) (P S7Parameter) {
	dec := decoder.NewDecoder(*m)
	if H.MessageType == Request {
		P = S7Parameter{
			SetupCom: S7SetupCom{
				Function:      dec.Byte(),
				Reserved:      dec.Byte(),
				MaxAmQCalling: dec.Uint16(),
				MaxAmQCalled:  dec.Uint16(),
				PDULength:     dec.Uint16(),
			},
		}

	} else if H.MessageType == UserData {

		ph := append([]byte{0x00}, dec.Copy(3)...)
		dec1 := decoder.NewDecoder(ph)
		paraHead := dec1.Uint32()

		P = S7Parameter{
			UserData: S7UserData{
				ParamHead:   paraHead,
				ParamLength: dec.Byte(),
				Method:      dec.Byte(),
				MethodType:  dec.Byte(),
				SubFunction: dec.Byte(),
				SequenceNum: dec.Byte(),
			},
		}
		P.UserData.MethodFunction = P.UserData.MethodType << 4 >> 4
		P.UserData.MethodType = P.UserData.MethodType >> 4

		if P.UserData.MethodType == S7DataResponse {
			P.UserData.DataRefNum = dec.Byte()
			P.UserData.LastDataUnit = dec.Byte()
			P.UserData.ErrorCode = dec.Uint16()
			(*m) = (*m)[12:]
			return
		}

	}

	(*m) = (*m)[8:]
	return
}

func createS7Data(mp *[]byte, H S7Header, P S7Parameter) (D S7Data) {
	dec := decoder.NewDecoder(*mp)
	var m = (*mp)
	if H.MessageType == UserData {
		if P.UserData.MethodType == S7DataRequest {
			D = S7Data{
				ReturnCode:    dec.Byte(),
				TransportSize: dec.Byte(),
				Length:        dec.Uint16(),
				SZLID:         dec.Uint16(),
				SZLIndex:      dec.Uint16(),
			}

		} else if P.UserData.MethodType == S7DataResponse {
			D = S7Data{
				ReturnCode:    dec.Byte(),
				TransportSize: dec.Byte(),
				Length:        dec.Uint16(),
				SZLID:         dec.Uint16(),
				SZLIndex:      dec.Uint16(),
				SZLListLength: dec.Uint16(),
				SZLListCount:  dec.Uint16(),
			}
			offset := int(D.SZLListLength)

			for i := 0; i < int(D.SZLListCount); i++ {
				Index := binary.BigEndian.Uint16(m[12+offset*i : 14+offset*i])
				MlfB := m[14+offset*i : 34+offset*i]
				BGType := binary.BigEndian.Uint16(m[34+offset*i : 36+offset*i])
				Ausbg := binary.BigEndian.Uint16(m[36+offset*i : 38+offset*i])
				Ausbe := binary.BigEndian.Uint16(m[38+offset*i : 40+offset*i])

				var DT = SLZDataTree{
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

func (sd *S7Data) AddSLZDataTree(dt SLZDataTree) {
	sd.SZLDataTree = append(sd.SZLDataTree, dt)
}
