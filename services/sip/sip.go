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

package sip

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

/*Example config:

[service.sip]
type="sip"
os="Linux"

[[port]]
port="tcp/5060"
services=["sip"]

*/

var (
	_             = services.Register("sip", SIP)
	ErrBadMessage = errors.New("bad message")
	Map_Method    = map[string]string{
		"MethodInvite":    "INVITE",
		"MethodAck":       "ACK",
		"MethodBye":       "BYE",
		"MethodCancel":    "CANCEL",
		"MethodRegister":  "REGISTER",
		"MethodOptions":   "OPTIONS",
		"MethodPrack":     "PRACK",
		"MethodSubscribe": "SUBSCRIBE",
		"MethodNotify":    "NOTIFY",
		"MethodPublish":   "PUBLISH",
		"MethodInfo":      "INFO",
		"MethodRefer":     "REFER",
		"MethodMessage":   "MESSAGE",
		"MethodUpdate":    "UPDATE",
	}
	sipRequest = map[string]func(*sipService) string{
		"OPTIONS": (*sipService).OptionMethod,
	}
)

func SIP(options ...services.ServicerFunc) services.Servicer {
	s := &sipService{
		sipServiceConfig: sipServiceConfig{
			Os: "Linux",
		},
	}
	for _, o := range options {
		o(s)
	}
	return s
}

type sipServiceConfig struct {
	Os string `toml:"os"`
}

type sipService struct {
	sipServiceConfig

	ch pushers.Channel

	Method, Uri, SIPVersion, Username, Domain string

	Body []byte
}

func (s *sipService) SetChannel(ch pushers.Channel) {
	s.ch = ch
}

func (s *sipService) Handle(ctx context.Context, conn net.Conn) error {
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil {
		return err
	}

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("sip"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(s.Body),
	))

	args := strings.Split(line, " ")
	if len(args) != 3 {
		return ErrBadMessage
	}

	s.Method = args[0]
	s.Uri = args[1]
	s.SIPVersion = strings.TrimSpace(args[2])

	ok := s.checkRequest(line)
	if !ok {
		return ErrBadMessage
	}

	fn := sipRequest[s.Method]

	s.Body = []byte(fn(s))

	conn.Write(s.Body)

	return nil
}
