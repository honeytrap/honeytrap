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
