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
	"io"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("copy", Copy)
)

// Copy is a placeholder
func Copy(options ...ServicerFunc) Servicer {
	s := &copyService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type copyService struct {
	c pushers.Channel

	d director.Director
}

func (s *copyService) SetDirector(d director.Director) {
	s.d = d
}

func (s *copyService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *copyService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()
	switch conn.(type) {
	case *listener.DummyUDPConn:
		defer s.c.Send(event.New(
			EventOptions,
			event.Category("copy"),
			event.Type("tcp"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
		))

		conn2, err := s.d.Dial(conn)
		if err != nil {
			return err
		}

		defer conn2.Close()

		go io.Copy(conn2, conn)
		_, err = io.Copy(conn, conn2)

		return err
	case *net.TCPConn:
		defer s.c.Send(event.New(
			EventOptions,
			event.Category("copy"),
			event.Type("udp"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
		))

		conn2, err := s.d.Dial(conn)
		if err != nil {
			return err
		}

		defer conn2.Close()

		go io.Copy(conn2, conn)
		_, err = io.Copy(conn, conn2)
		return err
	default:
		return nil
	}
}
