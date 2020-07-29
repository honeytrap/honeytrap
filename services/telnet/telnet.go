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
package telnet

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
	"github.com/rs/xid"
)

var log = logging.MustGetLogger("services:telnet")

var (
	_ = services.Register("telnet", Telnet)
)

var (
	motd = `********************************************************************************
*             Copyright(C) 2008-2015 Huawei Technologies Co., Ltd.             *
*                             All rights reserved                              *
*                  Without the owner's prior written consent,                  *
*           no decompiling or reverse-engineering shall be allowed.            *
* Notice:                                                                      *
*                   This is a private communication system.                    *
*             Unauthorized access or use may lead to prosecution.              *
********************************************************************************

Warning: Telnet is not a secure protocol, and it is recommended to use STelnet. 

Login authentication


`
	prompt = `$ `
)

// Telnet is a placeholder
func Telnet(options ...services.ServicerFunc) services.Servicer {
	s := &telnetService{
		MOTD:   motd,
		Prompt: prompt,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type telnetService struct {
	c pushers.Channel

	Prompt string `toml:"prompt"`
	MOTD   string `toml:"motd"`
}

func (s *telnetService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *telnetService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	defer conn.Close()
	log.Debug("Telnet handling started")

	var connOptions event.Option = nil

	if ec, ok := conn.(*event.Conn); ok {
		connOptions = ec.Options()
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("telnet"),
		event.Type("connect"),
		connOptions,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("telnet.sessionid", id.String()),
	))

	term := NewTerminal(conn, s.Prompt)

	term.Write([]byte(s.MOTD + "\n"))

	term.SetPrompt("Username: ")
	username, err := term.ReadLine()
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	password, err := term.ReadPassword("Password: ")
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("telnet"),
		event.Type("password-authentication"),
		connOptions,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("telnet.sessionid", id.String()),
		event.Custom("telnet.username", username),
		event.Custom("telnet.password", password),
	))

	term.SetPrompt(s.Prompt)

	for {
		line, err := term.ReadLine()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		s.c.Send(event.New(
			services.EventOptions,
			event.Category("telnet"),
			event.Type("session"),
			connOptions,
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("telnet.sessionid", id.String()),
			event.Custom("telnet.command", line),
		))

		if line == "" {
			continue
		}

		term.Write([]byte(fmt.Sprintf("sh: %s: command not found\n", line)))
	}
}
