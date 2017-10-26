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
package network

import (
	"fmt"
	"net"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/listener"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("listeners/socket")

var (
	_ = listener.Register("socket", New)
)

type socketListener struct {
	socketConfig

	ch chan net.Conn

	net.Listener
}

type socketConfig struct {
	Addresses []net.Addr
}

func (nc *socketConfig) AddAddress(a net.Addr) {
	nc.Addresses = append(nc.Addresses, a)
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	l := socketListener{
		socketConfig: socketConfig{},
		ch:           ch,
	}

	for _, option := range options {
		option(&l)
	}

	return &l, nil
}

func (nl *socketListener) Start() error {
	for _, address := range nl.Addresses {
		if _, ok := address.(*net.TCPAddr); ok {
			l, err := net.Listen(address.Network(), address.String())
			if err != nil {
				fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
				continue
			}

			log.Infof("Listener started: %s", address)

			go func() {
				for {
					c, err := l.Accept()
					if err != nil {
						log.Errorf("Error accepting connection: %s", err.Error())
						continue
					}

					nl.ch <- c
				}
			}()
		} else if ua, ok := address.(*net.UDPAddr); ok {
			l, err := net.ListenUDP(address.Network(), ua)
			if err != nil {
				fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
				continue
			}

			log.Infof("Listener started: %s", address)

			go func() {

				for {
					var buf [65535]byte

					n, raddr, err := l.ReadFromUDP(buf[:])
					if err != nil {
						log.Error("Error reading udp:", err.Error())
						continue
					}

					nl.ch <- &listener.DummyUDPConn{
						Buffer: buf[:n],
						Laddr:  ua,
						Raddr:  raddr,
						C:      l,
					}
				}
			}()
		}
	}

	return nil
}

func (l *socketListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
