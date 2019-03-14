package com

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

type ModInfo struct {
	SysName   []byte
	ModType   []byte
	PlantID   []byte
	Copyright []byte
	SerialNum []byte
	RSV       []byte
	CPUType   []byte
}
type Packet struct {
	TPKT TPKT
	COTP COTP
	S7   S7Packet
}

type TPKT struct {
	Version  uint8
	Reserved uint8
	Length   uint16
}

type S7Packet struct {
	Header    S7Header
	Parameter S7Parameter
	Data      S7Data
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

type S7CustomHead struct {
	ProtocolID  uint8
	MessageType uint8
	Reserved    uint16
	PDURef      uint16
	ParamLength uint16
	DataLength  uint16
}

type S7Parameter struct {
	SetupCom S7SetupCom
	UserData S7UserData
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

type S7DataNoSZL struct {
	ReturnCode    uint8
	TransportSize uint8
	Length        uint16
	SZLID         uint16
	SZLIndex      uint16
	SZLListLength uint16
	SZLListCount  uint16
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

type S7SetupCom struct {
	Function      uint8
	Reserved      uint8
	MaxAmQCalling uint16
	MaxAmQCalled  uint16
	PDULength     uint16
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

type SLZDataTree struct {
	Index  uint16
	MlfB   []byte
	BGType uint16
	Ausbg  uint16
	Ausbe  uint16
}

func (sd *S7Data) AddSLZDataTree(dt SLZDataTree) {
	sd.SZLDataTree = append(sd.SZLDataTree, dt)
}

type COTP struct {
	Length  uint8
	PDUType uint8
	DestRef uint8
}
