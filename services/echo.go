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
