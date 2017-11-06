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
	"crypto/tls"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("smtp", SMTP)
)

// SMTP
func SMTP(options ...ServicerFunc) Servicer {

	server, err := New()
	if err != nil {
		return nil
	}

	s := &SMTPService{
		srv: server,
	}
	//TODO: make certificate configurable
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err == nil {
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	for _, o := range options {
		o(s)
	}
	return s
}

type SMTPService struct {
	ch  pushers.Channel
	srv *Server
}

func (s *SMTPService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *SMTPService) Handle(conn net.Conn) error {

	receiveChan := make(chan Message)

	go func() {
		for {
			select {
			case message := <-receiveChan:
				log.Debug("Message Received")
				s.ch.Send(event.New(
					EventOptions,
					event.Category("smtp"),
					event.Type("mail"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("smtp.From", message.From),
					event.Custom("smtp.To", message.To),
					event.Custom("smtp.Body", message.Body.String()),
				))
			}
		}
	}()

	handler := HandleFunc(func(msg Message) error {
		receiveChan <- msg
		return nil
	})
	s.srv.Handler = handler

	c, err := s.srv.NewConn(conn)
	if err != nil {
		return err
	}

	// Use a go routine here???
	c.serve()

	return nil
}
