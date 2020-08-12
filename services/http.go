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
	defer conn.Close()

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
