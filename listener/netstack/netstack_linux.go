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

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"

	"strings"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/adapters/gonet"
	"github.com/google/netstack/tcpip/link/fdbased"
	"github.com/google/netstack/tcpip/link/rawfile"
	"github.com/google/netstack/tcpip/link/tun"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/network/ipv6"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/tcp"
	"github.com/google/netstack/tcpip/transport/udp"
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

func (l *netstackListener) SetChannel(eb pushers.Channel) {
	l.eb = eb
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

func (l *netstackListener) Start(ctx context.Context) error {
	// Parse the IP address. Support both ipv4 and ipv6.
	parsedAddr := net.ParseIP(l.Addr)
	if parsedAddr == nil {
		return fmt.Errorf("Bad IP address: %v", l.Addr)
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
		return fmt.Errorf("Unknown IP type: %v", l.Addr)
	}

	// Create the stack with ip and tcp protocols, then add a tun-based
	// NIC and address.
	clock := &tcpip.StdClock{}
	s := stack.New(clock, []string{ipv4.ProtocolName, ipv6.ProtocolName}, []string{tcp.ProtocolName, udp.ProtocolName})

	// todo: only one interface supported now
	tunName := l.Interfaces[0]

	mtu, err := rawfile.GetMTU(tunName)
	if err != nil {
		return err
	}

	var fd int
	if false {
		fd, err = tun.OpenTAP(tunName)
	} else {
		fd, err = tun.Open(tunName)
	}
	if err != nil {
		log.Fatal(err)
	}

	linkID := fdbased.New(&fdbased.Options{
		FD:  fd,
		MTU: mtu,
	})
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

	for _, address := range l.Addresses {
		go func(address net.Addr) {
			if ta, ok := address.(*net.TCPAddr); ok {
				listener, err := gonet.NewListener(s, tcpip.FullAddress{
					NIC:  0,
					Addr: tcpip.Address(ta.IP),
					Port: uint16(ta.Port),
				}, proto)
				if err != nil {
					log.Fatal(err)
				}

				defer listener.Close()

				for {
					conn, err := listener.Accept()
					if err != nil {
						log.Error(err.Error())
						continue
					}

					l.ch <- conn
				}
			} else if ua, ok := address.(*net.UDPAddr); ok {
				pc, err := gonet.NewPacketConn(s, tcpip.FullAddress{
					NIC:  0,
					Addr: tcpip.Address(ua.IP),
					Port: uint16(ua.Port),
				}, proto)
				if err != nil {
					fmt.Println(color.RedString("Error starting udp listener: %s", err.Error()))
					return
				}

				log.Infof("Listener started: udp/%s", address)

				go func() {
					for {
						var buf [65535]byte

						n, raddr, err := pc.ReadFrom(buf[:])
						if err != nil {
							log.Error("Error reading udp:", err.Error())
							continue
						}

						l.ch <- &listener.DummyUDPConn{
							Buffer: buf[:n],
							Laddr:  ua,
							Raddr:  raddr.(*net.UDPAddr),
							Fn: func(b []byte, addr *net.UDPAddr) (int, error) {
								return pc.WriteTo(b, addr)
							},
						}
					}
				}()
			}
		}(address)

	}
	return nil
}

func (l *netstackListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
