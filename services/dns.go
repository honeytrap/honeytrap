/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
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
package services

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/miekg/dns"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("dns", DNS)
)

// Dns is a placeholder
func DNS(options ...ServicerFunc) Servicer {
	s := &dnsService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type dnsService struct {
	c pushers.Channel
}

func (s *dnsService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *dnsService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 65535)

	if _, ok := conn.(*listener.DummyUDPConn); ok {
		n, err := conn.Read(buff[:])
		if err != nil {
			return err
		}

		buff = buff[:n]
	} else if _, ok := conn.(*net.TCPConn); ok {
		n, err := conn.Read(buff[:])
		if err != nil {
			return err
		}

		buff = buff[:n]
	} else {
		log.Error("Unsupported connection type: %s", reflect.TypeOf(conn))
		return nil
	}

	req := new(dns.Msg)
	if err := req.Unpack(buff[:]); err != nil {
		return err
	}

	s.c.Send(event.New(
		EventOptions,
		event.Category("dns"),
		event.Type("dns"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("dns.id", fmt.Sprintf("%d", req.Id)),
		event.Custom("dns.opcode", fmt.Sprintf("%d", req.Opcode)),
		event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", req.Question)),
		event.Custom("dns.questions", req.Question),
	))

	return nil
}
