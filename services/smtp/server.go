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
package smtp

import (
	"crypto/tls"
	"net"
	"sync"

	"github.com/honeytrap/honeytrap/event"
)

type Handler interface {
	Serve(msg Message) error
}

type HandlerFunc func(msg Message) error

type ServeMux struct {
	m  []HandlerFunc
	mu sync.RWMutex
}

func (mux *ServeMux) HandleFunc(handler func(msg Message) error) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	mux.m = append(mux.m, handler)
}

func HandleFunc(handler func(msg Message) error) *ServeMux {
	DefaultServeMux.HandleFunc(handler)
	return DefaultServeMux
}

var DefaultServeMux = NewServeMux()

func NewServeMux() *ServeMux { return &ServeMux{m: make([]HandlerFunc, 0)} }

func (mux *ServeMux) Serve(msg Message) error {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	for _, h := range mux.m {
		if err := h(msg); err != nil {
			return err
		}
	}
	return nil
}

type Server struct {
	Banner string

	Handler Handler

	tlsConfig *tls.Config
}

func (s *Server) newConn(rwc net.Conn, recv chan string, evnt chan event.Event) *conn {
	c := &conn{
		server: s,
		rwc:    rwc,
		rcv:    recv,
		i:      0,
		evnt:   evnt,
	}

	c.msg = c.newMessage()
	return c
}

func (s *Server) tlsConf(cert *tls.Certificate) {
	s.tlsConfig = &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
	}
}

type serverHandler struct {
	srv *Server
}

func (sh serverHandler) Serve(msg Message) {
	handler := sh.srv.Handler
	if handler == nil {
		handler = DefaultServeMux
	}

	handler.Serve(msg)
}
