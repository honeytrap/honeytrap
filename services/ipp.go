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
	"io"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("ipp", IPP)
)

func IPP(options ...ServicerFunc) Servicer {
	s := &ippService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type ippService struct {
	httpServiceConfig

	ch pushers.Channel
}

func (s *ippService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *ippService) Handle(conn net.Conn) error {

	for {
		br := bufio.NewReader(conn)

		req, err := http.ReadRequest(br)
		if err == io.EOF || req.Method != "POST" || req.Header.Get("Content-Type") != "application/ipp" {
			return nil
		} else if err != nil {
			return err
		}

		breq := bytes.Buffer
		n, err := req.Body.Read(breq)

		bresp := ippHandler(&b)

		s.c.Send(event.New(
			EventOptions,
			event.Category("http"),
			event.Type("request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("http.url", req.URL.String()),
		))

		resp := http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			Header: http.Header{
				"Server":       []string{s.Server},
				"Content-Type": "application/ipp",
			},
			Body: bresp,
		}
		if err := resp.Write(conn); err != nil {
			return err
		}
	}
	// Wait for a IPPMessage and push it in the eventbus
	go func() {
		for {
			select {
			case message := <-ippChan:
				log.Debug("IPP: Message received")
				s.ch.Send(event.New(
					EventOptions,
					event.Category("ipp"),
					event.Type("print"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("ipp.user", message.User),
				))
			}
		}
	}()

}
