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
package network

import (
	"context"
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

func (sc *socketConfig) AddAddress(a net.Addr) {
	sc.Addresses = append(sc.Addresses, a)
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

func (sl *socketListener) Close() error {
	return nil
}

func (sl *socketListener) Start(ctx context.Context) error {
	for _, address := range sl.Addresses {
		if _, ok := address.(*net.TCPAddr); ok {
			l, err := net.Listen(address.Network(), address.String())
			if err != nil {
				fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
				continue
			}

			log.Infof("Listener started: tcp/%s", address)

			go func() {
				for {
					c, err := l.Accept()
					if err != nil {
						log.Errorf("Error accepting connection: %s", err.Error())
						continue
					}

					sl.ch <- c
				}
			}()
		} else if ua, ok := address.(*net.UDPAddr); ok {
			l, err := net.ListenUDP(address.Network(), ua)
			if err != nil {
				fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
				continue
			}

			log.Infof("Listener started: udp/%s", address)

			go func() {
				for {
					var buf [65535]byte

					n, raddr, err := l.ReadFromUDP(buf[:])
					if err != nil {
						log.Error("Error reading udp:", err.Error())
						continue
					}

					sl.ch <- &listener.DummyUDPConn{
						Buffer: buf[:n],
						Laddr:  l.LocalAddr(),
						Raddr:  raddr,
						Fn:     l.WriteToUDP,
					}
				}
			}()
		}
	}

	return nil
}

func (sl *socketListener) Accept() (net.Conn, error) {
	c := <-sl.ch
	return c, nil
}
