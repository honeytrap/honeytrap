package arp

import (
	"encoding/binary"
	"fmt"
)

type ARPFrame struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	Opcode       uint16

	SenderMAC [6]byte
	SenderIP  [4]byte

	TargetMAC [6]byte
	TargetIP  [4]byte
}

func (a *ARPFrame) String() string {
	return fmt.Sprintf("HardwareType: %x, ProtocolType: %x, HardwareSize: %x, ProtocolSize: %x, Opcode: %x, SenderMAC: %#v, SenderIP: %#v, TargetMAC: %#v, TargetIP: %#v",
		a.HardwareType, a.ProtocolType, a.HardwareSize, a.ProtocolSize, a.Opcode, a.SenderMAC, a.SenderIP, a.TargetMAC, a.TargetIP)
}

func Parse(data []byte) (*ARPFrame, error) {
	eh := &ARPFrame{}
	return eh, eh.Unmarshal(data)
}

func (eh *ARPFrame) Unmarshal(data []byte) error {
	eh.HardwareType = binary.BigEndian.Uint16(data[0:2])
	eh.ProtocolType = binary.BigEndian.Uint16(data[2:4])
	eh.HardwareSize = data[4]
	eh.ProtocolSize = data[5]
	eh.Opcode = binary.BigEndian.Uint16(data[6:8])

	copy(eh.SenderMAC[:], data[8:14])
	copy(eh.SenderIP[:], data[14:18])
	copy(eh.TargetMAC[:], data[18:24])
	copy(eh.TargetIP[:], data[24:28])

	return nil
}
