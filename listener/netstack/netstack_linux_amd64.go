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
package netstack

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/vishvananda/netlink"

	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/sniffer"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type netstackConfig struct {
	Addresses []net.Addr

	Addr       string   `toml:"addr"`
	Interfaces []string `toml:"interfaces"`

	Debug bool `toml:"debug"`
}

func (nc *netstackConfig) AddAddress(a net.Addr) {
	nc.Addresses = append(nc.Addresses, a)
}

type netstackListener struct {
	netstackConfig

	s *stack.Stack

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

	if len(l.Interfaces) == 0 {
		return nil, fmt.Errorf("No interface defined")
	} else if len(l.Interfaces) > 1 {
		return nil, fmt.Errorf("Only one interface is supported currently")
	}

	return &l, nil
}

// ipToAddressAndProto converts IP to tcpip.Address and a protocol number.
//
// Note: don't use 'len(ip)' to determine IP version because length is always 16.
func ipToAddressAndProto(ip net.IP) (tcpip.NetworkProtocolNumber, tcpip.Address) {
	if i4 := ip.To4(); i4 != nil {
		return ipv4.ProtocolNumber, tcpip.Address(i4)
	}
	return ipv6.ProtocolNumber, tcpip.Address(ip)
}

// ipToAddress converts IP to tcpip.Address, ignoring the protocol.
func ipToAddress(ip net.IP) tcpip.Address {
	_, addr := ipToAddressAndProto(ip)
	return addr
}

func htons(n uint16) uint16 {
	var (
		high = n >> 8
		ret  = n<<8 + high
	)

	return ret
}

