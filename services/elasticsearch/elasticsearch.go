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
package elasticsearch

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"io/ioutil"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

var (
	_ = services.Register("elasticsearch", Elasticsearch)
)

// Elasticsearch is a placeholder
func Elasticsearch(options ...services.ServicerFunc) services.Servicer {
	s := &service{
		serviceConfig: serviceConfig{},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type serviceConfig struct {
	Name        string `toml:"name"`
	ClusterName string `toml:"cluster_name"`
	ClusterUUID string `toml:"cluster_uuid"`
}

type service struct {
	serviceConfig

	c pushers.Channel
}

func (s *service) SetChannel(c pushers.Channel) {
	s.c = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

func (s *service) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

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

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("elasticsearch"),
		event.Type("request"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Payload(body),
		Headers(req.Header),
	))

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header:     http.Header{},
	}

	if req.URL.Path == "/" {
		buff := bytes.Buffer{}

		if err := json.NewEncoder(&buff).Encode(map[string]interface{}{
			"name":         s.Name,
			"cluster_name": s.ClusterName,
			"cluster_uuid": s.ClusterUUID,
			"version": map[string]interface{}{
				"number":         "5.4.1",
				"build_hash":     "2cfe0df",
				"build_date":     "2017-05-29T16:05:51.443Z",
				"build_snapshot": false,
				"lucene_version": "6.5.1",
			},
			"tagline": "You Know, for Search",
		}); err != nil {
			return err
		}

		resp.Header.Add("content-type", "application/json; charset=UTF-8")
		resp.Header.Add("content-length", fmt.Sprintf("%d", buff.Len()))

		resp.Body = ioutil.NopCloser(&buff)
	}

	if err := resp.Write(conn); err != nil {
		return err
	}

	return nil
}
