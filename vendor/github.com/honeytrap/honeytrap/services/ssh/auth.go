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
package ssh

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"

	"golang.org/x/crypto/ssh"
)

var (
	_ = services.Register("ssh-auth", Auth)
)

func Auth(options ...services.ServicerFunc) services.Servicer {
	s, err := getStorage()
	if err != nil {
		log.Errorf("Could not initialize storage: %s", err.Error())
	}

	banner := "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"

	srvc := &sshAuthService{
		Key:    s.PrivateKey(),
		Banner: banner,
	}

	for _, o := range options {
		o(srvc)
	}

	return srvc
}

type sshAuthService struct {
	c pushers.Channel

	Banner string `toml:"banner"`

	Key    *privateKey `toml:"private-key"`
	config ssh.ServerConfig
}

func (s *sshAuthService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *sshAuthService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	config := ssh.ServerConfig{
		ServerVersion: s.Banner,
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("publickey-authentication"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ssh.publickey-type", key.Type()),
				event.Custom("ssh.publickey", hex.EncodeToString(key.Marshal())),
			))

			return nil, errors.New("Unknown key")
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("password-authentication"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ssh.username", conn.User()),
				event.Custom("ssh.password", string(password)),
			))

			return nil, errors.New("Unknown username or password")
		},
	}

	config.AddHostKey(s.Key)

	sconn, chans, reqs, err := ssh.NewServerConn(conn, &config)
	if err == io.EOF {
		// server closed connection
		return nil
	} else if err != nil {
		return err
	}

	defer sconn.Close()

	go ssh.DiscardRequests(reqs)
	_ = chans

	return nil
}
