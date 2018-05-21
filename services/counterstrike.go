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
