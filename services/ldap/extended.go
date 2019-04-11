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
	ber "github.com/go-asn1-ber/asn1-ber"
)

type tlsFunc func() error

type extFuncHandler struct {
	tlsFunc tlsFunc
}

var tlsOID = "1.3.6.1.4.1.1466.20037"

func (e *extFuncHandler) handle(p *ber.Packet, el eventLog) []*ber.Packet {
	reth := &resultCodeHandler{replyTypeID: AppExtendedResponse, resultCode: ResProtocolError}

	// too small to be an extended request
	if p == nil || len(p.Children) < 2 {
		return nil
	}

	// check if package is an extended request
	err := checkPacket(p.Children[1], ber.ClassApplication, ber.TypeConstructed, AppExtendedRequest)
	if err != nil {
		return nil
	}

	el["ldap.request-type"] = "extended"

	// is there an OID and optional value
	if len(p.Children[1].Children) > 0 {

		err = checkPacket(p.Children[1].Children[0], ber.ClassContext, ber.TypePrimitive, 0)
		if err != nil {
			// this is not an OID
			return nil
		}

		oid := p.Children[1].Children[0].Data.String()

		el["ldap.extended-oid"] = oid

		// catch value if we have one
		if len(p.Children[1].Children) > 1 {
			el["ldap.extended-oid-value"] = p.Children[1].Children[1].Bytes()
		}

	}

	if tlsOID == el["ldap.extended-oid"].(string) {
		el["ldap.request-type"] = "extended.tls"

		if err := e.tlsFunc(); err == nil {
			reth.resultCode = ResSuccess
			reth.matchedDN = ""
			reth.diagnosticMsg = ""
		}
	}

	return reth.handle(p, el)
}
