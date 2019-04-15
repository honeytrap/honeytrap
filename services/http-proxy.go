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
	"io"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("http-proxy", HTTPProxy)
)

// HTTP
func HTTPProxy(options ...ServicerFunc) Servicer {
	s := &httpProxy{}
	for _, o := range options {
		o(s)
	}

	// todo
	// if no director set
	// error
	return s
}

type httpProxy struct {
	c pushers.Channel
	d director.Director
}

func (s *httpProxy) SetDirector(d director.Director) {
	s.d = d
}

func (s *httpProxy) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *httpProxy) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	conn2, err := s.d.Dial(conn)
	if err != nil {
		return err
	}

	defer conn2.Close()

	for {
		reader := bufio.NewReader(conn)
		req, err := http.ReadRequest(reader)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		reqBody := &bytes.Buffer{}

		dsw := io.MultiWriter(conn2, reqBody)

		if err = req.Write(dsw); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		s.c.Send(event.New(
			SensorLow,
			event.Service("http-proxy"),
			event.Category("http"),
			event.Custom("method", req.Method),
			event.Custom("host", req.Host),
			event.Custom("user-agent", req.UserAgent()),
			event.Custom("referer", req.Referer()),
			event.Custom("url", req.URL.String()),
			event.Custom("content-length", req.ContentLength),
			event.RemoteAddr(conn.RemoteAddr().String()),
			event.Payload(reqBody.Bytes()),
		))

		var resp *http.Response

		reader2 := bufio.NewReader(conn2)
		resp, err = http.ReadResponse(reader2, req)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		err = resp.Write(conn)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}
