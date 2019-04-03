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

package s7comm

import (
	"bytes"
	"encoding/binary"
)

func (T *TPKT) serialize(m []byte) (r []byte) {

	T.Version = 0x03
	T.Reserved = 0x00
	T.Length = uint16(len(m) + 0x04)

	rb := &bytes.Buffer{}

	TErr := binary.Write(rb, binary.BigEndian, T)
	mErr := binary.Write(rb, binary.BigEndian, m)

	if TErr != nil || mErr != nil {
		return nil
	}
	return rb.Bytes()
}

func (T *TPKT) deserialize(m *[]byte) (verified bool) {
	Length := binary.BigEndian.Uint16((*m)[2:4])
	T.Version = (*m)[0]
	T.Reserved = (*m)[1]
	T.Length = Length

	if T.verify(*m) {
		*m = (*m)[4:]
		return true
	}
	return false
}

func (T *TPKT) verify(m []byte) (isTPKT bool) {
	if T.Version == 0x03 && T.Reserved == 0x00 && int(T.Length)-len(m) == 0 {
		return true
	}
	return false

}

type TPKT struct {
	Version  uint8
	Reserved uint8
	Length   uint16
}