func (l *netstackListener) Start(ctx context.Context) error {
	intfName := l.Interfaces[0]

	mtu, err := rawfile.GetMTU(intfName)
	if err != nil {
		return err
	}

	ifaceLink, err := netlink.LinkByName(intfName)
	if err != nil {
		return fmt.Errorf("unable to bind to %q: %v", "1", err)
	}

	var fd int

	if strings.HasPrefix(intfName, "tun") {
		fd, err = tun.Open(intfName)
		if err != nil {
			return fmt.Errorf("Could not open tun interface: %s", err.Error())
		}
	} else if strings.HasPrefix(intfName, "tap") {
		fd, err = tun.OpenTAP(intfName)
		if err != nil {
			return fmt.Errorf("Could not open tap interface: %s", err.Error())
		}
	} else {
		fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		if err != nil {
			return fmt.Errorf("Could not create socket: %s", err.Error())
		}

		if fd < 0 {
			return fmt.Errorf("Socket error: return < 0")
		}

		if err = syscall.SetNonblock(fd, true); err != nil {
			syscall.Close(fd)
			return fmt.Errorf("Error setting fd to nonblock: %s", err)
		}

		ll := syscall.SockaddrLinklayer{
			Protocol: htons(syscall.ETH_P_ALL),
			Ifindex:  ifaceLink.Attrs().Index,
			Hatype:   0, // No ARP type.
			Pkttype:  syscall.PACKET_HOST,
		}

		if err := syscall.Bind(fd, &ll); err != nil {
			return fmt.Errorf("unable to bind to %q: %v", "iface.Name", err)
		}
	}

	la := tcpip.LinkAddress(ifaceLink.Attrs().HardwareAddr)

	linkID, err := fdbased.New(&fdbased.Options{
		FDs:            []int{fd},
		MTU:            mtu,
		EthernetHeader: true,
		Address:        la,
		ClosedFunc: func(e *tcpip.Error) {
			if e != nil {
				log.Errorf("File descriptor closed: %v", err)
			}
		},
	})
	if err != nil {
		//TODO (jerry 2020-02-07): How to handle this error?
		log.Errorf("linkID: %v", err)
	}

	if l.Debug {
		linkID = sniffer.New(linkID)
	}

	linkID = NewFilter(linkID)

	routes := []tcpip.Route{}

	link := ifaceLink

	rs, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("error getting routes from %q: %v", link.Attrs().Name, err)
	}

	for _, r := range rs {
		// Is it a default route?
		if r.Dst == nil {
			if r.Gw == nil {
				return fmt.Errorf("default route with no gateway %q: %+v", link.Attrs().Name, r)
			}
			if r.Gw.To4() == nil {
				log.Warningf("IPv6 is not supported, skipping default route: %v", r)
				continue
			}

			subnet4, err := tcpip.NewSubnet(ipToAddress(net.IPv4zero), tcpip.AddressMask(net.IPv4zero))
			if err != nil {
				log.Warningf("Subnet: %v", err)
			}

			routes = append(routes, tcpip.Route{
				Destination: subnet4,
				Gateway:     ipToAddress(r.Gw),
			})
			continue
		}
		if r.Dst.IP.To4() == nil {
			log.Warningf("IPv6 is not supported, skipping route: %v", r)
			continue
		}
		subnet6, err := tcpip.NewSubnet(ipToAddress(r.Dst.IP.Mask(r.Dst.Mask)), tcpip.AddressMask(r.Dst.Mask))
		if err != nil {
			log.Warningf("Subnet: %v", err)
		}

		routes = append(routes, tcpip.Route{
			Destination: subnet6,
			Gateway:     ipToAddress(r.Gw),
		})
	}

	opts := stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{
			ipv4.NewProtocol(),
			ipv6.NewProtocol(),
		},
		TransportProtocols: []stack.TransportProtocol{
			tcp.NewProtocol(),
			udp.NewProtocol(),
		},
	}

	s := stack.New(opts)

	//TODO: can not find a suitable replacement for RSTDisabled.
	//see: pkg/tcpip/transport/tcp/protocol.go for Options.
	//
	//if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, tcp.RSTDisabled(true)); err != nil {
	//	return fmt.Errorf("Could not set transport protocol option: %s", err.String())
	//}

	s.AddTCPProbe(func(s stack.TCPEndpointState) {
	})

	s.SetRouteTable(routes)

	if err := s.CreateNIC(1, linkID); err != nil {
		return fmt.Errorf(err.String())
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("Error retrieving interface ip addresses: %s", err.Error())
	}

	if l.Addr != "" {
		if addr, err := netlink.ParseAddr(l.Addr); err == nil {
			addrs = []netlink.Addr{*addr}
		} else {
			return fmt.Errorf("Bad IP address: %v: %s", l.Addr, err)
		}
	}

	for _, parsedAddr := range addrs {
		var addr tcpip.Address
		var proto tcpip.NetworkProtocolNumber

		if _, bits := parsedAddr.Mask.Size(); bits == 32 {
			addr = tcpip.Address(parsedAddr.IP)
			proto = ipv4.ProtocolNumber
		} else if _, bits := parsedAddr.Mask.Size(); bits == 256 {
			addr = tcpip.Address(parsedAddr.IP)
			proto = ipv6.ProtocolNumber
		} else {
			return fmt.Errorf("Unknown IP type: %v, bits=%d", l.Addr, bits)
		}

		log.Debugf("Listening on: %s (%d)\n", parsedAddr.String(), proto)

		// s.AddSubnet()
		if err := s.AddAddress(1, proto, addr); err != nil {
			return fmt.Errorf(err.String())
		}
	}

	s.SetSpoofing(1, true)

	l.s = s

	for _, address := range l.Addresses {
		go func(address net.Addr) {
			if ta, ok := address.(*net.TCPAddr); ok {
				log.Infof("Listener started: tcp/%s", address)

				listener, err := gonet.NewListener(s, tcpip.FullAddress{
					NIC:  0,
					Addr: tcpip.Address(ta.IP),
					Port: uint16(ta.Port),
				}, ipv4.ProtocolNumber)
				if err != nil {
					log.Fatal(err)
				}
				defer listener.Close()

				for {
					conn, err := listener.Accept()
					if err != nil {
						log.Error("Error accepting tcp connection: %s", err.Error())
						continue
					}

					//TODO (jerry 2020-02-07): No suitable replacement for IRS()
					//if gc, ok := conn.(*gonet.Conn); !ok {
					//	//} else if irs, err := gc.IRS(); err != nil {
					//} else {
					//	conn = event.WithConn(conn, event.Custom("irs", irs))
					//}

					l.ch <- conn
				}
			} else if ua, ok := address.(*net.UDPAddr); ok {
				/*
					pc, err := gonet.NewPacketConn(s, tcpip.FullAddress{
						NIC:  0,
						Addr: tcpip.Address(ua.IP),
						Port: uint16(ua.Port),
					}, ipv4.ProtocolNumber)
				*/
				//laddr = nil = local address automatically chosen.
				pc, err := gonet.DialUDP(s, nil, &tcpip.FullAddress{
					NIC:  0,
					Addr: tcpip.Address(ua.IP),
					Port: uint16(ua.Port),
				}, ipv4.ProtocolNumber)
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
