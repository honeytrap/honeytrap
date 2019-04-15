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

func FTP(options ...services.ServicerFunc) services.Servicer {

	store, err := getStorage()
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

	opts := &ServerOpts{
		Auth: &User{
			users: map[string]string{
				"anonymous": "anonymous",
			},
		},
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

type Opts struct {
	Banner string `toml:"banner"`

	PsvPortRange string `toml:"passive-port-range"`

	ServerName string `toml:"name"`
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
		for msg := range s.recv {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ftp"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ftp.sessionid", ftpConn.sessionid),
				event.Custom("ftp.command", strings.Trim(msg, "\r\n")),
			))
		}
	}()

	ftpConn.Serve()

	return nil
}
