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
