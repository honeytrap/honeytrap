/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ldap

import (
	"reflect"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
)

var (
	// SASL bind
	crammd5SASL = []byte{
		0x30, 0x16, // Begin SEQUENCE
		0x02, 0x01, 0x01, // Message ID
		0x60, 0x11, // bind request protocol op
		0x02, 0x01, 0x03, // LDAP version
		0x04, 0x00, // Empty bindDN
		0xa3, 0x0a, // Begin SASL auth
		0x04, 0x08, 0x43, 0x52, 0x41,
		0x4d, 0x2d, 0x4d, 0x44, 0x35, // SASL mechanism name 'CRAM-MD5'
	}

	// Bind request: cn=root,dc=example,dc=com password: root
	testRequest = []byte{
		0x30, 0x29, 0x02, 0x01, 0x01, 0x60, 0x24, 0x02,
		0x01, 0x03, 0x04, 0x19, 0x63, 0x6e, 0x3d, 0x72,
		0x6f, 0x6f, 0x74, 0x2c, 0x64, 0x63, 0x3d, 0x65,
		0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2c, 0x64,
		0x63, 0x3d, 0x63, 0x6f, 0x6d, 0x80, 0x04, 0x72,
		0x6f, 0x6f, 0x74,
	}
	//(501ff)succes message
	testResponse = []byte{
		0x30, 0x0c, 0x02, 0x01, 0xff, 0x61, 0x07, 0x0a,
		0x01, 0x00, 0x04, 0x00, 0x04, 0x00,
	}
	anonBind = []byte{
		0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02,
		0x01, 0x03, 0x04, 0x00, 0x80, 0x00,
	}
)

func TestHandle(t *testing.T) {
	cases := []struct {
		arg, want []byte
	}{
		{anonBind, testResponse},
	}

	h := &bindFuncHandler{
		bindFunc: func(name string, pw []byte) bool {
			// Just check the name
			if name == "root" || name == "" {
				return true
			}
			return false
		},
	}

	for _, c := range cases {
		p := ber.DecodePacket(c.arg)
		el := eventLog{}
		got := h.handle(p, el)

		if len(got) > 0 {
			if !reflect.DeepEqual(got[0].Bytes(), c.want) {
				t.Errorf("Bind: want %v got %v arg %v", c.want, got[0].Bytes(), p.Bytes())
			}
		} else {
			t.Errorf("Bind: want %v got nothing", c.want)
		}

		if rtype, ok := el["ldap.request-type"]; !ok || rtype != "bind" {
			t.Errorf("Bind: Wrong request type, want bind, got %s", rtype)
		}
	}
}

func TestHandleBad(t *testing.T) {
	cases := [][]byte{
		{},
		{0, 0, 0, 0, 0},
	}

	h := &bindFuncHandler{
		bindFunc: func(name string, pw []byte) bool {
			// Just check the name, so pw can be nil
			if name == "root" {
				return true
			}
			return false
		},
	}

	for _, c := range cases {
		el := make(eventLog)
		p := ber.DecodePacket(c)

		got := h.handle(p, el) // should return nil on bad input

		if got != nil {
			t.Errorf("Bind: No nil package on bad input. %v", p.Bytes())
		}

		// TODO: Check resultcode on bad/good creds
	}
}
