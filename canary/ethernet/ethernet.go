package ethernet

import (
	"encoding/binary"
	"net"
	"syscall"
)

type EthernetFrame struct {
	Source      net.HardwareAddr
	Destination net.HardwareAddr
	Type        uint16

	Payload []byte
}

func Parse(data []byte) (*EthernetFrame, error) {
	eh := &EthernetFrame{}
	return eh, eh.Unmarshal(data)
}

func (eh *EthernetFrame) Unmarshal(data []byte) error {
	copy(eh.Destination[:], data[0:6])
	copy(eh.Source[:], data[6:12])
	eh.Type = binary.BigEndian.Uint16(data[12:14])
	eh.Payload = data[14:]
	return nil
}

// Marshal returns the binary encoding of the IPv4 header h.
func (ef *EthernetFrame) Marshal() ([]byte, error) {
	if ef == nil {
		return nil, syscall.EINVAL
	}

	data := [14]byte{}
	copy(data[0:6], ef.Destination)
	copy(data[6:12], ef.Source)
	data[12] = uint8((ef.Type >> 8) & 0xFF)
	data[13] = uint8(ef.Type & 0xFF)
	return data[:], nil
}
