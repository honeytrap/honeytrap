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
package agent

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"net"
	"runtime"

	bus "github.com/dutchcoders/gobus"
	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/messages"
	"github.com/mimoo/disco/libdisco"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("listeners/agent")

//  Register the listener
var (
	_ = listener.Register("agent", New)
)

type agentListener struct {
	agentConfig

	ch        chan net.Conn
	Addresses []net.Addr

	net.Listener
}

type agentConfig struct {
	Listen string `toml:"listen"`
}

// AddAddress will add the addresses to listen to
func (al *agentListener) AddAddress(a net.Addr) {
	al.Addresses = append(al.Addresses, a)
}

// New will initialize the agent listener
func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	l := agentListener{
		agentConfig: agentConfig{},
		ch:          ch,
	}

	for _, option := range options {
		option(&l)
	}

	return &l, nil
}

func (al *agentListener) serv(c *conn2) {
	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Errorf("Error: %s", err)
			log.Errorf("Stack of %d bytes: %s\n", count, string(trace))
			return
		}
	}()

	log.Debugf("Agent connecting from remote address: %s", c.RemoteAddr())

	p, err := c.receive()
	if err == io.EOF {
		return
	}
	if err != nil {
		log.Errorf("Error receiving object: %s", err.Error())
		return
	}
	h, ok := p.(*Handshake)
	if !ok {
		log.Errorf("Expected handshake from Agent")
		return
	}

	version := h.Version
	shortCommitID := h.ShortCommitID
	token := h.Token

	log.Infof(color.YellowString("Agent connected (version=%s, commitid=%s, token=%s)...", version, shortCommitID, token))
	defer log.Infof(color.YellowString("Agent disconnected"))

	bus.Emit("agent-connect", &messages.AgentConnect{
		Agent: &messages.Agent{
			Version:       version,
			ShortCommitID: shortCommitID,
			Token:         token,
			RemoteAddr:    c.RemoteAddr().String(),
		},
	})

	defer bus.Emit("agent-disconnect", &messages.AgentDisconnect{
		Agent: &messages.Agent{
			Version:       version,
			ShortCommitID: shortCommitID,
			Token:         token,
			RemoteAddr:    c.RemoteAddr().String(),
		},
	})

	c.send(HandshakeResponse{
		al.Addresses,
	})

	out := make(chan interface{})

	conns := Connections{}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()

		c.Close()

		go func() {
			// drain
			for _ = range out {
			}
		}()

		conns.Each(func(conn *agentConnection) {
			conn.Close()
		})

		close(out)
	}()

	go func() {
		for {
			select {
			case p := <-out:
				if bm, ok := p.(encoding.BinaryMarshaler); !ok {
					log.Errorf("Error marshalling object")
					return
				} else if err := c.send(bm); err != nil {
					log.Errorf("Error sending object: %s", err.Error())
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		o, err := c.receive()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Errorf("Error receiving object: %s", err.Error())
			return
		}

		switch v := o.(type) {
		case *Hello:
			ac := &agentConnection{
				Laddr: v.Laddr,
				Raddr: v.Raddr,
				in:    make(chan []byte),
				out:   out,
			}

			conns.Add(ac)

			conn := event.WithConn(ac, event.Custom("agent", token))
			al.ch <- conn
		case *ReadWriteTCP:
			conn := conns.Get(v.Laddr, v.Raddr)
			if conn == nil {
				continue
			}

			conn.receive(v.Payload)
		case *ReadWriteUDP:
			al.ch <- &listener.DummyUDPConn{
				Buffer: v.Payload,
				Laddr:  v.Laddr.(*net.UDPAddr),
				Raddr:  v.Raddr.(*net.UDPAddr),
				Fn: func(b []byte, addr *net.UDPAddr) (int, error) {
					payload := make([]byte, len(b))
					copy(payload, b)

					p := ReadWriteUDP{
						Laddr:   v.Laddr,
						Raddr:   v.Raddr,
						Payload: payload[:],
					}

					out <- p
					return len(b), nil
				},
			}
		case *EOF:
			conn := conns.Get(v.Laddr, v.Raddr)
			if conn == nil {
				continue
			}

			conns.Delete(conn)

			conn.Close()
		case *Ping:
			log.Debugf("Received ping from agent: %s", c.RemoteAddr())

			bus.Emit("agent-ping", messages.AgentPing{
				Agent: &messages.Agent{
					Version:       version,
					ShortCommitID: shortCommitID,
					Token:         token,
					RemoteAddr:    c.RemoteAddr().String(),
				},
			})
		}
	}
}

func (al *agentListener) Close() error {
	return nil
}

// Start the listener
func (al *agentListener) Start(ctx context.Context) error {
	storage, err := Storage()
	if err != nil {
		return err
	}

	keyPair, err := storage.KeyPair()
	if err != nil {
		return err
	}

	fmt.Println(color.YellowString("Honeytrap Agent Server public key: %s", keyPair.ExportPublicKey()))

	serverConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_NK,
		KeyPair:          keyPair,
	}

	listen := ":1339"
	if al.Listen != "" {
		listen = al.Listen
	}

	listener, err := libdisco.Listen("tcp", listen, &serverConfig)
	if err != nil {
		fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
		return err
	}

	log.Infof("Listener started: %s", listen)

	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				log.Errorf("Error accepting connection: %s", err.Error())
				continue
			}

			go al.serv(Conn2(c))
		}
	}()

	return nil
}

// Accept a new connection
func (al *agentListener) Accept() (net.Conn, error) {
	c := <-al.ch
	return c, nil
}
