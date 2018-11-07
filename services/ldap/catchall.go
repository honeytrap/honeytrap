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

import ber "github.com/go-asn1-ber/asn1-ber"

//CatchAll handles the not implemented LDAP requests
type CatchAll struct {
	isLogin func() bool
}

func (c *CatchAll) handle(p *ber.Packet, el eventLog) []*ber.Packet {

	if p == nil || len(p.Children) < 2 {
		return nil
	}

	el["ldap.payload"] = p.Bytes()

	opcode := int(p.Children[1].Tag)

	reth := &resultCodeHandler{
		resultCode: ResSuccess,
	}

	if !c.isLogin() {
		// Not authenticated
		reth.resultCode = ResUnwillingToPerform
	}

	switch opcode {
	case AppModifyRequest:
		el["ldap.request-type"] = "modify"
		reth.replyTypeID = AppModifyResponse
	case AppAddRequest:
		el["ldap.request-type"] = "add"
		reth.replyTypeID = AppAddResponse
	case AppDelRequest:
		el["ldap.request-type"] = "delete"
		reth.replyTypeID = AppDelResponse
	case AppModifyDNRequest:
		el["ldap.request-type"] = "modify-dn"
		reth.replyTypeID = AppModifyDNResponse
	case AppCompareRequest:
		el["ldap.request-type"] = "compare"
		reth.replyTypeID = AppCompareResponse
	case AppAbandonRequest:
		el["ldap.request-type"] = "abandon"
		return nil // This needs no answer
	default:
		//el["ldap.request-type"] = opcode
		//reth.replyTypeID = 1 // protocolError
	}
	return reth.handle(p, el)
}
