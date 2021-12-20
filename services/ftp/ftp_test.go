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
	"net"
	"os"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/storage"
)

const (
	user     = "anonymous"
	password = "anonymous"
)

var (
	clt, srv net.Conn
)

func TestMain(m *testing.M) {
	storage.SetDataDir("/tmp")
	os.Exit(m.Run())
}

func TestFTP(t *testing.T) {

	//Setup client and server
	clt, srv = net.Pipe()
	defer clt.Close()
	defer srv.Close()

	s := FTP().(*ftpService)

	c, _ := pushers.Dummy()
	s.SetChannel(c)

	//set user
	s.server.Auth = User{user: password}

	//Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Error(err)
		}
	}(srv)

	client, err := Connect(clt)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Login(user, password); err != nil {
		t.Errorf("Could not login user: %s password: %s", user, password)
	}

	if err := client.Quit(); err != nil {
		t.Errorf("Error with Quit: %s", err.Error())
	}
}
