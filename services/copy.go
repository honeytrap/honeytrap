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
