/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (snmps://dutchsec.com/)
*
* This program is free software; you can snmptribute it and/or modify it under
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
* <snmp://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See snmps://honeytrap.io/ for more details. All requests should be sent to
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
package snmp

import (
	"bufio"
	"context"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("services/snmp")

var (
	_ = services.Register("snmp", SNMP)
)

func SNMP(options ...services.ServicerFunc) services.Servicer {
	s := &snmpService{
		limiter: services.NewLimiter(),
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type snmpService struct {
	limiter *services.Limiter
	c       pushers.Channel
}

func (s *snmpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func getOIDs(p Pdu) string {
	var oids []string
	for _, v := range p.Variables {
		oids = append(oids, v.Name.String())
	}
	return strings.Join(oids, ",")
}

func (s *snmpService) Handle(_ context.Context, conn net.Conn) error {
	if conn.RemoteAddr().Network() != "udp" {
		log.Errorf("SNMP is an UDP-only protocol (received %s data)", conn.RemoteAddr().Network())
		return nil
	}

	b := bufio.NewReader(conn)
	// Type + length
	hdr, err := b.Peek(2)
	if err != nil {
		return err
	}
	asnSize := 2 + int(hdr[1])
	buf := make([]byte, asnSize)
	n, err := b.Read(buf)
	if err != nil {
		return err
	}

	request := Message{}
	ctx := Asn1Context()
	remaining, err := ctx.Decode(buf, &request)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		log.Errorf("Invalid ASN.1: expected %d bytes, %d read, %d remaining", asnSize, n, b.Buffered())
		return nil
	}

	// SNMPv1 only for now
	if request.Version != 0 {
		log.Errorf("SNMP version not supported: %d", request.Version)
		s.c.Send(event.New(
			services.EventOptions,
			event.Category("snmp"),
			event.Type("unknown-packet"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("snmp.version", request.Version),
			event.Custom("snmp.community", request.Community),
			event.Payload(buf),
		))
		return nil
	}

	var packetType string
	var res GetResponsePdu
	var oids string
	switch pdu := request.Pdu.(type) {
	case GetRequestPdu:
		packetType = "get-request"
		oids = getOIDs(Pdu(pdu))
		res = processPdu(Pdu(pdu), false, false)
	case GetNextRequestPdu:
		packetType = "get-next-request"
		oids = getOIDs(Pdu(pdu))
		res = processPdu(Pdu(pdu), true, false)
	case SetRequestPdu:
		packetType = "set-request"
		oids = getOIDs(Pdu(pdu))
		/*
			if rw {
		*/
		res = processPdu(Pdu(pdu), false, true)
		/*
			} else {
				res = GetResponsePdu(pdu)
				res.ErrorIndex = 1
				res.ErrorStatus = NoSuchName
			}
		*/
	default:
		// SNMPv2 PDUs are ignored
		log.Errorf("Unsupported PDU: %T", request.Pdu)
		return nil
	}
	s.c.Send(event.New(
		services.EventOptions,
		event.Category("snmp"),
		event.Type(packetType),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("snmp.version", request.Version),
		event.Custom("snmp.community", request.Community),
		event.Custom("snmp.oids", oids),
		event.Payload(buf),
	))

	if !s.limiter.Allow(conn.RemoteAddr()) {
		return nil
	}

	response := request

	// Set response
	response.Pdu = res

	responseBytes, err := ctx.Encode(response)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = conn.Write(responseBytes)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func processPdu(pdu Pdu, next bool, set bool) GetResponsePdu {
	res := GetResponsePdu(pdu)
	// Return "No such name" for the first item
	res.ErrorIndex = 1
	res.ErrorStatus = NoSuchName
	return res
	/*
		// Keep returned values in a separated slice for a Get request
		var variables []Variable

		for i, v := range pdu.Variables {
			log.Debugf("oid: %s\n", v.Name)
			// Retrieve the managed object
			// h := ...
			// Set or get the value
			var value interface{}
			if set {
				err = h.set(h.oid, v.Value)
			} else {
				value, err = h.get(h.oid)
			}
			if !set {
				variables = append(variables, Variable{h.oid, value})
			}
		}
		if !set {
			// Update all values, since all variables were processed without error:
			res.Variables = variables
		}
		return res
	*/
}
