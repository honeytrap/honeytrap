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
package mongodb

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

/*


[service.mongodb]
type="mongodb"
version="3.4.2"
dbs=[ {Name="My_DB", SizeOnDisk="8192", Empty="false"},
      {Name="example", SizeOnDisk="32768", Empty="false"}, ]

[[port]]
port="tcp/27017"
services=["mongodb"]

*/

var (
	log = logging.MustGetLogger("services/mongodb")
	_   = services.Register("mongodb", Mongodb)
)

func Mongodb(options ...services.ServicerFunc) services.Servicer {

	s := &mongodbService{
		mongodbServiceConfig: mongodbServiceConfig{
			Version: "3.4.2",
			Dbs: []Db{
				{"admin", "32768", "false", []User{
					{"user", "pencil", "9QWKl3mbUxIkOafHoI1nbQ=="},
				},
				},

				// {"config", "12288", "false"},
				// {"local", "122880", "false"},
			},
			authActivated: false,
			logged:        false,
			itercounts:    10000,
		},
	}
	// TODO parse/check the config

	for _, o := range options {
		o(s)
	}
	return s
}

type mongodbServiceConfig struct {
	Version     string `toml:"version"`
	Dbs         `toml:"dbs"`
	clAddr      net.Addr
	versionBoth string
	port        int
	logged      bool
	username    string
	Client
	authActivated bool
	itercounts    int
}

type Dbs []Db

type Db struct {
	Name       string
	SizeOnDisk string
	Empty      string
	Users
}

type Client struct {
	username string
	clNonce  string
	clProof  string
	svNonce  string
	svProof  string
	cbNonce  string
	password string
	salt     string
}

type Users []User

type User struct {
	username string
	password string
	salt     string
}

type Client struct {
	username string
	clNonce  string
	clProof  string
	svNonce  string
	svProof  string
	cbNonce  string
	password string
	salt     string
}

type Users []User

type User struct {
	username string
	password string
	salt     string
}

type mongodbService struct {
	mongodbServiceConfig
	ch pushers.Channel
}

func (s *mongodbService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *mongodbService) Handle(ctx context.Context, conn net.Conn) error {

	defer conn.Close()

	br := bufio.NewReader(conn)
	s.port = conn.RemoteAddr().(*net.TCPAddr).Port
	s.clAddr = conn.RemoteAddr()

	for {

		buff := make([]byte, 1024)
		n, err := br.Read(buff[:])
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		bb := bytes.NewBuffer(buff[:n])
		payload := bb.Bytes()

		response, ev := s.reqHandler(bb)

		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("mongodb"),
			event.Type("mongodb-request"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Payload(payload),
			event.CopyFrom(ev),
		))
		conn.Write(response)
		br.Reset(conn)
	}
	return nil
}

/*

"user + ':mongo:' + pencil"

Ho+Vgk7qvUOKUwuWLIWg4l/9SraGMHEE


























*/
