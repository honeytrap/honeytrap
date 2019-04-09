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

// Handle simple bind requests

import (
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
)

//bindFunc checks simple auth credentials (username/password style)
type bindFunc func(binddn string, bindpw []byte) bool

//bindFuncHandler: responds to bind requests
type bindFuncHandler struct {
	bindFunc bindFunc
}

func (h *bindFuncHandler) handle(p *ber.Packet, el eventLog) []*ber.Packet {
	reth := &resultCodeHandler{replyTypeID: AppBindResponse, resultCode: ResInvalidCred}

	// check for bind request contents
	if p == nil || len(p.Children) < 2 {
		// Package is not meant for us
		return nil
	}
	err := checkPacket(p.Children[1], ber.ClassApplication, ber.TypeConstructed, AppBindRequest)
	if err != nil {
		// Package is not meant for us
		return nil
	}

	// If we are here we have a bind request
	el["ldap.request-type"] = "bind"

	version := readVersion(p)
	el["ldap.version"] = version

	if version < 2 {
		reth.resultCode = ResProtocolError
		return reth.handle(p, el)
	}

	// make sure we have at least our version number, bind dn and bind password
	if len(p.Children[1].Children) < 3 {
		el["ldap.malformed-payload"] = p.Data.Bytes()
		log.Debugf("At least 3 elements required in bind request, found %v", len(p.Children[1].Children))
		return nil
	}

	// the bind DN (the "username")
	err = checkPacket(p.Children[1].Children[1], ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString)
	if err != nil {
		el["ldap.malformed-payload"] = p.Data.Bytes()
		log.Debugf("Error verifying packet: %v", err)
		return nil
	}

	bindDn := string(p.Children[1].Children[1].ByteValue)

	if index := strings.Index(bindDn, ","); index > -1 {
		bindDn = bindDn[:index]
	}

	if strings.HasPrefix(bindDn, "cn=") || strings.HasPrefix(bindDn, "sn=") {
		bindDn = bindDn[3:]
	}

	el["ldap.username"] = bindDn

	err = checkPacket(p.Children[1].Children[2], ber.ClassContext, ber.TypePrimitive, 0x0)
	if err != nil {
		el["ldap.malformed-payload"] = p.Data.Bytes()
		log.Debugf("Error verifying packet: %v", err)
		return nil
	}

	bindPw := p.Children[1].Children[2].Data.Bytes()

	if len(bindPw) == 0 && bindDn != "" {
		// username without password (rfc4513)
		reth.resultCode = ResUnwillingToPerform
	}

	el["ldap.password"] = string(bindPw)

	// call back to the auth handler
	if h.bindFunc(bindDn, bindPw) {
		// it worked, result code should be zero for success
		reth.resultCode = ResSuccess
	}

	return reth.handle(p, el)
}
