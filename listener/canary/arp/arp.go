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
package arp

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ARP_OPC_RESERVED          ArpOpcode = 0
	ARP_OPC_REQUEST                     = 1
	ARP_OPC_REPLY                       = 2
	ARP_OPC_REQUEST_R                   = 3
	ARP_OPC_REPLY_R                     = 4
	ARP_OPC_DRAPP_REQUEST               = 5
	ARP_OPC_DRAPP_REPLY                 = 6
	ARP_OPC_DRAPP_ERROR                 = 7
	ARP_OPC_INAPP_REQUEST               = 8
	ARP_OPC_INAPP_REPLY                 = 9
	ARP_OPC_ARP_NAK                     = 10
	ARP_OPC_MARS_REQUEST                = 11
	ARP_OPC_MARS_MULTI                  = 12
	ARP_OPC_MARS_MSERV                  = 13
	ARP_OPC_MARS_JOIN                   = 14
	ARP_OPC_MARS_LEAVE                  = 15
	ARP_OPC_MARS_NAK                    = 16
	ARP_OPC_MARS_UNSERV                 = 17
	ARP_OPC_MARS_SJOIN                  = 18
	ARP_OPC_MARS_SLEAVE                 = 19
	ARP_OPC_MARS_GL_REQUEST             = 20
	ARP_OPC_MARS_GL_REPLY               = 21
	ARP_OPC_MARS_REDIRECT_MAP           = 22
	ARP_OPC_MAPOS_UNARP                 = 23
	ARP_OPC_OP_EXP1                     = 24
	ARP_OPC_OP_EXP2                     = 25
)

type ArpOpcode uint16

func (o ArpOpcode) String() string {
	switch o {
	case ARP_OPC_RESERVED:
		return "reserved"
	case ARP_OPC_REQUEST:
		return "request"
	case ARP_OPC_REPLY:
		return "reply"
	case ARP_OPC_REQUEST_R:
		return "request-response"
	case ARP_OPC_REPLY_R:
		return "reply-response"
	case ARP_OPC_DRAPP_REQUEST:
		return "drapp-request"
	case ARP_OPC_DRAPP_REPLY:
		return "drapp-reply"
	case ARP_OPC_DRAPP_ERROR:
		return "drapp-error"
	case ARP_OPC_INAPP_REQUEST:
		return "inapp-request"
	case ARP_OPC_INAPP_REPLY:
		return "inapp-reply"
	case ARP_OPC_ARP_NAK:
		return "nak"
	case ARP_OPC_MARS_REQUEST:
		return "mars-request"
	case ARP_OPC_MARS_MULTI:
		return "mars-multi"
	case ARP_OPC_MARS_MSERV:
		return "mars-mserv"
	case ARP_OPC_MARS_JOIN:
		return "mars-join"
	case ARP_OPC_MARS_LEAVE:
		return "mars-leave"
	case ARP_OPC_MARS_NAK:
		return "mars-nak"
	case ARP_OPC_MARS_UNSERV:
		return "mars-unserv"
	case ARP_OPC_MARS_SJOIN:
		return "mars-sjoin"
	case ARP_OPC_MARS_SLEAVE:
		return "mars-sleave"
	case ARP_OPC_MARS_GL_REQUEST:
		return "mars-gl-request"
	case ARP_OPC_MARS_GL_REPLY:
		return "mars-gl-reply"
	case ARP_OPC_MARS_REDIRECT_MAP:
		return "mars-redirect-map"
	case ARP_OPC_MAPOS_UNARP:
		return "mapos-unarp"
	case ARP_OPC_OP_EXP1:
		return "op-exp1"
	case ARP_OPC_OP_EXP2:
		return "op-exp2"
	default:
		return fmt.Sprintf("unknown opcode %d", uint16(o))
	}
}

type Frame struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareSize uint8
	ProtocolSize uint8
	Opcode       uint16

	SenderMAC net.HardwareAddr
	SenderIP  net.IP

	TargetMAC net.HardwareAddr
	TargetIP  net.IP

	SenderHardwareAddress []byte
	SenderProtocolAddress []byte

	TargetHardwareAddress []byte
	TargetProtocolAddress []byte
}

func (f *Frame) String() string {
	return fmt.Sprintf("HardwareType: %x, ProtocolType: %x, HardwareSize: %x, ProtocolSize: %x, Opcode: %x, SenderMAC: %#v, SenderIP: %#v, TargetMAC: %#v, TargetIP: %#v",
		f.HardwareType, f.ProtocolType, f.HardwareSize, f.ProtocolSize, f.Opcode, f.SenderMAC, f.SenderIP, f.TargetMAC, f.TargetIP)
}

func Parse(data []byte) (*Frame, error) {
	eh := &Frame{}
	return eh, eh.Unmarshal(data)
}

func (f *Frame) Unmarshal(data []byte) error {
	if len(data) < 28 {
		return fmt.Errorf("Incorrect ARP header size: %d", len(data))
	}

	f.HardwareType = binary.BigEndian.Uint16(data[0:2])
	f.ProtocolType = binary.BigEndian.Uint16(data[2:4])
	f.HardwareSize = data[4]
	f.ProtocolSize = data[5]
	f.Opcode = binary.BigEndian.Uint16(data[6:8])

	if f.HardwareSize > 20 {
		return fmt.Errorf("Oversized ARP hardware size: %d", f.HardwareSize)
	}

	if f.ProtocolSize > 20 {
		return fmt.Errorf("Oversized ARP protocol size: %d", f.ProtocolSize)
	}

	data = data[8:]

	f.SenderHardwareAddress = make([]byte, f.HardwareSize)
	copy(f.SenderHardwareAddress[:], data[:f.HardwareSize])

	data = data[f.HardwareSize:]

	f.SenderProtocolAddress = make([]byte, f.ProtocolSize)
	copy(f.SenderProtocolAddress[:], data[:f.ProtocolSize])

	data = data[f.ProtocolSize:]

	f.TargetHardwareAddress = make([]byte, f.HardwareSize)
	copy(f.TargetHardwareAddress[:], data[:f.HardwareSize])

	data = data[f.HardwareSize:]

	f.TargetProtocolAddress = make([]byte, f.ProtocolSize)
	copy(f.TargetProtocolAddress[:], data[:f.ProtocolSize])

	data = data[f.ProtocolSize:]

	if f.HardwareSize == 6 {
		f.SenderMAC = net.HardwareAddr(f.SenderHardwareAddress)
		f.TargetMAC = net.HardwareAddr(f.TargetHardwareAddress)
	}

	if f.ProtocolType == 2048 && f.ProtocolSize == 4 {
		f.SenderIP = net.IP(f.SenderProtocolAddress)
		f.TargetIP = net.IP(f.TargetProtocolAddress)
	}

	return nil
}
