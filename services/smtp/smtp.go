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
package smtp

import (
	"context"
	"net"

	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var (
	_   = services.Register("smtp", SMTP)
	log = logging.MustGetLogger("services/smtp")
)

// SMTP
func SMTP(options ...services.ServicerFunc) services.Servicer {

	s := &Service{
		Config: Config{
			Banner: "SMTPd",
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

	s.srv.Banner = s.Banner

	handler := HandleFunc(func(msg Message) error {
		s.receiveChan <- msg
		return nil
	})

	s.srv.Handler = handler

	return s
}

type Config struct {
	Banner string `toml:"banner"`

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

	rcvLine := make(chan string)

	// Wait for a message and send it into the eventbus
	go func() {
		for {
			select {
			case <-time.After(time.Minute * 2):
				log.Error("timeout expired")
				return
			case message := <-s.receiveChan:
				s.ch.Send(event.New(
					services.EventOptions,
					event.Category("smtp"),
					event.Type("email"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("smtp.from", message.From),
					event.Custom("smtp.to", message.To),
					event.Custom("smtp.body", message.Body.String()),
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
			}
		}
	}()

	//Create new smtp server connection
	c := s.srv.newConn(conn, rcvLine)
	// Start server loop
	c.serve()
	return nil
}
