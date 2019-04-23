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

/**********************************************
TKTP related datastructures
**********************************************/

type TPKT struct {
	Version  uint8
	Reserved uint8
	Length   uint16
}

/**********************************************
COTP related datastructures
**********************************************/

type COTP struct {
	Length  uint8
	PDUType uint8
	DestRef uint8
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

/**********************************************
s7comm related datastructures
**********************************************/

type S7Packet struct {
	T         TPKT
	C         COTP
	Header    S7Header
	Parameter S7Parameter
	Data      S7Data
	ui        userinput
}
type S7CustomHead struct {
	ProtocolID  uint8
	MessageType uint8
	Reserved    uint16
	PDURef      uint16
	ParamLength uint16
	DataLength  uint16
}
type S7DataNoSZL struct {
	ReturnCode    uint8
	TransportSize uint8
	Length        uint16
	SZLID         uint16
	SZLIndex      uint16
	SZLListLength uint16
	SZLListCount  uint16
}
type S7SetupCom struct {
	Function      uint8
	Reserved      uint8
	MaxAmQCalling uint16
	MaxAmQCalled  uint16
	PDULength     uint16
}
type S7Header struct {
	ProtocolID  uint8
	MessageType uint8
	Reserved    uint16
	PDURef      uint16
	ParamLength uint16
	DataLength  uint16
	ErrorClass  uint8
	ErrorCode   uint8
}
type S7Parameter struct {
	SetupCom S7SetupCom
	UserData S7UserData
}
type S7UserData struct {
	ParamHead      uint32
	ParamLength    uint8
	Method         uint8
	MethodType     uint8
	MethodFunction uint8
	SubFunction    uint8
	SequenceNum    uint8
	DataRefNum     uint8
	LastDataUnit   uint8
	ErrorCode      uint16
}
type S7Data struct {
	ReturnCode    uint8
	TransportSize uint8
	Length        uint16
	SZLID         uint16
	SZLIndex      uint16
	SZLListLength uint16
	SZLListCount  uint16
	SZLDataTree   []SLZDataTree
}
type UserDataSmallHead struct {
	Reserved     uint8
	ParamHead    uint16
	ParamLength  uint8
	Method       uint8
	MethodType   uint8
	SubFunction  uint8
	SequenceNum  uint8
	DataRefNum   uint8
	LastDataUnit uint8
	ErrorCode    uint16
}
type userinput struct {
	Hardware  string
	SysName   string
	Copyright string
	Version   string
	ModType   string
	Mod       string
	SerialNum string
	PlantID   string
	CPUType   string
}
type SLZDataTree struct {
	Index  uint16
	MlfB   []byte
	BGType uint16
	Ausbg  uint16
	Ausbe  uint16
}
type Packet struct {
	TPKT TPKT
	COTP COTP
	S7   S7Packet
}
type ModInfo struct {
	SysName   []byte
	ModType   []byte
	PlantID   []byte
	Copyright []byte
	SerialNum []byte
	RSV       []byte
	CPUType   []byte
}
type S7CommPlus struct {
	ID uint8
	PDUType uint8
	DataLen uint16
	Reserved uint16
	SubType	uint16
	SecNum uint32
}

type S7ComPlusData struct{
	hostname string
	networkInt string
	dataType string
}

const (
	Request        = 0x01
	Ack            = 0x02
	AckData        = 0x03
	UserData       = 0x07
	S7ConReq       = 0xf0
	S7DataRequest  = 0x04
	S7DataResponse = 0x08
	CR             = 0xe0
	CC             = 0xd0
	COTPData       = 0xf0
)
