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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"golang.org/x/crypto/ssh"
)

var (
	_ = Register("ssh-auth", SSHAuth)
)

func generateKey() (*PrivateKey, error) {
	// TODO: cache generated key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		return nil, cerr
	}

	privder := x509.MarshalPKCS1PrivateKey(priv)

	privblk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privder,
	}

	privateBytes := pem.EncodeToMemory(&privblk)

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{private}, nil
}

// PrivateKey holds the ssh.Signer instance to unsign received data.
type PrivateKey struct {
	ssh.Signer
}

// UnmarshalText unmarshalls the giving text as the Signers data.
func (t *PrivateKey) UnmarshalText(data []byte) (err error) {
	private, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return err
	}

	(*t) = PrivateKey{private}
	return err
}

func SSHAuth(options ...ServicerFunc) Servicer {
	key, err := generateKey()
	if err != nil {
		log.Errorf("Could not generate ssh key: %s", err.Error())
		return nil
	}

	// TODO(nl5887): from configuration file
	banner := "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"

	s := &sshAuthService{
		key: key,
		//Banner: banner,
	}

	s.config = ssh.ServerConfig{
		ServerVersion: banner,
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				EventOptions,
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
				EventOptions,
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

	s.config.AddHostKey(key)

	for _, o := range options {
		o(s)
	}

	return s
}

type sshAuthService struct {
	c pushers.Channel

	Banner string `toml:"banner"`

	key    *PrivateKey
	config ssh.ServerConfig
}

func (s *sshAuthService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *sshAuthService) Handle(conn net.Conn) error {
	defer conn.Close()

	sconn, chans, reqs, err := ssh.NewServerConn(conn, &s.config)
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
