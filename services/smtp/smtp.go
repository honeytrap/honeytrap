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

/*
  config options:

	[service.smtp]
	type="smtp"
	host="smtp.mailer.org"
	name="HT SMTP-E v2.3.0.1b"
	banner-fmt='{{.Host}} {{.Name}} - Ready'

	# Standard smtp server port
	[[port]]
	port="tcp/25"
	services=["smtp"]

	# Standard smtp client port
	[[port]]
	port="tcp/587"
	services=["smtp"]

	[service.smtps]
	type="smtp"
	name="SMTPS"
	implicit_tls = true
	banner-fmt='{{.Host}} {{.Name}} ready'

	# Standard smtps port (Can only connect with tls)
	[[port]]
	port="tcp/465"
	services=["smtps"]
*/

import (
	"bufio"
	"context"
	"crypto/tls"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/bannerfmt"
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

type bufferedConn struct {
	r *bufio.Reader

	net.Conn
}

func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b bufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

func smtpConn(c net.Conn, srv *Server) net.Conn {
	sconn := bufferedConn{
		bufio.NewReader(c),
		c,
	}

	// Peek first 3 bytes for tls handshake
	buf, err := sconn.Peek(3)
	if err != nil {
		log.Debug(err.Error())
		return sconn
	}

	log.Debugf("Peeked bytes %#v", buf)

	// validate header byte 0 [record type], 1 [version major], 2 [version minor]
	if len(buf) == 3 && buf[0] == 0x16 && buf[1] == 0x03 && buf[2] <= 0x03 {

		log.Debug("TLS detected")
		// Most likely tls
		tlsconn := tls.Server(sconn, srv.tlsConfig)
		if err := tlsconn.Handshake(); err != nil {
			log.Debugf("TLS detected, but handshake errors: %s", err.Error())
			return sconn
		}

		return tlsconn
	}

	return sconn
}

func (s *Service) Handle(ctx context.Context, conn net.Conn) error {

	// Check for tls
	sc := smtpConn(conn, s.srv)
	defer sc.Close()

	rcvLine := make(chan string)
	done := make(chan struct{})

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
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	//Create new smtp server connection
	c := s.srv.newConn(sc, rcvLine)
	// Start server loop
	c.serve()

	// close go routine
	done <- struct{}{}

	return nil
}
