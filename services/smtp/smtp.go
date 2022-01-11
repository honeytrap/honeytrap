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
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/bannerfmt"
	logging "github.com/op/go-logging"
)

const readDeadline = 5 // connection deadline in minutes
/*
[service.smtp]
type="smtp"

[[port]]
port="tcp/25"
services=["smtp"]
*/

var (
	_   = services.Register("smtp", SMTP)
	log = logging.MustGetLogger("services/smtp")
)

// SMTP
func SMTP(options ...services.ServicerFunc) services.Servicer {

	s := &Service{
		Config: Config{
			bannerData: bannerData{
				BannerTemplate: "{{.Host}} {{.Name}} Ready",
				Host:           "remailer.ru",
				Name:           "SMTP",
			},
			srv: &Server{
				tlsConfig: nil,
			},
			receiveChan: make(chan Message),
		},
	}

	// Load config options
	for _, o := range options {
		o(s)
	}

	if store, err := getStorage(); err != nil {
		log.Errorf("Could not initialize storage: %s", err.Error())
	} else {

		cert, err := store.Certificate()
		if err != nil {
			log.Debugf("SMTP: No TLS, %s", err.Error())
		} else {
			s.srv.tlsConf(cert)

			log.Debug("SMTP server: set tls certificate")
		}
	}

	banner, err := bannerfmt.New(s.BannerTemplate, s.Config.bannerData)
	if err != nil {
		log.Error(err.Error())
	}

	s.srv.Banner = banner.String()

	handler := HandleFunc(func(msg Message) error {
		s.receiveChan <- msg
		return nil
	})

	s.srv.Handler = handler

	return s
}

type bannerData struct {
	BannerTemplate string `toml:"banner-fmt"`

	Host string `toml:"host"`

	Name string `toml:"name"`
}

type Config struct {
	bannerData

	srv *Server

	receiveChan chan Message
}

type Service struct {
	Config

	ch pushers.Channel
}

func (s *Service) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *Service) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(time.Minute * readDeadline)); err != nil {
		return errors.New("Can't set ReadDeadline on connection")
	}

	rcvLine := make(chan string)

	evntChan := make(chan event.Event)

	// Wait for a message and send it into the eventbus
	go func() {
		for {
			select {
			case message := <-s.receiveChan:
				header := []event.Option{}

				for key, values := range message.Header {
					var vals strings.Builder

					for _, s := range values {
						if vals.Len() > 0 {
							_, _ = vals.WriteRune(',')
						}
						_, _ = vals.WriteString(s)
					}

					header = append(header, event.Custom("smtp."+key, vals.String()))
				}

				s.ch.Send(event.New(
					services.EventOptions,
					event.Category("smtp"),
					event.Type("email"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("smtp.body", message.Body.String()),
					event.NewWith(header...),
				))
			case line := <-rcvLine:
				s.ch.Send(event.New(
					services.EventOptions,
					event.Category("smtp"),
					event.Type("input"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("smtp.line", line),
				))
			case evnt := <-evntChan:
				s.ch.Send(evnt)
			}
		}
	}()

	//Create new smtp server connection
	c := s.srv.newConn(conn, rcvLine, evntChan)
	// Start server loop
	c.serve()
	return nil
}
