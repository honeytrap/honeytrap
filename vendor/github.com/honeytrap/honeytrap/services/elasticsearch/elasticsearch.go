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
