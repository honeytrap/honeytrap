/*
Copyright 2013-2014 Graham King

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

For full license details see <http://www.gnu.org/licenses/>.
*/

package tcp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type Flag uint8

const (
	FIN Flag = 1  // 00 0001
	SYN      = 2  // 00 0010
	RST      = 4  // 00 0100
	PSH      = 8  // 00 1000
	ACK      = 16 // 01 0000
	URG      = 32 // 10 0000
)

type Header struct {
	Source      uint16
	Destination uint16
	SeqNum      uint32
	AckNum      uint32
	DataOffset  uint8 // 4 bits
	Reserved    uint8 // 3 bits
	ECN         uint8 // 3 bits
	Ctrl        Flag  // 6 bits
	Window      uint16
	Checksum    uint16 // Kernel will set this if it's 0
	Urgent      uint16
	Options     []Option
	Padding     []byte
	Payload     []byte

	opts [4]Option
}

// TCPOptionKind represents a TCP option code.
type TCPOptionKind uint8

const (
	TCPOptionKindEndList                         = 0
	TCPOptionKindNop                             = 1
	TCPOptionKindMSS                             = 2  // len = 4
	TCPOptionKindWindowScale                     = 3  // len = 3
	TCPOptionKindSACKPermitted                   = 4  // len = 2
	TCPOptionKindSACK                            = 5  // len = n
	TCPOptionKindEcho                            = 6  // len = 6, obsolete
	TCPOptionKindEchoReply                       = 7  // len = 6, obsolete
	TCPOptionKindTimestamps                      = 8  // len = 10
	TCPOptionKindPartialOrderConnectionPermitted = 9  // len = 2, obsolete
	TCPOptionKindPartialOrderServiceProfile      = 10 // len = 3, obsolete
	TCPOptionKindCC                              = 11 // obsolete
	TCPOptionKindCCNew                           = 12 // obsolete
	TCPOptionKindCCEcho                          = 13 // obsolete
	TCPOptionKindAltChecksum                     = 14 // len = 3, obsolete
	TCPOptionKindAltChecksumData                 = 15 // len = n, obsolete
)

