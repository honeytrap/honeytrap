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
	"encoding/binary"
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

type conn2 struct {
	net.Conn

	al *agentListener
}

func (c *conn2) Handshake() error {
	buff := make([]byte, 2)

	log.Debugf("Agent connecting from remote address: %s", c.RemoteAddr())

	n, err := c.Read(buff[:])
	if err != nil {
		return err
	}

	size := binary.LittleEndian.Uint16(buff)

	buff = make([]byte, size)
	n, err = c.Read(buff[:])
	if err != nil {
		return err
	}

	h := Handshake{}

	if err := h.UnmarshalBinary(buff[:n]); err != nil {
		return err
	}

	fmt.Println("Handhake received")

	p := HandshakeResponse{
		c.al.Addresses,
	}

	if data, err := p.MarshalBinary(); err == nil {
		buff := make([]byte, 2)
		binary.LittleEndian.PutUint16(buff[0:2], uint16(len(data)))

		c.Write(buff)
		c.Write(data)
	} else {
		return err
	}

	return nil
}

func (ac conn2) send(o encoding.BinaryMarshaler) error {
	data, err := o.MarshalBinary()
	if err != nil {
		return err
	}

	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff[0:2], uint16(len(data)))

	if _, err := ac.Conn.Write(buff); err != nil {
		return err
	}

	if _, err := ac.Conn.Write(data); err != nil {
		return err
	}

	return nil
}

func (sl *agentListener) serv(c conn2) {
	conns := Connections{}

	if err := c.Handshake(); err == io.EOF {
		return
	} else if err != nil {
		fmt.Println(color.RedString(err.Error()))
		return
	}

	fmt.Println("Agent connected...")
	defer fmt.Println("Agent disconnected...")

	out := make(chan ReadWrite)
	go func(ch chan ReadWrite) {
		for p := range ch {
			err := c.send(p)
			if err != nil {
				log.Error("Error marshaling hello: %s", err.Error())
				break
			}
		}
	}(out)

	func(ch chan ReadWrite) {
		for {
			buff := make([]byte, 2)

			n, err := c.Read(buff[:])
			if err == io.EOF {
				log.Errorf("REMCO EOF", err.Error())
			} else if err != nil {
				log.Errorf("REMCO ", err.Error())
				return
			}

			size := binary.LittleEndian.Uint16(buff)

			buff = make([]byte, size)
			n, err = c.Read(buff[:])
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			if int(buff[0]) == TypeHello {
				h := Hello{}

				err := h.UnmarshalBinary(buff[:n])
				if err != nil {
					log.Errorf(err.Error())
					return
				}

				ac := &agentConnection{
					Laddr: h.Laddr,
					Raddr: h.Raddr,
					in:    make(chan []byte),
					out:   ch,
				}

				conns = append(conns, ac)

				sl.ch <- ac

			} else if int(buff[0]) == TypeReadWrite {
				r := ReadWrite{}

				err := r.UnmarshalBinary(buff[:n])
				if err != nil {
					log.Errorf(err.Error())
					return
				}

				conn := conns.Get(r.Laddr, r.Raddr)
				if conn == nil {
					continue
				}

				conn.in <- r.Payload
			} else if int(buff[0]) == TypeEOF {
				// read
				fmt.Println("EOF")
				r := EOF{}

				err := r.UnmarshalBinary(buff[:n])
				if err != nil {
					log.Errorf(err.Error())
					return
				}

				conn := conns.Get(r.Laddr, r.Raddr)
				if conn == nil {
					continue
				}

				// remove connection

				conn.closed = true
				close(conn.in)
			} else if int(buff[0]) == TypePing {
				log.Debugf("Received ping from agent: %s", c.RemoteAddr())
			}
		}
	}(out)

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

			sl.serv(conn2{
				c,
				sl,
			})
		}
	}()

	return nil
}

func (sl *agentListener) Accept() (net.Conn, error) {
	c := <-sl.ch
	return c, nil
}
