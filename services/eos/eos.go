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
package eos

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
	logging "github.com/op/go-logging"
)

var (
	_ = services.Register("eos", EOS)
)

var log = logging.MustGetLogger("honeytrap/services/eos")

// EOS is a placeholder
func EOS(options ...services.ServicerFunc) services.Servicer {
	s := &eosService{
		eosServiceConfig: eosServiceConfig{},
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type eosServiceConfig struct {
	Name string `toml:"name"`
}

type eosService struct {
	eosServiceConfig

	c pushers.Channel
}

func (s *eosService) SetChannel(c pushers.Channel) {
	s.c = c
}

func Headers(headers map[string][]string) event.Option {
	return func(m event.Event) {
		for name, h := range headers {
			m.Store(fmt.Sprintf("http.header.%s", strings.ToLower(name)), h)
		}
	}
}

var eosMethods = map[string]func() interface{}{
	"/v1/wallet/list_keys": func() interface{} {
		return []interface{}{
			[]string{
				"EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV",
				"5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3",
			},
		}
	},
}

// Todo: implement CanHandle

func (s *eosService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	br := bufio.NewReader(conn)

	req, err := http.ReadRequest(br)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(req.Body)
	if err == io.EOF {
	} else if err != nil {
		return err
	}

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("eos"),
		event.Type("rpc-api"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.user-agent", req.UserAgent()),
		event.Custom("http.method", req.Method),
		event.Custom("http.proto", req.Proto),
		event.Custom("http.host", req.Host),
		event.Custom("http.url", req.URL.String()),
		event.Custom("eos.method", req.URL.Path),
		event.Payload(data),
		Headers(req.Header),
	))

	buff := bytes.Buffer{}

	fn, ok := eosMethods[req.URL.Path]
	if ok {
		v := fn()

		if err := json.NewEncoder(&buff).Encode(v); err != nil {
			return err
		}
	} else {
		log.Errorf("Method %s not supported", req.URL.Path)
	}

	resp := http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Request:    req,
		Header:     http.Header{},
	}

	resp.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp.Header.Add("Content-Length", fmt.Sprintf("%d", buff.Len()))

	resp.Body = ioutil.NopCloser(&buff)

	return resp.Write(conn)
}
