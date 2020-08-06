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
package tap

import (
	"context"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"

	water "github.com/songgao/water"

	"github.com/songgao/packets/ethernet"
)

var log = logging.MustGetLogger("honeytrap:listener:tap")

var (
	_ = listener.Register("tap", New)
)

type tapConfig struct {
	Addresses []string
}

func (nc *tapConfig) AddAddress(v string) {
	nc.Addresses = append(nc.Addresses, v)
}

type tapListener struct {
	tapConfig

	ch chan net.Conn

	eb pushers.Channel

	net.Listener
}

func (l *tapListener) SetChannel(eb pushers.Channel) {
	l.eb = eb
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	l := tapListener{
		tapConfig: tapConfig{},
		ch:        ch,
	}

	for _, option := range options {
		option(&l)
	}

	return &l, nil
}

func (l *tapListener) Close() error {
	return nil
}

func (l *tapListener) Start(ctx context.Context) error {
	config := water.Config{
		DeviceType: water.TAP,
	}

	// config.Name = "O_O"

	ifce, err := water.New(config)
	if err != nil {
		return err
	}

	var frame ethernet.Frame

	go func() {
		for {
			fmt.Println("BLA")

			frame.Resize(1500)

			n, err := ifce.Read([]byte(frame))
			if err != nil {
				log.Fatal(err)
			}

			frame = frame[:n]

			log.Debugf("Dst: %s\n", frame.Destination())
			log.Debugf("Src: %s\n", frame.Source())
			log.Debugf("Ethertype: % x\n", frame.Ethertype())
			log.Debugf("Payload: % x\n", frame.Payload())
		}
	}()

	return nil
	/*
		for _, address := range cfg.Addresses {
			l, err := net.Listen("tcp", address)
			if err != nil {
				return nil, err
			}

			log.Infof("Listener started: %s", address)

			go func() {
				for {
					c, err := l.Accept()
					if err != nil {
						log.Errorf("Error accepting connection: %s", err.Error())
						continue
					}

					ch <- c
				}
			}()
		}*/
}

func (l *tapListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
