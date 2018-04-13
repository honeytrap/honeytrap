/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ftp

import (
	"context"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/filesystem"
	logging "github.com/op/go-logging"
)

var (
	_   = services.Register("ftp", FTP)
	log = logging.MustGetLogger("services/ftp")
)

// FTP setup the FTP service
func FTP(options ...services.ServicerFunc) services.Servicer {

	store, err := Storage()
	if err != nil {
		log.Errorf("FTP: Could not initialize storage. %s", err.Error())
	}

	cert, err := store.Certificate()
	if err != nil {
		log.Errorf("TLS error: %s", err.Error())
	}

	s := &ftpService{
		Opts: Opts{},
		recv: make(chan string),
	}

	for _, o := range options {
		o(s)
	}

	ftpusers := make(User)

	for _, u := range s.Users {
		ftpusers[u[0]] = u[1]
	}

	opts := &ServerOpts{
		Auth:           ftpusers,
		Name:           s.ServerName,
		WelcomeMessage: s.Banner,
		PassivePorts:   s.PsvPortRange,
	}

	s.server = NewServer(opts)

	s.server.tlsConfig = simpleTLSConfig(cert)
	if s.server.tlsConfig != nil {
		//s.server.TLS = true
		s.server.ExplicitFTPS = true
	}

	base, root := store.FileSystem()
	if base == "" {
		base = s.FsRoot
	}

	fs, err := filesystem.New(base, "ftp", root)
	if err != nil {
		log.Debugf("FTP Filesystem error: %s", err.Error())
	}

	log.Debugf("FileSystem rooted at %s", fs.RealPath("/"))

	s.driver = NewFileDriver(fs)

	return s
}

// Opts are the options neccesary for runnung the server
// They can be set in config file
type Opts struct {
	Banner string `toml:"banner"`

	PsvPortRange string `toml:"passive-port-range"`

	ServerName string `toml:"name"`

	Users [][]string `toml:"users"`
}

type ftpService struct {
	Opts

	server *Server

	driver Driver

	FsRoot string `toml:"fs_base"`

	recv chan string

	c pushers.Channel
}

func (s *ftpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *ftpService) Handle(ctx context.Context, conn net.Conn) error {

	ftpConn := s.server.newConn(conn, s.driver, s.recv)

	go func() {
		for {
			select {
			case msg := <-s.recv:
				s.c.Send(event.New(
					services.EventOptions,
					event.Category("ftp"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("ftp.sessionid", ftpConn.sessionid),
					event.Custom("ftp.command", strings.Trim(msg, "\r\n")),
				))
			}
		}
	}()

	ftpConn.Serve()

	return nil
}
