// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ldap

import (
	"reflect"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
)

var (
	bindRequest = []byte{
		0x30, 0x14,
		0x02, 0x01, 0x01, // message ID(1)
		0x60, 0x0f, // begin bind request
		0x02, 0x01, 0x03, // LDAP version
		0x04, 0x04, 0x72, 0x6f, 0x6f, 0x74, // bind DN("root")
		0x80, 0x04, 0x72, 0x6f, 0x6f, 0x74, // password("root")
	}
	bindRequest2 = []byte{
		0x30, 0x39, // Begin the LDAPMessage sequence
		0x02, 0x01, 0x01, // The message ID (integer value 1)
		0x60, 0x34, // Begin the bind request protocol op
		0x02, 0x01, 0x03, // The LDAP protocol version (integer value 3)
		0x04, 0x24, 0x75, 0x69, 0x64, 0x3d, 0x6a, 0x64, // The bind DN (36-byte octet string "uid=jdoe,ou=People,dc=example,dc=com")
		0x65, 0x6f, 0x2c, 0x6f, 0x75, 0x3d, 0x50, 0x65,
		0x6f, 0x70, 0x6c, 0x65, 0x2c, 0x64, 0x63, 0x3d,
		0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2c,
		0x64, 0x63, 0x3d, 0x63, 0x6f, 0x6d,
		0x80, 0x09, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, // The password 9-byte octet string "secret123"
		0x31, 0x32, 0x33,
	}
	bindRequest_anon = []byte{
		0x30, 0x0c,
		0x02, 0x01, 0x01, // message ID(1)
		0x60, 0x07, // begin bind request
		0x02, 0x01, 0x03, // LDAP version
		0x04, 0x00, // empty bind DN
		0x80, 0x00, // empty password
	}
	bindRequest_nopassw = []byte{
		0x30, 0x10,
		0x02, 0x01, 0x01, // message ID(1)
		0x60, 0x0b, // begin bind request
		0x02, 0x01, 0x03, // LDAP version
		0x04, 0x04, 0x72, 0x6f, 0x6f, 0x74, // bind DN("root")
		0x80, 0x00, // empty password
	}
	/*
		bindRequest_crammd5 = []byte{
			0x30, 0x16, // Begin the LDAPMessage sequence
			0x02, 0x01, 0x01, // The message ID (integer value 1)
			0x60, 0x11, // Begin the bind request protocol op
			0x02, 0x01, 0x03, // The LDAP protocol version (integer value 3)
			0x04, 0x00, // Empty bind DN (0-byte octet string)
			0xa3, 0x0a, // Begin the SASL authentication sequence
			0x04, 0x08, 0x43, 0x52, 0x41, 0x4d, // The SASL mechanism name, (the octet string "CRAM-MD5")
			0x2d, 0x4d, 0x44, 0x35,
		}
	*/
	bindResponse_succes = []byte{
		0x30, 0x0c, // Begin the LDAPMessage sequence
		0x02, 0x01, 0x01, // The message ID (integer value 1)
		0x61, 0x07, // Begin the bind response protocol op
		0x0a, 0x01, 0x00, // success result code (enumerated value 0)
		0x04, 0x00, // No matched DN (0-byte octet string)
		0x04, 0x00, // No diagnostic message (0-byte octet string)
	}
	bindResponse_fail = []byte{
		0x30, 0x0c, // Begin the LDAPMessage sequence
		0x02, 0x01, 0x01, // The message ID (integer value 1)
		0x61, 0x07, // Begin the bind response protocol op
		0x0a, 0x01, 0x31, // bind fail result code (enumerated value 49)
		0x04, 0x00, // No matched DN (0-byte octet string)
		0x04, 0x00, // No diagnostic message (0-byte octet string)
	}
)

func TestHandle(t *testing.T) {
	cases := []struct {
		arg, want []byte
	}{
		{bindRequest_anon, bindResponse_fail},
		{bindRequest2, bindResponse_fail},
		{bindRequest, bindResponse_succes},
		{bindRequest_nopassw, bindResponse_succes},
	}

	h := &bindFuncHandler{
		bindFunc: func(name string, pw []byte) bool {
			if name == "root" {
				return true
			}
			return false
		},
	}

	for _, c := range cases {
		p, err := ber.DecodePacketErr(c.arg)
		if err != nil {
			t.Errorf("Bind: DecodePacket(%#v) returns error: %s", c.arg, err)
		}

		el := make(eventLog)
		got := h.handle(p, el)

		if len(got) > 0 {
			if !reflect.DeepEqual(got[0].Bytes(), c.want) {
				t.Errorf("Bind: DecodePacket(%#v)\nwant %v\ngot %v", c.arg, c.want, got[0].Bytes())
			}
		} else {
			t.Errorf("Bind: DecodePacket(%#v)\nwant %v got nothing", c.arg, c.want)
		}

		if rtype, ok := el["ldap.request-type"]; ok && rtype != "bind" {
			t.Errorf("Bind: DecodePacket(%#v) Wrong request type, want bind, got %s", c.arg, rtype)
		} else if !ok {
			t.Errorf("Bind: DecodePacket(%#v) No ldap-request-type", c.arg)
		}
	}
}