func (k TCPOptionKind) String() string {
	switch k {
	case TCPOptionKindEndList:
		return "EndList"
	case TCPOptionKindNop:
		return "NOP"
	case TCPOptionKindMSS:
		return "MSS"
	case TCPOptionKindWindowScale:
		return "WindowScale"
	case TCPOptionKindSACKPermitted:
		return "SACKPermitted"
	case TCPOptionKindSACK:
		return "SACK"
	case TCPOptionKindEcho:
		return "Echo"
	case TCPOptionKindEchoReply:
		return "EchoReply"
	case TCPOptionKindTimestamps:
		return "Timestamps"
	case TCPOptionKindPartialOrderConnectionPermitted:
		return "PartialOrderConnectionPermitted"
	case TCPOptionKindPartialOrderServiceProfile:
		return "PartialOrderServiceProfile"
	case TCPOptionKindCC:
		return "CC"
	case TCPOptionKindCCNew:
		return "CCNew"
	case TCPOptionKindCCEcho:
		return "CCEcho"
	case TCPOptionKindAltChecksum:
		return "AltChecksum"
	case TCPOptionKindAltChecksumData:
		return "AltChecksumData"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

type Option struct {
	OptionType   TCPOptionKind
	OptionLength uint8
	OptionData   []byte
}

func (t Option) String() string {
	hd := hex.EncodeToString(t.OptionData)
	if len(hd) > 0 {
		hd = " 0x" + hd
	}
	switch t.OptionType {
	case TCPOptionKindMSS:
		return fmt.Sprintf("Option(%s:%v%s)",
			t.OptionType,
			binary.BigEndian.Uint16(t.OptionData),
			hd)

	case TCPOptionKindTimestamps:
		if len(t.OptionData) == 8 {
			return fmt.Sprintf("Option(%s:%v/%v%s)",
				t.OptionType,
				binary.BigEndian.Uint32(t.OptionData[:4]),
				binary.BigEndian.Uint32(t.OptionData[4:8]),
				hd)
		}
	}
	return fmt.Sprintf("Option(%s:%s)", t.OptionType, hd)
}

var ErrInvalidChecksum = fmt.Errorf("Invalid checksum")

func UnmarshalWithChecksum(data []byte, src, dest net.IP) (*Header, error) {
	hdr := Header{}

	err := hdr.UnmarshalWithChecksum(data, src, dest)
	if err == ErrInvalidChecksum {
		return &hdr, err
	} else if err != nil {
		return nil, err
	}

	return &hdr, nil
}

func (hdr *Header) UnmarshalWithChecksum(data []byte, src, dest net.IP) error {
	err := hdr.Unmarshal(data)

	checksum := csum(data, to4byte(src.String()), to4byte(dest.String()))
	if checksum != hdr.Checksum {
		return ErrInvalidChecksum
	}

	return err
}

func Parse(data []byte) (Header, error) {
	h := Header{}
	return h, h.Unmarshal(data)
}

func (hdr *Header) String() string {
	return fmt.Sprintf("sport=%d, dport=%d, ctrl=%d, seqnum=%d, acknum=%d", hdr.Source, hdr.Destination, hdr.Ctrl, hdr.SeqNum, hdr.AckNum)
}

// why EOF on ubuntu with 22?
// https://github.com/google/gopacket/blob/master/layers/tcp.go<Paste>
func (hdr *Header) Unmarshal(data []byte) error {
	hdr.Source = binary.BigEndian.Uint16(data[0:2])
	hdr.Destination = binary.BigEndian.Uint16(data[2:4])
	hdr.SeqNum = binary.BigEndian.Uint32(data[4:8])
	hdr.AckNum = binary.BigEndian.Uint32(data[8:12])

	hdr.DataOffset = data[12] >> 4
	hdr.ECN = byte(data[13] >> 6 & 7)      // 3 bits
	hdr.Ctrl = Flag(byte(data[13] & 0x3f)) // bottom 6 bits

	hdr.Window = binary.BigEndian.Uint16(data[14:16])
	hdr.Checksum = binary.BigEndian.Uint16(data[16:18])
	hdr.Urgent = binary.BigEndian.Uint16(data[18:20])

	hdr.Options = hdr.opts[:0]

	if hdr.DataOffset < 5 {
		return fmt.Errorf("Invalid TCP data offset %d < 5", hdr.DataOffset)
	}

	dataStart := int(hdr.DataOffset) * 4
	if dataStart > len(data) {
		hdr.Payload = nil
		//hdr.Contents = data
		return errors.New("TCP data offset greater than packet length")
	}
	//hdr.Contents = data[:dataStart]
	hdr.Payload = data[dataStart:]
	// From here on, data points just to the header options.
	data = data[20:dataStart]
	for len(data) > 0 {
		if hdr.Options == nil {
			// Pre-allocate to avoid allocating a slice.
			hdr.Options = hdr.opts[:0]
		}
		hdr.Options = append(hdr.Options, Option{OptionType: TCPOptionKind(data[0])})
		opt := &hdr.Options[len(hdr.Options)-1]
		switch opt.OptionType {
		case TCPOptionKindEndList: // End of options
			opt.OptionLength = 1
			hdr.Padding = data[1:]
			break
		case TCPOptionKindNop: // 1 byte padding
			opt.OptionLength = 1
		default:
			opt.OptionLength = data[1]
			if opt.OptionLength < 2 {
				return fmt.Errorf("Invalid TCP option length %d < 2", opt.OptionLength)
			} else if int(opt.OptionLength) > len(data) {
				return fmt.Errorf("Ivalid TCP option length %d exceeds remaining %d bytes", opt.OptionLength, len(data))
			}
			opt.OptionData = data[2:opt.OptionLength]
		}
		data = data[opt.OptionLength:]
	}

	return nil
}

func (hdr *Header) HasFlag(flagBit Flag) bool {
	return hdr.Ctrl&flagBit == flagBit
}

func to4byte(addr string) [4]byte {
	parts := strings.Split(addr, ".")
	b0, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalf("to4byte: %s (latency works with IPv4 addresses only, but not IPv6!)\n", err)
	}
	b1, _ := strconv.Atoi(parts[1])
	b2, _ := strconv.Atoi(parts[2])
	b3, _ := strconv.Atoi(parts[3])
	return [4]byte{byte(b0), byte(b1), byte(b2), byte(b3)}
}

func (hdr *Header) CalcChecksum(src, dest net.IP) uint16 {
	return 0 //csum(data, to4byte(src.String()), to4byte(dest.String()))
}

func (hdr *Header) MarshalWithChecksum(src, dest net.IP) ([]byte, error) {
	data, err := hdr.Marshal()
	checksum := csum(data, to4byte(src.String()), to4byte(dest.String()))
	data[16] = byte(checksum >> 8)
	data[17] = byte(checksum & 0xFF)
	return data, err
}

var lotsOfZeros [1024]byte

func (t *Header) Marshal() ([]byte, error) {
	var optionLength int
	for _, o := range t.Options {
		switch o.OptionType {
		case 0, 1:
			optionLength += 1
		default:
			optionLength += 2 + len(o.OptionData)
		}
	}

	if rem := optionLength % 4; rem != 0 {
		t.Padding = lotsOfZeros[:4-rem]
	}

	t.DataOffset = uint8((len(t.Padding) + optionLength + 20) / 4)

	/*
		bytes, err := b.PrependBytes(20 + optionLength + len(t.Padding))
		if err != nil {
			return err
		}
	*/

	bytes := make([]byte, 20+optionLength+len(t.Padding)+len(t.Payload))
	copy(bytes[20+optionLength+len(t.Padding):], t.Payload)

	binary.BigEndian.PutUint16(bytes, uint16(t.Source))
	binary.BigEndian.PutUint16(bytes[2:], uint16(t.Destination))
	binary.BigEndian.PutUint32(bytes[4:], t.SeqNum)
	binary.BigEndian.PutUint32(bytes[8:], t.AckNum)

	bytes[12] = t.DataOffset << 4
	bytes[13] = ((t.ECN << 6) | uint8(t.Ctrl))
	binary.BigEndian.PutUint16(bytes[14:], t.Window)
	binary.BigEndian.PutUint16(bytes[18:], t.Urgent)

	start := 20
	for _, o := range t.Options {
		bytes[start] = byte(o.OptionType)
		switch o.OptionType {
		case 0, 1:
			start++
		default:
			o.OptionLength = uint8(len(o.OptionData) + 2)

			bytes[start+1] = o.OptionLength
			copy(bytes[start+2:start+len(o.OptionData)+2], o.OptionData)
			start += int(o.OptionLength)
		}
	}

	copy(bytes[start:], t.Padding)

	/*
		if /* opts.ComputeChecksums * true {
			// zero out checksum bytes in current serialization.
			bytes[16] = 0
			bytes[17] = 0
			csum, err := t.computeChecksum(b.Bytes(), IPProtocolTCP)
			if err != nil {
				return err
			}
			t.Checksum = csum
		}
		binary.BigEndian.PutUint16(bytes[16:], t.Checksum)
	*/
	return bytes, nil

}

// TCP Checksum
func csum(data []byte, srcip, dstip [4]byte) uint16 {
	csum := uint32(0)

	csum += (uint32(srcip[0]) << 8) + uint32(srcip[1])
	csum += (uint32(srcip[2]) << 8) + uint32(srcip[3])
	csum += (uint32(dstip[0]) << 8) + uint32(dstip[1])
	csum += (uint32(dstip[2]) << 8) + uint32(dstip[3])

	csum += uint32(6)

	length := uint32(len(data))
	csum += uint32(length)

	for i := uint32(0); i+1 < length; i += 2 {
		// skip checksum
		if i == 16 {
			continue
		}

		csum += uint32(uint16(data[i])<<8 + uint16(data[i+1]))
	}

	if len(data)%2 == 1 {
		csum += uint32(data[len(data)-1]) << 8
	}

	for csum>>16 > 0 {
		csum = (csum & 0xffff) + (csum >> 16)
	}

	return uint16(^csum)
}
