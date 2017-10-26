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
