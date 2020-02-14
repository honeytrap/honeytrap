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
package xnetstack

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/vishvananda/netlink"

	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/listener/netstack-experimental/arp"
	udpf "github.com/honeytrap/honeytrap/listener/netstack-experimental/udp"
	"github.com/honeytrap/honeytrap/pushers"

	"math/big"

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
	"gvisor.dev/gvisor/pkg/waiter"
)

// todo
// port /whitelist filtering (8022)
// custom (irs )
// arp
// detect half open syn scans
// check listening addresses / port (ip:port in service config)

type netstackConfig struct {
	Addresses []net.Addr

	Addr       []string `toml:"addr"`
	Interfaces []string `toml:"interfaces"`

	// use https://github.com/goccmack/gocc for filter?
	// Filter expression: !1.2.3.4 and !22
	// https://blog.gopheracademy.com/advent-2014/parsers-lexers/
	Filter []string `toml:"filter"`

	Debug bool `toml:"debug"`
}

// should be AddPort?
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

	// linkID = NewFilter(linkID)

	nicID := tcpip.NICID(1)

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
		/*
				routes = append(routes, tcpip.Route{
					Destination: ipToAddress(net.IPv4zero),
					Mask:        tcpip.AddressMask(net.IPv4zero),
					Gateway:     ipToAddress(r.Gw),
				})
				continue
			}
			if r.Dst.IP.To4() == nil {
				log.Warningf("IPv6 is not supported, skipping route: %v", r)
				continue
			}
			routes = append(routes, tcpip.Route{
				Destination: ipToAddress(r.Dst.IP.Mask(r.Dst.Mask)),
				Mask:        tcpip.AddressMask(r.Dst.Mask),
				NIC:         nicID,
			})
		*/
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

	filterAddrs := []tcpip.Subnet{}
	for _, s := range l.Filter {
		_, net, err := net.ParseCIDR(s)
		if err != nil {
			log.Fatalf("Could not parse filter address: %s: %s", s, err.Error())
		}

		mask := tcpip.AddressMask([]byte{255, 255, 255, 255})
		if net != nil {
			mask = tcpip.AddressMask(net.Mask)
		}

		subnet, err := tcpip.NewSubnet(tcpip.Address(net.IP).To4(), mask)
		if err != nil {
			log.Fatalf("Could not create subnet: %s: %s", s, err.Error())
		}

		filterAddrs = append(filterAddrs, subnet)
	}

	canHandle := func(addr2 net.Addr) bool {
		for _, addr := range l.Addresses {
			if ta, ok := addr.(*net.TCPAddr); ok {
				ta2, ok := addr2.(*net.TCPAddr)
				if !ok {
					// compare apples with pears
					continue
				}

				if ta.Port != ta2.Port {
					continue
				}

				if ta.IP == nil {
				} else if !ta.IP.Equal(ta2.IP) {
					continue
				}

				return true
			} else if ua, ok := addr.(*net.UDPAddr); ok {
				ua2, ok := addr2.(*net.UDPAddr)
				if !ok {
					// compare apples with pears
					continue
				}
				if ua.Port != ua2.Port {
					continue
				}

				if ua.IP == nil {
				} else if !ua.IP.Equal(ua2.IP) {
					continue
				}

				return true
			}
		}

		return false
	}

	shouldFilter := func(id stack.TransportEndpointID) bool {
		for _, addr := range filterAddrs {
			// don't respond to filtered addresses
			if addr.Contains(id.RemoteAddress) {
				return true
			}
		}

		return false
	}

	tcpForwarder := tcp.NewForwarder(s, 30000, 5000, func(r *tcp.ForwarderRequest) {
		// got syn
		// check for ports to ignore
		id := r.ID()

		if !canHandle(&net.TCPAddr{
			IP:   net.IP(id.LocalAddress),
			Port: int(id.LocalPort),
		}) {
			// catch all?
			// not listening to
			r.Complete(false)
			return
		}

		if shouldFilter(id) {
			// not listening to
			r.Complete(false)
			return
		}

		// should check here if port is being supported
		// do we have a port mapping for this service?

		// l.ch <- Accepter()
		// accepter.Accept() -> will run CreateEndpoint

		// perform handshake
		var wq waiter.Queue

		ep, err := r.CreateEndpoint(&wq)
		if err != nil {
			// handshake failed, cleanup
			r.Complete(false)

			log.Errorf("Error creating endpoint: %s", err)
			return
		}

		r.Complete(false)

		c := gonet.NewConn(&wq, ep)
		l.ch <- c
	})

	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

	udpForwarder := udpf.NewForwarder(s, func(fr *udpf.ForwarderRequest) {
		id := fr.ID()

		if !canHandle(
			&net.UDPAddr{
				IP:   net.IP(id.LocalAddress),
				Port: int(id.LocalPort),
			}) {
			return
		}

		if shouldFilter(id) {
			// not listening to
			return
		}

		l.ch <- &listener.DummyUDPConn{
			Buffer: fr.Payload(),
			Laddr: &net.UDPAddr{
				IP:   net.IP(id.LocalAddress),
				Port: int(id.LocalPort),
			},
			Raddr: &net.UDPAddr{
				IP:   net.IP(id.RemoteAddress),
				Port: int(id.RemotePort),
			},
			Fn: fr.Write,
		}

	})

	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)

	if err := s.CreateNIC(nicID, linkID); err != nil {
		return fmt.Errorf(err.String())
	}

	// use address list
	ips := []net.IP{}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("Error retrieving interface ip addresses: %s", err.Error())
	}

	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}

	if len(l.Addr) != 0 {
		// l.Addr will override network configuration
		ips = []net.IP{}

		for _, s := range l.Addr {
			parts := strings.Split(s, "-")

			ip4ToNumber := func(ip net.IP) *big.Int {
				b := &big.Int{}
				b.SetBytes(ip.To4())
				return b
			}

			numberToIP4 := func(b *big.Int) net.IP {
				return net.IP(b.Bytes())
			}

			var ip1 *big.Int
			if ip := net.ParseIP(parts[0]); ip == nil {
				log.Errorf("Bad IP address: %v", s, ip)
				continue
			} else {
				ip1 = ip4ToNumber(ip)
			}

			ip2 := &big.Int{}
			ip2.Set(ip1)

			if len(parts) == 1 {
			} else if ip := net.ParseIP(parts[1]); ip == nil {
				log.Errorf("Bad IP address: %v", s)
				continue
			} else {
				ip2 = ip4ToNumber(ip)
			}

			ip := &big.Int{}
			ip.Set(ip1)

			for {
				log.Debug(numberToIP4(ip).String())

				ips = append(ips, numberToIP4(ip))

				ip.Add(ip, big.NewInt(1))
				if ip.Cmp(ip2) >= 0 {
					break
				}
			}
		}
	}

	for _, ip := range ips {
		var addr tcpip.Address
		var proto tcpip.NetworkProtocolNumber

		addr = tcpip.Address(ip).To4()
		proto = ipv4.ProtocolNumber

		/*
			if _, bits := parsedAddr.Mask.Size(); bits == 32 {
				addr = tcpip.Address(ip).To4()
				proto = ipv4.ProtocolNumber
			} else if _, bits := parsedAddr.Mask.Size(); bits == 256 {
				addr = tcpip.Address(parsedAddr.IP)
				proto = ipv6.ProtocolNumber
			} else {
				return fmt.Errorf("Unknown IP type: %v, bits=%d", l.Addr, bits)
			}
		*/

		if err := s.AddAddress(nicID, proto, addr); err != nil {
			return fmt.Errorf("Error listening on: %s: %s", addr.String(), err.String())
		}

		log.Debugf("Listening on: %s (%d)", ip.String(), proto)
	}

	_ = addrs

	/*
		if err := s.AddAddress(nicID, ipv4.ProtocolNumber, tcpip.Address(net.ParseIP("145.220.137.242")).To4()); err != nil {
			log.Fatalf("Could not register custom ip: %s", err.String())
		}
	*/

	// add l.addresses too
	// by importing our own arp handler, we're overriding default netstack behaviour
	if err := s.AddAddress(nicID, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		log.Fatalf("Could not register arp handler: %s", err.String())
	}

	s.SetSpoofing(1, true)

	l.s = s

	return nil
}

func (l *netstackListener) Accept() (net.Conn, error) {
	c := <-l.ch
	return c, nil
}
