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
