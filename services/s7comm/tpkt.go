/* Copyright 2016-2019 DutchSec (https://dutchsec.com/)
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package s7comm

import (
	"bytes"

	"github.com/honeytrap/honeytrap/services/decoder"
)

func (T *TPKT) serialize(m []byte) (r []byte) {

	T.Version = 0x03
	T.Reserved = 0x00
	T.Length = uint16(len(m) + 0x04)

	rb := &bytes.Buffer{}

	var eh errHandler

	eh.serializer(rb, T)
	eh.serializer(rb, m)

	if eh.err == nil {
		return rb.Bytes()
	}
	return nil
}

func (T *TPKT) deserialize(m *[]byte) (verified bool) {
	dec := decoder.NewDecoder(*m)

	T.Version = dec.Byte()
	T.Reserved = dec.Byte()
	T.Length = dec.Uint16()

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
