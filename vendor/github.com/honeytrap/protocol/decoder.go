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
package protocol

import (
	"bufio"
	"encoding/binary"
	"io"
)

func NewDecoder(r io.Reader, bo binary.ByteOrder) *Decoder {
	return &Decoder{
		Reader:    bufio.NewReader(r),
		bo:        bo,
		LastError: nil,
	}
}

type Decoder struct {
	*bufio.Reader

	bo        binary.ByteOrder
	LastError error
}

func (d *Decoder) ReadUint8() int {
	if d.LastError != nil {
		return 0
	}

	b, err := d.ReadByte()
	if err != nil {
		d.LastError = err
	}

	return int(b)
}

func (d *Decoder) ReadUint16() int {
	if d.LastError != nil {
		return 0
	}

	buffer := [2]byte{}
	if _, err := d.Read(buffer[:]); err != nil {
		d.LastError = err
		return 0
	}

	return int(d.bo.Uint16(buffer[:]))
}
