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
package telnet

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/rs/xid"
)

var (
	_ = services.Register("telnet", Telnet)
)

var (
	motd = `Welcome to Microsoft Telnet Service 


login:`
	prompt = `C:\DOCUME~1\Doug> `
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

	term := NewTerminal(conn, s.Prompt)

	term.Write([]byte(s.MOTD))

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
			event.Conn(conn),
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
