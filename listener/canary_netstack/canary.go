// Package nscanary provides canary listener using gvisors netstack.
// https://github.com/google/gvisor/tree/master/pkg/tcpip
package nscanary

//
// config.toml
//  listener="canary-netstack"
//  interfaces=["iface"]
//  addr=""
//

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/prometheus/common/log"
	"github.com/vishvananda/netlink"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/raw"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

type Canary struct {
	Addr       string   `toml:"addr"`
	Interfaces []string `toml:"interfaces"`

	events pushers.Channel
	nconn  chan net.Conn

	stack *stack.Stack
}

func New(options ...func(listener.Listener) error) (*Canary, error) {
	c := &Canary{
		events: pushers.MustDummy(),
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if len(c.Interfaces) == 0 {
		return nil, fmt.Errorf("no interface defined")
	} else if len(c.Interfaces) > 1 {
		return nil, fmt.Errorf("only one interface is supported currently")
	}

	iface := c.Interfaces[0]

	ifaceLink, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("unable to bind to %q: %v", "1", err)
	}

	var fd int

	if strings.HasPrefix(iface, "tun") {
		fd, err = tun.Open(iface)
		if err != nil {
			return nil, fmt.Errorf("could not open tun interface: %s", err.Error())
		}
	} else if strings.HasPrefix(iface, "tap") {
		fd, err = tun.OpenTAP(iface)
		if err != nil {
			return nil, fmt.Errorf("could not open tap interface: %s", err.Error())
		}
	} else {
		fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		if err != nil {
			return nil, fmt.Errorf("could not create socket: %s", err.Error())
		}

		if fd < 0 {
			return nil, fmt.Errorf("socket error: return < 0")
		}

		if err = syscall.SetNonblock(fd, true); err != nil {
			syscall.Close(fd)
			return nil, fmt.Errorf("error setting fd to nonblock: %s", err)
		}

		ll := syscall.SockaddrLinklayer{
			Protocol: htons(syscall.ETH_P_ALL),
			Ifindex:  ifaceLink.Attrs().Index,
			Hatype:   0, // No ARP type.
			Pkttype:  syscall.PACKET_HOST,
		}

		if err := syscall.Bind(fd, &ll); err != nil {
			return nil, fmt.Errorf("unable to bind to %q: %v", "iface.Name", err)
		}
	}

	// create a new stack
	opts := stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{
			ipv4.NewProtocol(),
			ipv6.NewProtocol(),
		},
		TransportProtocols: []stack.TransportProtocol{
			tcp.NewProtocol(),
			udp.NewProtocol(),
			icmp.NewProtocol4(),
			icmp.NewProtocol6(),
		},
		RawFactory: raw.EndpointFactory{},
	}

	s := stack.New(opts)

	// setup a link endpoint

	mtu, err := rawfile.GetMTU(iface)
	if err != nil {
		return nil, err
	}

	la := tcpip.LinkAddress(ifaceLink.Attrs().HardwareAddr)

	linkEP, err := fdbased.New(&fdbased.Options{
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
		return nil, fmt.Errorf("failed creating a link endpoint: %w", err)
	}

	//TODO (jerry): wrap the linkID with filter??

	s.CreateNIC(1, linkEP)

	// set the route table.

	link := ifaceLink

	routes, err := Routes(link)
	if err != nil {
		return nil, fmt.Errorf("get routes: %w", err)
	}
	s.SetRouteTable(routes)

	// stack.AddAddress()

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("error retrieving interface ip addresses: %s", err.Error())
	}

	if c.Addr != "" {
		addr, err := netlink.ParseAddr(c.Addr)
		if err != nil {
			return nil, fmt.Errorf("bad IP address: %v: %s", c.Addr, err)
		}
		addrs = []netlink.Addr{*addr}
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
			return nil, fmt.Errorf("unknown IP type: %v, bits=%d", c.Addr, bits)
		}

		log.Debugf("Listening on: %s (%d)\n", parsedAddr.String(), proto)

		//stack.AddAddressRange() from subnet??
		if err := s.AddAddress(1, proto, addr); err != nil {
			return nil, fmt.Errorf(err.String())
		}
	}

	s.SetSpoofing(1, true)

	c.stack = s

	return c, nil
}

func (c *Canary) Accept() (net.Conn, error) {
	conn := <-c.nconn
	return conn, nil
}

func (c *Canary) SetChannel(ch pushers.Channel) {
	c.events = ch
}

func (c *Canary) Start(ctx context.Context) error {
	var wq waiter.Queue

	endpoint, err := c.stack.NewPacketEndpoint(true, tcpip.NetworkProtocolNumber(htons(syscall.ETH_P_ALL)), &wq)
	if err != nil {
		return fmt.Errorf("create NewPacketEndpoint: %s", err.String())
	}
	defer endpoint.Close()

	// Wait for connections to appear.
	waitEntry, notifyCh := waiter.NewChannelEntry(nil)
	wq.EventRegister(&waitEntry, waiter.EventIn)
	defer wq.EventUnregister(&waitEntry)

	for {
		n, wq, err := endpoint.Read()
		if err != nil {
			if err == tcpip.ErrWouldBlock {
				<-notifyCh
				continue
			}
			return fmt.Errorf("read error: %s", err.String())
		}
	}
}

func Routes(link netlink.Link) ([]tcpip.Route, error) {
	rs, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("error getting routes from %q: %v", link.Attrs().Name, err)
	}

	var (
		subnet tcpip.Subnet
		routes = make([]tcpip.Route, 0, len(rs))
	)

	for _, route := range rs {
		if route.Dst == nil && route.Gw != nil { //default route.
			if route.Gw.To4() == nil {
				subnet, err = tcpip.NewSubnet(ipToAddress(net.IPv6zero), tcpip.AddressMask(net.IPv6zero))
			} else {
				subnet, err = tcpip.NewSubnet(ipToAddress(net.IPv4zero), tcpip.AddressMask(net.IPv4zero))
			}
		} else {
			subnet, err = tcpip.NewSubnet(ipToAddress(route.Dst.IP.Mask(route.Dst.Mask)), tcpip.AddressMask(route.Dst.Mask))
		}
		if err != nil {
			return nil, err
		}
		routes = append(routes, tcpip.Route{
			Destination: subnet,
			NIC:         1,
		})
	}
	return routes, nil
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
