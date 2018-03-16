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

func (s *httpProxy) SetDataDir(string) {}

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
