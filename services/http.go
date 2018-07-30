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
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/rs/xid"
)

var (
	_ = Register("http", HTTP)
)

// Http is a placeholder
func HTTP(options ...ServicerFunc) Servicer {
	s := &httpService{
		httpServiceConfig: httpServiceConfig{
			Server: "Apache",
		},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type httpServiceConfig struct {
	Server string `toml:"server"`
}

type httpService struct {
	httpServiceConfig

	c pushers.Channel
}

func (s *httpService) CanHandle(payload []byte) bool {
	if bytes.HasPrefix(payload, []byte("GET")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("HEAD")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("POST")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("PUT")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("DELETE")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("PATCH")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("TRACE")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("CONNECT")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("OPTIONS")) {
		return true
	}

	return false
}

func (s *httpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

func Cookies(cookies []*http.Cookie) event.Option {
	return func(m event.Event) {
		for _, c := range cookies {
			m.Store(fmt.Sprintf("http.cookie.%s", strings.ToLower(c.Name)), c.Value)
		}
	}
}

func (s *httpService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	for {
		br := bufio.NewReader(conn)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		defer req.Body.Close()

		body := make([]byte, 1024)

		n, err := req.Body.Read(body)
		if err == io.EOF {
		} else if err != nil {
			return err
		}

		body = body[:n]

		io.Copy(ioutil.Discard, req.Body)

		var connOptions event.Option = nil

		if ec, ok := conn.(*event.Conn); ok {
			connOptions = ec.Options()
		}

		s.c.Send(event.New(
			EventOptions,
			connOptions,
			event.Category("http"),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("http.sessionid", id.String()),
			event.Custom("http.method", req.Method),
			event.Custom("http.proto", req.Proto),
			event.Custom("http.host", req.Host),
			event.Custom("http.url", req.URL.String()),
			event.Payload(body),
			Headers(req.Header),
			Cookies(req.Cookies()),
		))

		resp := http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			Header: http.Header{
				"Server": []string{s.Server},
			},
		}

		if err := resp.Write(conn); err != nil {
			return err
		}
	}
}
