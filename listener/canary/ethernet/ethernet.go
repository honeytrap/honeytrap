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
package ethernet

import (
	"encoding/binary"
	"net"
	"syscall"
)

type EthernetFrame struct {
	Source      net.HardwareAddr
	Destination net.HardwareAddr

	Type uint16

	Payload []byte
}

func Parse(data []byte) (*EthernetFrame, error) {
	eh := &EthernetFrame{
		Source:      make([]byte, 6),
		Destination: make([]byte, 6),
	}
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
