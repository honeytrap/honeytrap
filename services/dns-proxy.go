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
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/miekg/dns"
)

var (
	_ = Register("dns-proxy", DNSProxy)
)

// Dns is a placeholder
func DNSProxy(options ...ServicerFunc) Servicer {
	s := &dnsProxy{}
	for _, o := range options {
		o(s)
	}
	return s
}

type dnsProxy struct {
	c pushers.Channel

	d director.Director
}

func (s *dnsProxy) SetDirector(d director.Director) {
	s.d = d
}

func (s *dnsProxy) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *dnsProxy) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	buff := [65535]byte{}

	if _, ok := conn.(*listener.DummyUDPConn); ok {
		n, err := conn.Read(buff[:])
		if err != nil {
			return err
		}

		conn2, err := s.d.Dial(conn)
		if err != nil {
			return err
		}

		defer conn2.Close()

		if _, err = conn2.Write(buff[:n]); err != nil {
			return err
		}

		req := new(dns.Msg)
		if err := req.Unpack(buff[:n]); err != nil {
			return err
		}

		s.c.Send(event.New(
			EventOptions,
			event.Category("dns-proxy"),
			event.Type("dns"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("dns.id", fmt.Sprintf("%d", req.Id)),
			event.Custom("dns.opcode", fmt.Sprintf("%d", req.Opcode)),
			event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", req.Question)),
			event.Custom("dns.questions", req.Question),
		))

		if n, err = conn2.Read(buff[:]); err != nil {
			return err
		}

		if _, err = conn.Write(buff[:n]); err != nil {
			return err
		}

		return err
	} else if _, ok := conn.(*net.TCPConn); ok {
		n, err := conn.Read(buff[:])
		if err != nil {
			return err
		}

		req := new(dns.Msg)
		if err := req.Unpack(buff[:n]); err != nil {
			return err
		}

		s.c.Send(event.New(
			EventOptions,
			event.Category("dns-proxy"),
			event.Type("dns"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("dns.id", fmt.Sprintf("%d", req.Id)),
			event.Custom("dns.opcode", fmt.Sprintf("%d", req.Opcode)),
			event.Custom("dns.message", fmt.Sprintf("Querying for: %#q", req.Question)),
			event.Custom("dns.questions", req.Question),
		))

		conn2, err := s.d.Dial(conn)
		if err != nil {
			return err
		}

		defer conn2.Close()

		if _, err = conn2.Write(buff[:n]); err != nil {
			return err
		}

		if n, err = conn2.Read(buff[:]); err != nil {
			return err
		}

		if _, err = conn.Write(buff[:n]); err != nil {
			return err
		}

		return nil
	} else {
		return nil
	}
}
