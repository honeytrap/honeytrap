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

// OptionKind represents a TCP option code.
type OptionKind uint8

const (
	optionKindEndList                         = 0
	optionKindNop                             = 1
	optionKindMSS                             = 2  // len = 4
	optionKindWindowScale                     = 3  // len = 3
	optionKindSACKPermitted                   = 4  // len = 2
	optionKindSACK                            = 5  // len = n
	optionKindEcho                            = 6  // len = 6, obsolete
	optionKindEchoReply                       = 7  // len = 6, obsolete
	optionKindTimestamps                      = 8  // len = 10
	optionKindPartialOrderConnectionPermitted = 9  // len = 2, obsolete
	optionKindPartialOrderServiceProfile      = 10 // len = 3, obsolete
	optionKindCC                              = 11 // obsolete
	optionKindCCNew                           = 12 // obsolete
	optionKindCCEcho                          = 13 // obsolete
	optionKindAltChecksum                     = 14 // len = 3, obsolete
	optionKindAltChecksumData                 = 15 // len = n, obsolete
)

func (k OptionKind) String() string {
	switch k {
	case optionKindEndList:
		return "EndList"
	case optionKindNop:
		return "NOP"
	case optionKindMSS:
		return "MSS"
	case optionKindWindowScale:
		return "WindowScale"
	case optionKindSACKPermitted:
		return "SACKPermitted"
	case optionKindSACK:
		return "SACK"
	case optionKindEcho:
		return "Echo"
	case optionKindEchoReply:
		return "EchoReply"
	case optionKindTimestamps:
		return "Timestamps"
	case optionKindPartialOrderConnectionPermitted:
		return "PartialOrderConnectionPermitted"
	case optionKindPartialOrderServiceProfile:
		return "PartialOrderServiceProfile"
	case optionKindCC:
		return "CC"
	case optionKindCCNew:
		return "CCNew"
	case optionKindCCEcho:
		return "CCEcho"
	case optionKindAltChecksum:
		return "AltChecksum"
	case optionKindAltChecksumData:
		return "AltChecksumData"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

type Option struct {
	OptionType   OptionKind
	OptionLength uint8
	OptionData   []byte
}

func (t Option) String() string {
	hd := hex.EncodeToString(t.OptionData)
	if len(hd) > 0 {
		hd = " 0x" + hd
	}
	switch t.OptionType {
	case optionKindMSS:
		return fmt.Sprintf("Option(%s:%v%s)",
			t.OptionType,
			binary.BigEndian.Uint16(t.OptionData),
			hd)

	case optionKindTimestamps:
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
Loop:
	for len(data) > 0 {
		if hdr.Options == nil {
			// Pre-allocate to avoid allocating a slice.
			hdr.Options = hdr.opts[:0]
		}
		hdr.Options = append(hdr.Options, Option{OptionType: OptionKind(data[0])})
		opt := &hdr.Options[len(hdr.Options)-1]
		switch opt.OptionType {
		case optionKindEndList: // End of options
			opt.OptionLength = 1
			hdr.Padding = data[1:]
			break Loop
		case optionKindNop: // 1 byte padding
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

func (hdr *Header) Marshal() ([]byte, error) {
	var optionLength int
	for _, o := range hdr.Options {
		switch o.OptionType {
		case 0, 1:
			optionLength++
		default:
			optionLength += 2 + len(o.OptionData)
		}
	}

	if rem := optionLength % 4; rem != 0 {
		hdr.Padding = lotsOfZeros[:4-rem]
	}

	hdr.DataOffset = uint8((len(hdr.Padding) + optionLength + 20) / 4)

	/*
		bytes, err := b.PrependBytes(20 + optionLength + len(hdr.Padding))
		if err != nil {
			return err
		}
	*/

	bytes := make([]byte, 20+optionLength+len(hdr.Padding)+len(hdr.Payload))
	copy(bytes[20+optionLength+len(hdr.Padding):], hdr.Payload)

	binary.BigEndian.PutUint16(bytes, uint16(hdr.Source))
	binary.BigEndian.PutUint16(bytes[2:], uint16(hdr.Destination))
	binary.BigEndian.PutUint32(bytes[4:], hdr.SeqNum)
	binary.BigEndian.PutUint32(bytes[8:], hdr.AckNum)

	bytes[12] = hdr.DataOffset << 4
	bytes[13] = ((hdr.ECN << 6) | uint8(hdr.Ctrl))
	binary.BigEndian.PutUint16(bytes[14:], hdr.Window)
	binary.BigEndian.PutUint16(bytes[18:], hdr.Urgent)

	start := 20
	for _, o := range hdr.Options {
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

	copy(bytes[start:], hdr.Padding)

	/*
		if /* opts.ComputeChecksums * true {
			// zero out checksum bytes in current serialization.
			bytes[16] = 0
			bytes[17] = 0
			csum, err := hdr.computeChecksum(b.Bytes(), IPProtocolTCP)
			if err != nil {
				return err
			}
			hdr.Checksum = csum
		}
		binary.BigEndian.PutUint16(bytes[16:], hdr.Checksum)
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
