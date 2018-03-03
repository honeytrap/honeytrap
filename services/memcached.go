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
	"bufio"
	"context"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("memcached", Memcached)
)

func Memcached(options ...ServicerFunc) Servicer {
	s := &memcachedService{}

	for _, o := range options {
		o(s)
	}

	return s
}

type memcachedServiceConfig struct {
}

type memcachedService struct {
	memcachedServiceConfig

	ch pushers.Channel
}

func (s *memcachedService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *memcachedService) Handle(ctx context.Context, conn net.Conn) error {
	b := bufio.NewReader(conn)

	// memcached behaves differently over UDP: it has an 8-bytes header
	if conn.RemoteAddr().Network() == "udp" {
		_, err := b.Discard(8)
		if err != nil {
			log.Error("Error processing UDP header: %s", err.Error())
		}
	}

	for {
		command, err := b.ReadBytes('\n')
		if err != nil {
			break
		}
		// Strip trailing \r\n
		sz := len(command)
		if (sz >= 2) {
			command = command[:sz - 2]
		}

		s.ch.Send(event.New(
			EventOptions,
			event.Category("memcached"),
			event.Type("memcached-command"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("memcached.command", command),
		))

		conn.Write([]byte("ERROR\r\n"))
	}

	return nil
}
