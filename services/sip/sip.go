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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/marv2097/siprocket"
)

// please use "go get -u github.com/marv2097/siprocket" to import the siprocket library

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
	ErrBadMessage = errors.New("sip: bad message")
	letters       = []rune("abcdefghijklmnopqrstuvwxyz_0123456789")
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
	sipRequest = map[string]func(*sipService, *request) map[string][]string{
		"INVITE":  (*sipService).InviteMethod,
		"OPTIONS": (*sipService).OptionsMethod,
		"PUBLISH": (*sipService).PublishMethod,
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
}

type request struct {
	Method, UriType, User, Host, SIPVersion, Uri, Src, RemoteIP, LocalIP string
}

func (s *sipService) SetChannel(ch pushers.Channel) {
	s.ch = ch
}

func randomID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *sipService) Handle(ctx context.Context, conn net.Conn) error {
	var ok bool
	rand.Seed(time.Now().UnixNano())
	r := &request{}
	var data bytes.Buffer
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil {
		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("sip"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Payload([]byte(line)),
		))
		return err
	}

	args := strings.Split(line, " ")
	if len(args) != 3 {
		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("sip"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Payload([]byte(line)),
		))
		return ErrBadMessage
	}

	sip := siprocket.Parse([]byte(line))

	sip_request := strings.Split(line, " ")

	r.Host = fmt.Sprintf("%s", sip.Req.Host)
	r.Method = fmt.Sprintf("%s", sip.Req.Method)
	r.UriType = fmt.Sprintf("%s", sip.Req.UriType)
	r.User = fmt.Sprintf("%s", sip.Req.User)
	r.Src = fmt.Sprintf("%s", sip.Req.Src)
	r.Uri = strings.TrimSpace(sip_request[1])
	r.SIPVersion = strings.TrimSpace(sip_request[2])
	r.LocalIP = fmt.Sprintf("%s", conn.LocalAddr().(*net.TCPAddr).IP)
	r.RemoteIP = fmt.Sprintf("%s", conn.RemoteAddr().(*net.TCPAddr).IP)

	s.ch.Send(event.New(
		services.EventOptions,
		event.Category("sip"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("sip.method", r.Method),
		event.Custom("sip.uri", r.Uri),
		event.Custom("sip.source", r.Src),
		event.Custom("sip.version", r.SIPVersion),
		event.Custom("sip.uritype", r.UriType),
		event.Custom("sip.user", r.User),
		event.Custom("sip.host", r.Host),
	))

	for i, _ := range Map_Method {
		if r.Method == Map_Method[i] {
			ok = true
			break
		}
	}

	if !ok || r.SIPVersion != "SIP/2.0" {
		s.ch.Send(event.New(
			event.Payload([]byte(line)),
		))
		return ErrBadMessage
	}

	fn := sipRequest[r.Method]

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     fn(s, r),
	}

	encode, err := json.Marshal(fn(s, r))
	if err != nil {
		s.ch.Send(event.New(
			event.Payload(encode),
		))
		return err
	}

	data.Write(encode)

	if r.Method == "INVITE" {
		resp.Body = ioutil.NopCloser(strings.NewReader(s.InviteBody(r)))
		data.Write([]byte(s.InviteBody(r)))
	}

	s.ch.Send(event.New(
		event.Payload(data.Bytes()),
	))

	return resp.Write(conn)
}
