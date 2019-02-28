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
	"reflect"
	"testing"
)

func TestTPKTValidation(t *testing.T) {
	var T = TPKT{
		Version:  0x03,
		Reserved: 0x00,
		Length:   0x03,
	}

	m := []byte{0x03, 0x00, 0x03}

	if !T.verify(m) {
		t.Errorf("TPKT validation check failed. Values are: %d %d %d", T.Version, T.Reserved, T.Length)
	}
}

func TestTPKTSerialize(t *testing.T) {
	var T TPKT
	m := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	s := T.serialize(m)

	rCk := []byte{0x03, 0x00, 0x00, 0x09, 0x01, 0x02, 0x03, 0x04, 0x05}

	if !reflect.DeepEqual(s, rCk) {
		t.Errorf("TPKT serialization test failed")
	}
}

func TestTPKTDeserialize(t *testing.T) {
	var T TPKT
	m := []byte{0x03, 0x00, 0x00, 0x09, 0x01, 0x02, 0x03, 0x04, 0x05}

	if !T.deserialize(&m) {
		t.Errorf("TPKT deserialization failed")
	}

	m = []byte{0x03, 0x00, 0x00, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05}
	if T.deserialize(&m) {
		t.Errorf("Invalid TPKT packet has been marked as valid")
	}
}

func TestCOTPSerialize(t *testing.T) {
	var C COTP
	m := []byte{0x00, 0x01, 0x02, 0x03}
	r := C.serialize(m)
	rCk := []byte{0x02, 0xf0, 0x80, 0x00, 0x01, 0x02, 0x03}

	if !reflect.DeepEqual(r, rCk) {
		t.Errorf("COTP serialization test failed")
	}

}
func TestCOTPDeserialize(t *testing.T) {
	var C COTP
	m := []byte{0x02, 0xf0, 0x80, 0x00, 0x01, 0x02, 0x03}

	if !C.deserialize(&m) {
		t.Errorf("COTP deserialization test failed")
	}

	bm := []byte{0x02, 0x03, 0x80, 0x00, 0x01, 0x02, 0x03}
	if C.deserialize(&bm) {
		t.Errorf("Broken COTP PDU marked as valid.")
	}
}

func TestCreateCOTPCon(t *testing.T) {
	var C COTP

	m := []byte{0x03, 0x00, 0x00, 0x16, 0x11, 0xe0, 0x00, 0x00, 0x00, 0x14, 0x00, 0xc1, 0x02, 0x01, 0x00, 0xc2, 0x02, 0x01, 0x02, 0xc0, 0x01, 0x0a}
	r := C.connect(m)
	if r == nil {
		t.Errorf("Connection Request get bad response")

	}

	er := []byte{0x03, 0x00, 0x00, 0x16, 0x11, 0xd0, 0x00, 0x14, 0x00, 0x00, 0x00, 0xc0, 0x01, 0x0a, 0xc1, 0x02, 0x01, 0x00, 0xc2, 0x02, 0x01, 0x02}
	if !reflect.DeepEqual(r, er) {
		t.Errorf("COTP handshake test failed")
	}

}

func TestS7Deserialize(t *testing.T) {
	var s S7Packet
	m := []byte{0x03, 0x00, 0x00, 0x21, 0x02, 0xf0, 0x80, 0x32, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x08, 0x00, 0x01, 0x12, 0x04, 0x11, 0x44, 0x01, 0x00, 0xff, 0x09, 0x00, 0x04, 0x00, 0x11, 0x00, 0x01}
	_, S7 := s.deserialize(m)

	if !S7 {
		t.Errorf("S7 packet not recognized as S7. Value returned: %v", S7)
	}
}
