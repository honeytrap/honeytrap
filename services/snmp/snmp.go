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
package snmp

import (
	"bufio"
	"context"
	"fmt"
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
		res = processPdu(Pdu(pdu), false, true)
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
		log.Warningf("Rate limit exceeded for host: %s", conn.RemoteAddr())
		return nil
	}

	response := request

	// Set response
	response.Pdu = res

	responseBytes, err := ctx.Encode(response)
	if err != nil {
		return err
	}

	_, err = conn.Write(responseBytes)
	if err != nil {
		return fmt.Errorf("Error writing response: %s: %s", conn.RemoteAddr().String(), err.Error())
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
