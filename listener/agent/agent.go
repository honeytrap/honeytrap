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
package agent

import (
	"encoding"
	"fmt"
	"io"
	"net"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/listener"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("listeners/agent")

var (
	_ = listener.Register("agent", New)
)

type agentListener struct {
	agentConfig

	ch chan net.Conn

	net.Listener
}

type agentConfig struct {
	Addresses []net.Addr
}

func (sc *agentConfig) AddAddress(a net.Addr) {
	sc.Addresses = append(sc.Addresses, a)
}

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

func (sl *agentListener) serv(c *conn2) {
	log.Debugf("Agent connecting from remote address: %s", c.RemoteAddr())

	if p, err := c.receive(); err == io.EOF {
		return
	} else if err != nil {
		log.Errorf("Error receiving object: %s", err.Error())
		return
	} else if _, ok := p.(*Handshake); !ok {
		log.Errorf("Expected handshake from Agent")
		return
	}

	c.send(HandshakeResponse{
		sl.Addresses,
	})

	fmt.Println("Agent connected...")
	defer fmt.Println("Agent disconnected...")

	conns := Connections{}

	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	out := make(chan interface{})

	go func() {
		for p := range out {
			if bm, ok := p.(encoding.BinaryMarshaler); !ok {
				log.Errorf("Error marshalling object")
				break
			} else if err := c.send(bm); err != nil {
				log.Errorf("Error sending object: %s", err.Error())
				break
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

			sl.ch <- ac
		case *ReadWrite:
			conn := conns.Get(v.Laddr, v.Raddr)
			if conn == nil {
				break
			}

			conn.in <- v.Payload
		case *EOF:
			conn := conns.Get(v.Laddr, v.Raddr)
			if conn == nil {
				continue
			}

			conn.Close()
		case *Ping:
			log.Debugf("Received ping from agent: %s", c.RemoteAddr())
		}
	}

	return
}

func (sl *agentListener) Start() error {
	l, err := net.Listen("tcp", ":1339")
	if err != nil {
		fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
		return err
	}

	log.Infof("Listener started: %s", ":1339")

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Errorf("Error accepting connection: %s", err.Error())
				continue
			}

			sl.serv(Conn2(c))
		}
	}()

	return nil
}

func (sl *agentListener) Accept() (net.Conn, error) {
	c := <-sl.ch
	return c, nil
}
