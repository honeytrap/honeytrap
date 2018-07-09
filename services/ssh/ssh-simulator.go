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
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/honeytrap/honeytrap/services/decoder"

	"bytes"

	"github.com/rs/xid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	_ = services.Register("ssh-simulator", Simulator)
)

var motd = `Welcome to Ubuntu 16.04.1 LTS (GNU/Linux 4.4.0-31-generic x86_64)

* Documentation:  https://help.ubuntu.com
* Management:     https://landscape.canonical.com
* Support:        https://ubuntu.com/advantage

524 packages can be updated.
270 updates are security updates.


----------------------------------------------------------------
Ubuntu 16.04.1 LTS                          built 2016-12-10
----------------------------------------------------------------
last login: Sun Nov 19 19:40:44 2017 from 172.16.84.1
`

func Simulator(options ...services.ServicerFunc) services.Servicer {
	s, err := getStorage()
	if err != nil {
		log.Errorf("Could not initialize storage: ", err.Error())
	}

	banner := "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2"

	service := &sshSimulatorService{
		key:          s.PrivateKey(),
		Banner:       banner,
		MOTD:         motd,
		MaxAuthTries: -1,
		Credentials: []string{
			"*",
		},
	}

	for _, o := range options {
		o(service)
	}

	return service
}

type sshSimulatorService struct {
	c pushers.Channel

	Banner string `toml:"banner"`
	MOTD   string `toml:"motd"`

	MaxAuthTries int `toml:"max-auth-tries"`

	Credentials []string    `toml:"credentials"`
	key         *privateKey `toml:"private-key"`
}

func (s *sshSimulatorService) CanHandle(payload []byte) bool {
	return bytes.HasPrefix(payload, []byte("SSH"))
}

func (s *sshSimulatorService) SetChannel(c pushers.Channel) {
	s.c = c
}

type payloadDecoder struct {
	decoder.Decoder
}

func (pd *payloadDecoder) String() string {
	length := int(pd.Uint32())
	payload := pd.Copy(length)
	return string(payload)
}

func PayloadDecoder(payload []byte) *payloadDecoder {
	return &payloadDecoder{
		decoder.NewDecoder(payload),
	}
}

