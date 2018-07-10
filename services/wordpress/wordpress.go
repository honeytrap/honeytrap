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
package wordpress

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

var (
	_   = services.Register("wordpress", WordPress)
	log = logging.MustGetLogger("services/wordpress")
)

func WordPress(options ...services.ServicerFunc) services.Servicer {
	s := &wpService{
		wpServiceConfig: wpServiceConfig{
			Server: "Apache",
		},
	}

	for _, o := range options {
		o(s)
	}

	target := path.Join(s.assets, "index.html")
	_, err := ioutil.ReadFile(target)
	if err != nil {
		log.Errorf("Failed to read WordPress assets (couldn't find %s). Please download them from https://github.com/honeytrap/honeytrap-services-wordpress into %s.", target, s.assets)
	}

	return s
}

type wpServiceConfig struct {
	Server string `toml:"server"`
}

type wpService struct {
	wpServiceConfig

	assets string
	c      pushers.Channel
}

func (s *wpService) CanHandle(payload []byte) bool {
	if bytes.HasPrefix(payload, []byte("GET")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("HEAD")) {
		return true
	} else if bytes.HasPrefix(payload, []byte("POST")) {
		return true
	}

	return false
}

func (s *wpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *wpService) SetDataDir(dir string) {
	s.assets = path.Join(dir, "wordpress")
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

func (s *wpService) makeResponse(req *http.Request) http.Response {
	return http.Response{
		StatusCode: http.StatusOK, // default value, can be overridden of course
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header: http.Header{
			"Server": []string{s.Server},
		},
	}
}

func (s *wpService) Handle(ctx context.Context, conn net.Conn) error {
	for {
		br := bufio.NewReader(conn)

		req, err := http.ReadRequest(br)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		s.c.Send(event.New(
			services.EventOptions,
			event.Category("wordpress"),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("wordpress.method", req.Method),
			event.Custom("wordpress.proto", req.Proto),
			event.Custom("wordpress.host", req.Host),
			event.Custom("wordpress.url", req.URL.String()),
			Headers(req.Header),
			Cookies(req.Cookies()),
		))

		if req.Method == "POST" && strings.Contains(req.URL.Path, "wp-login") {
			// Log the attempt. It will look like a failed login, because the
			// login page is returned again by the rest of the code
			req.ParseForm()
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("wordpress"),
				event.Type("login-attempt"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("login.path", req.URL.Path),
				event.Custom("login.username", req.Form["log"][0]),
				event.Custom("login.password", req.Form["pwd"][0]),
				Headers(req.Header),
				Cookies(req.Cookies()),
			))
		}

		_reqPath := req.URL.Path
		reqPath := _reqPath[1:]
		filePath := path.Clean(reqPath)
		// Reject directory trasversal attacks
		if strings.Contains(filePath, "..") {
			resp := s.makeResponse(req)
			resp.StatusCode = http.StatusBadRequest
			resp.Status = http.StatusText(http.StatusBadRequest)
			return resp.Write(conn)
		}
		if filePath == "." {
			filePath = "index.html"
		}
		filePath = path.Join(s.assets, filePath)
		data, err := ioutil.ReadFile(filePath)

		if err != nil {
			resp := s.makeResponse(req)
			resp.StatusCode = http.StatusNotFound
			resp.Status = http.StatusText(http.StatusNotFound)
			resp.Body = ioutil.NopCloser(bytes.NewBufferString("Not found"))
			return resp.Write(conn)
		}
		resp := s.makeResponse(req)
		resp.Body = ioutil.NopCloser(bytes.NewReader(data))
		return resp.Write(conn)
	}
}
