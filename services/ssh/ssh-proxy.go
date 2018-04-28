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
package ssh

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"

	"encoding/base64"

	"github.com/rs/xid"
	"golang.org/x/crypto/ssh"
)

var (
	_ = services.Register("ssh-proxy", Proxy)
)

func Proxy(options ...services.ServicerFunc) services.Servicer {
	s, err := getStorage()
	if err != nil {
		log.Errorf("Could not initialize storage: ", err.Error())
	}

	banner := "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"

	service := &sshProxyService{
		key:    s.PrivateKey(),
		Banner: banner,
	}

	for _, o := range options {
		o(service)
	}

	return service
}

type sshProxyService struct {
	c pushers.Channel

	Banner string `toml:"banner"`

	key *privateKey `toml:"private-key"`

	d director.Director
}

func (s *sshProxyService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *sshProxyService) SetDirector(d director.Director) {
	s.d = d
}

func (s *sshProxyService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	var client *ssh.Client

	config := ssh.ServerConfig{
		ServerVersion: s.Banner,
		MaxAuthTries:  -1,
		PublicKeyCallback: func(cm ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("publickey-authentication"),
				event.SourceAddr(cm.RemoteAddr()),
				event.DestinationAddr(cm.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.username", cm.User()),
				event.Custom("ssh.publickey-type", key.Type()),
				event.Custom("ssh.publickey", hex.EncodeToString(key.Marshal())),
			))

			return nil, errors.New("Unknown key")
		},
		PasswordCallback: func(cm ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("password-authentication"),
				event.SourceAddr(cm.RemoteAddr()),
				event.DestinationAddr(cm.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.username", cm.User()),
				event.Custom("ssh.password", string(password)),
			))

			clientConfig := &ssh.ClientConfig{}

			clientConfig.User = cm.User()
			clientConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			}

			clientConfig.Auth = []ssh.AuthMethod{
				ssh.Password(string(password)),
			}

			cconn, err := s.d.Dial(conn)
			if err != nil {
				return nil, err
			}

			c, chans, reqs, err := ssh.NewClientConn(cconn, "", clientConfig)
			if err != nil {
				return nil, err
			}

			log.Debug("User authenticated successfully. user=%s password=%s", cm.User(), string(password))

			client = ssh.NewClient(c, chans, reqs)
			return nil, err
		},
	}

	config.AddHostKey(s.key)

	defer conn.Close()

	sconn, chans, reqs, err := ssh.NewServerConn(conn, &config)
	if err == io.EOF {
		// server closed connection
		return nil
	} else if err != nil {
		return err
	}

	defer func() {
		client.Close()
		sconn.Close()
	}()

	go ssh.DiscardRequests(reqs)

	// https://www.centos.org/docs/5/html/Deployment_Guide-en-US/s1-ssh-conn.html
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			log.Debugf("Unknown channel type: %s\n", newChannel.ChannelType())
			continue
		}

		channel2, requests2, err := client.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
		if err != nil {
			log.Errorf("Could not accept client channel: %s", err.Error())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Errorf("Could not accept server channel: %s", err.Error())
			continue
		}

		s.c.Send(event.New(
			services.EventOptions,
			event.Category("ssh"),
			event.Type("ssh-channel"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("ssh.sessionid", id.String()),
			event.Custom("ssh.channel-type", newChannel.ChannelType()),
		))

		requestFn := func(in <-chan *ssh.Request, dst ssh.Channel) {
			defer dst.Close()

			for req := range in {
				log.Debugf("Request: %s %s %s %s\n", dst, req.Type, req.WantReply, req.Payload)

				b, err := dst.SendRequest(req.Type, req.WantReply, req.Payload)
				if err == io.EOF {
					return
				} else if err != nil {
					log.Errorf("Error sending request: %s", err)
					return
				}

				options := []event.Option{
					services.EventOptions,
					event.Category("ssh"),
					event.Type("ssh-request"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("ssh.sessionid", id.String()),
					event.Custom("ssh.request-type", req.Type),
					event.Custom("ssh.payload", req.Payload),
				}

				switch req.Type {
				case "exit-status":
					fallthrough
				case "shell":
					fallthrough
				case "pty-req":
					fallthrough
				case "env":
					if v, err := base64.StdEncoding.DecodeString(string(req.Payload)); err == nil {
						options = append(options, event.Custom("ssh.env", string(v)))
					}
				case "exec":
					if v, err := base64.StdEncoding.DecodeString(string(req.Payload)); err == nil {
						options = append(options, event.Custom("ssh.exec", string(v)))
					}
				case "subsystem":
					log.Debugf("request type=%s payload=%s", req.Type, string(req.Payload))
				default:
					log.Errorf("Unsupported request type=%s payload=%s", req.Type, string(req.Payload))
				}

				if err := req.Reply(b, nil); err != nil {
					log.Errorf("wantreply: ", err)
				}

				s.c.Send(event.New(
					options...,
				))
			}
		}

		go requestFn(requests, channel2)
		go requestFn(requests2, channel)

		copyFn := func(dst io.ReadWriteCloser, src io.ReadCloser) {
			_, err := io.Copy(dst, src)
			if err == io.EOF {
			} else if err != nil {
				log.Error(err.Error())
			}

			dst.Close()
		}

		var wrappedChannel io.ReadCloser = channel

		twrc := NewTypeWriterReadCloser(channel2)
		var wrappedChannel2 io.ReadCloser = twrc

		go copyFn(channel2, wrappedChannel)
		copyFn(channel, wrappedChannel2)

		s.c.Send(event.New(
			services.EventOptions,
			event.Category("ssh"),
			event.Type("ssh-session"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("ssh.sessionid", id.String()),
			event.Custom("ssh.recording", twrc.String()),
		))

	}

	return nil
}
