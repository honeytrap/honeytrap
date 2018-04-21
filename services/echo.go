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
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"io"
)

var (
	_ = Register("echo", Echo)
)

// Echo is a placeholder
func Echo(options ...ServicerFunc) Servicer {
	s := &echoService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type echoService struct {
	c pushers.Channel
}

func (s *echoService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *echoService) Handle(ctx context.Context, conn net.Conn) error {
	if _, ok := conn.(*listener.DummyUDPConn); !ok {
		_, err := io.Copy(conn, conn)
		return err
	}

	defer conn.Close()

	buff := [65535]byte{}

	n, err := conn.Read(buff[:])
	if err != nil {
		return err
	}
	s.c.Send(event.New(
		EventOptions,
		event.Category("echo"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	_, err = conn.Write(buff[:n])
	return err
}
