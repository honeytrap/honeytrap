// +build linux
// +build !arm

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
package netstack

import (
	"context"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"

	"strings"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/link/fdbased"
	"github.com/google/netstack/tcpip/link/rawfile"
	"github.com/google/netstack/tcpip/link/tun"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/network/ipv6"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/tcp"
	"github.com/google/netstack/waiter"
)

var (
	SensorNetstack = event.Sensor("netstack")
)

var log = logging.MustGetLogger("listener/netstack")

var (
	_ = listener.Register("netstack", New)
)

type netstackConfig struct {
	Addresses []net.Addr

	Addr       string   `toml:"addr"`
	Interfaces []string `toml:"interfaces"`
}

func (nc *netstackConfig) AddAddress(a net.Addr) {
	nc.Addresses = append(nc.Addresses, a)
}

type netstackListener struct {
	netstackConfig

	ch chan net.Conn

	eb pushers.Channel
}

func (d *netstackListener) SetChannel(eb pushers.Channel) {
	d.eb = eb
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	l := netstackListener{
		netstackConfig: netstackConfig{},
		eb:             pushers.MustDummy(),
		ch:             ch,
	}

	for _, option := range options {
		option(&l)
	}

	return &l, nil
}

func (nl *netstackListener) Start(ctx context.Context) error {
	// Parse the IP address. Support both ipv4 and ipv6.
	parsedAddr := net.ParseIP(nl.Addr)
	if parsedAddr == nil {
		return fmt.Errorf("Bad IP address: %v", nl.Addr)
	}

	var addr tcpip.Address
	var proto tcpip.NetworkProtocolNumber
	if parsedAddr.To4() != nil {
		addr = tcpip.Address(parsedAddr.To4())
		proto = ipv4.ProtocolNumber
	} else if parsedAddr.To16() != nil {
		addr = tcpip.Address(parsedAddr.To16())
		proto = ipv6.ProtocolNumber
	} else {
		return fmt.Errorf("Unknown IP type: %v", nl.Addr)
	}

	// Create the stack with ip and tcp protocols, then add a tun-based
	// NIC and address.
	s := stack.New([]string{ipv4.ProtocolName, ipv6.ProtocolName}, []string{tcp.ProtocolName})

	// todo: only one interface supported now
	tunName := nl.Interfaces[0]

	mtu, err := rawfile.GetMTU(tunName)
	if err != nil {
		return err
	}

	fd, err := tun.Open(tunName)
	if err != nil {
		log.Fatal(err)
	}

	linkID := fdbased.New(fd, mtu, nil)
	if err := s.CreateNIC(1, linkID); err != nil {
		return fmt.Errorf(err.String())
	}

	if err := s.AddAddress(1, proto, addr); err != nil {
		return fmt.Errorf(err.String())
	}

	// Add default route.
	s.SetRouteTable([]tcpip.Route{
		{
			Destination: tcpip.Address(strings.Repeat("\x00", len(addr))),
			Mask:        tcpip.Address(strings.Repeat("\x00", len(addr))),
			Gateway:     "",
			NIC:         1,
		},
	})

	for _, address := range nl.Addresses {
		go func() {
			if ta, ok := address.(*net.TCPAddr); ok {
				// Create TCP endpoint, bind it, then start listening.
				var wq waiter.Queue
				ep, e := s.NewEndpoint(tcp.ProtocolNumber, proto, &wq)
				if err != nil {
					log.Fatal(e)
				}

				defer ep.Close()

				if err := ep.Bind(tcpip.FullAddress{0, "", uint16(ta.Port)}, nil); err != nil {
					log.Fatal("Bind failed: ", err)
				}

				if err := ep.Listen(10); err != nil {
					log.Fatal("Listen failed: ", err)
				}

				// Wait for connections to appear.
				waitEntry, notifyCh := waiter.NewChannelEntry(nil)
				wq.EventRegister(&waitEntry, waiter.EventIn)
				defer wq.EventUnregister(&waitEntry)

				for {
					n, wq, err := ep.Accept()
					if err == nil {
					} else if err == tcpip.ErrWouldBlock {
						<-notifyCh
						continue
					} else {
						log.Fatal("Accept() failed:", err)
					}

					nl.ch <- newConn(wq, n)
				}
			}
		}()

	}
	return nil
}

func (l *netstackListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
