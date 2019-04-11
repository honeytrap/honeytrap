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
	"bufio"
	"bytes"
	"context"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

// Documentation: https://developer.valvesoftware.com/wiki/Server_Queries

var (
	_ = Register("counterstrike", CounterStrike)
)

const (
	COUNTERSTRIKE_A2S_INFO                     byte = 0x54
	COUNTERSTRIKE_A2S_PLAYER                        = 0x55
	COUNTERSTRIKE_A2S_RULES                         = 0x56
	COUNTERSTRIKE_A2S_SERVERQUERY_GETCHALLENGE      = 0x57
	COUNTERSTRIKE_A2A_PING                          = 0x69
)

var (
	messages = map[byte][]byte{
		COUNTERSTRIKE_A2S_INFO: []byte{
			0xff, 0xff, 0xff, 0xff, 0x49, 0x11, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65,
			0x72, 0x2d, 0x53, 0x74, 0x72, 0x69, 0x6b, 0x65, 0x3a, 0x20, 0x47, 0x6c,
			0x6f, 0x62, 0x61, 0x6c, 0x20, 0x4f, 0x66, 0x66, 0x65, 0x6e, 0x73, 0x69,
			0x76, 0x65, 0x00, 0x64, 0x65, 0x5f, 0x64, 0x75, 0x73, 0x74, 0x32, 0x00,
			0x63, 0x73, 0x67, 0x6f, 0x00, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72,
			0x2d, 0x53, 0x74, 0x72, 0x69, 0x6b, 0x65, 0x3a, 0x20, 0x47, 0x6c, 0x6f,
			0x62, 0x61, 0x6c, 0x20, 0x4f, 0x66, 0x66, 0x65, 0x6e, 0x73, 0x69, 0x76,
			0x65, 0x00, 0xda, 0x02, 0x00, 0x0a, 0x00, 0x64, 0x6c, 0x00, 0x01, 0x31,
			0x2e, 0x33, 0x36, 0x2e, 0x33, 0x2e, 0x34, 0x00, 0xa1, 0x88, 0x69, 0x76,
			0x61, 0x6c, 0x76, 0x65, 0x5f, 0x64, 0x73, 0x2c, 0x65, 0x6d, 0x70, 0x74,
			0x79, 0x2c, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65, 0x00, 0xda, 0x02, 0x00,
		},
	}
)

func CounterStrike(options ...ServicerFunc) Servicer {
	s := &counterStrikeService{
		limiter: NewLimiter(),
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type counterStrikeService struct {
	limiter *Limiter

	ch pushers.Channel
}

func (s *counterStrikeService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *counterStrikeService) Handle(ctx context.Context, conn net.Conn) error {
	b := bufio.NewReader(conn)

	// counterStrike behaves differently over UDP: it has an 8-bytes header
	if conn.RemoteAddr().Network() != "udp" {
		return nil
	}

	buf := make([]byte, 1024)

	n, err := b.Read(buf)
	if err != nil {
		log.Error("Error processing UDP header: %s", err.Error())
	}

	buf = buf[:n]

	if bytes.Compare(buf[0:4], []byte{0xff, 0xff, 0xff, 0xff}) == 0 {
		// Simple response
	} else if bytes.Compare(buf[0:4], []byte{0xff, 0xff, 0xff, 0xfe}) == 0 {
		// Multi-packet response
	} else {
		return nil
	}

	query := buf[4]

	if query == COUNTERSTRIKE_A2S_INFO {
		payload := string(buf[5:])

		s.ch.Send(event.New(
			EventOptions,
			event.Category("counterstrike"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("counterstrike.query", "a2s_info"),
			event.Custom("counterstrike.payload", payload),
			event.Payload(buf),
		))

	} else if query == COUNTERSTRIKE_A2S_PLAYER {
		// Challenge number
		s.ch.Send(event.New(
			EventOptions,
			event.Category("counterstrike"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("counterstrike.query", "a2s_player"),
			event.Payload(buf),
		))

	} else if query == COUNTERSTRIKE_A2S_RULES {
		s.ch.Send(event.New(
			EventOptions,
			event.Category("counterstrike"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("counterstrike.query", "a2s_rules"),
			event.Payload(buf),
		))

	} else if query == COUNTERSTRIKE_A2S_SERVERQUERY_GETCHALLENGE {
		s.ch.Send(event.New(
			EventOptions,
			event.Category("counterstrike"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("counterstrike.query", "a2s_serverquery_challenge"),
			event.Payload(buf),
		))

	} else if query == COUNTERSTRIKE_A2A_PING {
		s.ch.Send(event.New(
			EventOptions,
			event.Category("counterstrike"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("counterstrike.query", "a2s_ping"),
			event.Payload(buf),
		))

	}

	// we return errors for udp connections, to prevent udp amplification
	if conn.RemoteAddr().Network() != "udp" {
	} else if !s.limiter.Allow(conn.RemoteAddr()) {
		return nil
	}

	if msg, ok := messages[COUNTERSTRIKE_A2S_INFO]; ok {
		conn.Write(msg)
	}

	return nil
}