func (s *sshSimulatorService) Handle(ctx context.Context, conn net.Conn) error {
	id := xid.New()

	var connOptions event.Option = nil

	if ec, ok := conn.(*event.Conn); ok {
		connOptions = ec.Options()
	}

	config := ssh.ServerConfig{
		ServerVersion: s.Banner,
		MaxAuthTries:  s.MaxAuthTries,
		PublicKeyCallback: func(cm ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("publickey-authentication"),
				connOptions,
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
				connOptions,
				event.SourceAddr(cm.RemoteAddr()),
				event.DestinationAddr(cm.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.username", cm.User()),
				event.Custom("ssh.password", string(password)),
			))

			for _, credential := range s.Credentials {
				if credential == "*" {
					return nil, nil
				}

				parts := strings.Split(credential, ":")
				if len(parts) != 2 {
					continue
				}

				if cm.User() == parts[0] && string(password) == parts[1] {
					log.Debug("User authenticated successfully. user=%s password=%s", cm.User(), string(password))
					return nil, nil
				}
			}

			return nil, fmt.Errorf("Password rejected for %q", cm.User())
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
		sconn.Close()
	}()

	go ssh.DiscardRequests(reqs)

	// https://tools.ietf.org/html/rfc4254
	for newChannel := range chans {
		switch newChannel.ChannelType() {
		case "session":
			// handleSession()
		case "forwarded-tcpip":
			decoder := PayloadDecoder(newChannel.ExtraData())

			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("ssh-channel"),
				connOptions,
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.channel-type", newChannel.ChannelType()),
				event.Custom("ssh.forwarded-tcpip.address-that-was-connected", decoder.String()),
				event.Custom("ssh.forwarded-tcpip.port-that-was-connected", fmt.Sprintf("%d", decoder.Uint32())),
				event.Custom("ssh.forwarded-tcpip.originator-host", decoder.String()),
				event.Custom("ssh.forwarded-tcpip.originator-port", fmt.Sprintf("%d", decoder.Uint32())),
				event.Payload(newChannel.ExtraData()),
			))

			newChannel.Reject(ssh.UnknownChannelType, "not allowed")
			continue
		case "direct-tcpip":
			decoder := PayloadDecoder(newChannel.ExtraData())

			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("ssh-channel"),
				connOptions,
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.channel-type", newChannel.ChannelType()),
				event.Custom("ssh.direct-tcpip.host-to-connect", decoder.String()),
				event.Custom("ssh.direct-tcpip.port-to-connect", fmt.Sprintf("%d", decoder.Uint32())),
				event.Custom("ssh.direct-tcpip.originator-host", decoder.String()),
				event.Custom("ssh.direct-tcpip.originator-port", fmt.Sprintf("%d", decoder.Uint32())),
				event.Payload(newChannel.ExtraData()),
			))

			newChannel.Reject(ssh.UnknownChannelType, "not allowed")
			continue
		default:
			s.c.Send(event.New(
				services.EventOptions,
				event.Category("ssh"),
				event.Type("ssh-channel"),
				connOptions,
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("ssh.sessionid", id.String()),
				event.Custom("ssh.channel-type", newChannel.ChannelType()),
				event.Payload(newChannel.ExtraData()),
			))

			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			log.Debugf("Unknown channel type: %s\n", newChannel.ChannelType())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err == io.EOF {
			continue
		} else if err != nil {
			log.Errorf("Could not accept server channel: %s", err.Error())
			continue
		}

		func() {
			for req := range requests {
				log.Debugf("Request: %s %s %s %s\n", channel, req.Type, req.WantReply, req.Payload)

				options := []event.Option{
					services.EventOptions,
					event.Category("ssh"),
					event.Type("ssh-request"),
					connOptions,
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("ssh.sessionid", id.String()),
					event.Custom("ssh.request-type", req.Type),
					event.Custom("ssh.payload", req.Payload),
				}

				b := false

				switch req.Type {
				case "shell":
					b = true
				case "pty-req":
					b = true
				case "env":
					b = true

					decoder := PayloadDecoder(req.Payload)

					payloads := []string{}

					for {
						if decoder.Available() == 0 {
							break
						}

						payload := decoder.String()
						payloads = append(payloads, payload)
					}

					options = append(options, event.Custom("ssh.env", payloads))
				case "tcpip-forward":
					decoder := PayloadDecoder(req.Payload)

					options = append(options, event.Custom("ssh.tcpip-forward.address-to-bind", decoder.String()))
					options = append(options, event.Custom("ssh.tcpip-forward.port-to-bind", fmt.Sprintf("%d", decoder.Uint32())))
				case "exec":
					b = true

					decoder := PayloadDecoder(req.Payload)

					payloads := []string{}

					for {
						if decoder.Available() == 0 {
							break
						}

						payload := decoder.String()
						payloads = append(payloads, payload)
					}

					options = append(options, event.Custom("ssh.exec", payloads))
				case "subsystem":
					b = true

					decoder := PayloadDecoder(req.Payload)
					options = append(options, event.Custom("ssh.subsystem", decoder.String()))
				default:
					log.Errorf("Unsupported request type=%s payload=%s", req.Type, string(req.Payload))
				}

				if !b {
					// no reply
				} else if err := req.Reply(b, nil); err != nil {
					log.Errorf("wantreply: ", err)
				}

				s.c.Send(event.New(
					options...,
				))

				func() {
					if req.Type == "shell" {
						defer channel.Close()

						// should only be started in req.Type == shell
						twrc := NewTypeWriterReadCloser(channel)
						var wrappedChannel io.ReadWriteCloser = twrc

						prompt := "root@host:~$ "

						term := terminal.NewTerminal(wrappedChannel, prompt)

						term.Write([]byte(s.MOTD))

						for {
							line, err := term.ReadLine()
							if err == io.EOF {
								return
							} else if err != nil {
								log.Errorf("Error reading from connection: %s", err.Error())
								return
							}

							if line == "exit" {
								channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
								return
							}

							if line == "" {
								continue
							}

							s.c.Send(event.New(
								services.EventOptions,
								event.Category("ssh"),
								event.Type("ssh-channel"),
								event.SourceAddr(conn.RemoteAddr()),
								event.DestinationAddr(conn.LocalAddr()),
								event.Custom("ssh.sessionid", id.String()),
								event.Custom("ssh.command", line),
							))

							term.Write([]byte(fmt.Sprintf("%s: command not found\n", line)))
						}
					} else if req.Type == "exec" {
						defer channel.Close()

						channel.Write([]byte(fmt.Sprintf("%s: command not found\n", "ls")))
						channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						return
					} else {
					}
				}()
			}
		}()
	}

	return nil
}
