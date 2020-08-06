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
package tun

import (
	"context"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"

	water "github.com/songgao/water"
)

var (
	SensorTun = event.Sensor("tun")

	EventCategoryUDP = event.Category("udp")
)

var log = logging.MustGetLogger("listener/tun")

var (
	_ = listener.Register("tun", New)
)

type tunConfig struct {
	Addresses []string
}

func (nc *tunConfig) AddAddress(v string) {
	nc.Addresses = append(nc.Addresses, v)
}

type tunListener struct {
	tunConfig

	ch chan net.Conn

	eb pushers.Channel

	net.Listener
}

func (l *tunListener) SetChannel(eb pushers.Channel) {
	l.eb = eb
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	l := tunListener{
		tunConfig: tunConfig{},
		eb:        pushers.MustDummy(),
		ch:        ch,
	}

	for _, option := range options {
		option(&l)
	}

	return &l, nil
}

func (l *tunListener) Close() error {
	return nil
}

func (l *tunListener) Start(ctx context.Context) error {
	config := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(config)
	if err != nil {
		return err
	}

	log.Infof("Created new tun interface Name: %s\n", ifce.Name())

	packet := make([]byte, 2000)
	go func() {
		for {
			n, err := ifce.Read(packet)
			if err != nil {
				log.Errorf("Error reading packet: %s", err.Error())
				return
			}

			l.eb.Send(event.New(
				SensorTun,
				EventCategoryUDP,
				event.Payload(packet[:n]),
			))
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
						log.errorf("error accepting connection: %s", err.error())
						continue
					}

					ch <- c
				}
			}()
		}*/
}

func (l *tunListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
