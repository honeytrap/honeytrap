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
package udp

import (
	"encoding/binary"
	"fmt"
)

type Header struct {
	Source      uint16
	Destination uint16
	Length      uint16
	Checksum    uint16
	Payload     []byte
}

func Unmarshal(data []byte) (*Header, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("Incorrect UDP header size: %d", len(data))
	}

	hdr := Header{}
	hdr.Source = binary.BigEndian.Uint16(data[0:2])
	hdr.Destination = binary.BigEndian.Uint16(data[2:4])
	hdr.Length = binary.BigEndian.Uint16(data[4:6])
	hdr.Checksum = binary.BigEndian.Uint16(data[6:8])
	hdr.Payload = data[8:]

	if len(data) != int(hdr.Length) {
		return nil, fmt.Errorf("UDP payload length and size doesn't match, got %d, expected %d", len(data), hdr.Length)
	}

	return &hdr, nil
}

func (hdr *Header) String() string {
	return fmt.Sprintf("sport=%d, sdest=%d, length=%d, checksum=%x",
		hdr.Source, hdr.Destination, hdr.Length, hdr.Checksum)
}

func (hdr *Header) Marshal() ([]byte, error) {
	buf := make([]byte, 8+len(hdr.Payload))
	binary.BigEndian.PutUint16(buf[0:2], hdr.Source)
	binary.BigEndian.PutUint16(buf[2:4], hdr.Destination)
	binary.BigEndian.PutUint16(buf[4:6], hdr.Length)
	binary.BigEndian.PutUint16(buf[6:8], hdr.Checksum)
	copy(buf[8:], hdr.Payload)
	return buf, nil
}
