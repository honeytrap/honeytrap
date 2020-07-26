package nscanary

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/vishvananda/netlink"
	"gvisor.dev/gvisor/pkg/sentry/socket/netstack"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/packetsocket"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/raw"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/runsc/boot"
)

type LinkEndpointWrapper func(stack.LinkEndpoint) stack.LinkEndpoint

type setupStack struct {
	s       *stack.Stack
	iface   *net.Interface
	addrs   []net.IP
	args    *boot.CreateLinksAndRoutesArgs
	wrapper LinkEndpointWrapper
}

func SetupNetworkStack(s *stack.Stack, iface *net.Interface, addrs []net.IP, wrap LinkEndpointWrapper) (*stack.Stack, error) {
	setup := setupStack{
		s:       s,
		iface:   iface,
		addrs:   addrs,
		wrapper: wrap,
		args:    &boot.CreateLinksAndRoutesArgs{},
	}

	if setup.s == nil {
		setup.s = newEmptyNetworkStack()
	}

	if setup.wrapper == nil {
		// set a default wrap func.
		setup.wrapper = func(ep stack.LinkEndpoint) stack.LinkEndpoint { return ep }
	}

	err := setup.CreateLinksAndRoutes()
	if err != nil {
		return nil, err
	}

	//setup.s.SetSpoofing(1, true)

	return setup.s, nil
}

func newEmptyNetworkStack() *stack.Stack {
	netProtos := []stack.NetworkProtocol{ipv4.NewProtocol(), ipv6.NewProtocol(), arp.NewProtocol()}
	transProtos := []stack.TransportProtocol{tcp.NewProtocol(), udp.NewProtocol(), icmp.NewProtocol4(), icmp.NewProtocol6()}
	s := stack.New(stack.Options{
		NetworkProtocols:   netProtos,
		TransportProtocols: transProtos,
		HandleLocal:        true,
		// Enable raw sockets for users with sufficient
		// privileges.
		RawFactory: raw.EndpointFactory{},
	})

	// Enable SACK Recovery.
	if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, tcp.SACKEnabled(true)); err != nil {
		log.Errorf("failed to enable SACK: %s", err)
	}

	// Set default TTLs as required by socket/netstack.
	s.SetNetworkProtocolOption(ipv4.ProtocolNumber, tcpip.DefaultTTLOption(netstack.DefaultTTL))
	s.SetNetworkProtocolOption(ipv6.ProtocolNumber, tcpip.DefaultTTLOption(netstack.DefaultTTL))

	// Enable Receive Buffer Auto-Tuning.
	if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, tcpip.ModerateReceiveBufferOption(true)); err != nil {
		log.Errorf("SetTransportProtocolOption failed: %s", err)
	}

	return s
}

func (s setupStack) createInterfacesAndRoutes() error {

	// Collect addresses and routes from the interfaces.

	// Scrape the routes.
	routes, defv4, defv6, err := routesForIface(*s.iface)
	if err != nil {
		return fmt.Errorf("getting routes for interface %q: %v", s.iface.Name, err)
	}
	if defv4 != nil {
		s.args.Defaultv4Gateway.Route = *defv4
		s.args.Defaultv4Gateway.Name = s.iface.Name
	}

	if defv6 != nil {
		s.args.Defaultv6Gateway.Route = *defv6
		s.args.Defaultv6Gateway.Name = s.iface.Name
	}

	link := boot.FDBasedLink{
		Name:        s.iface.Name,
		MTU:         s.iface.MTU,
		Routes:      routes,
		NumChannels: 1,
	}

	// Get the link for the interface.
	ifaceLink, err := netlink.LinkByName(s.iface.Name)
	if err != nil {
		return fmt.Errorf("getting link for interface %q: %v", s.iface.Name, err)
	}
	link.LinkAddress = ifaceLink.Attrs().HardwareAddr

	// Create the socket for the device.
	socketEntry, err := createSocket(s.iface, ifaceLink)
	if err != nil {
		return fmt.Errorf("failed to createSocket for %s : %v", s.iface.Name, err)
	}
	s.args.FilePayload.Files = append(s.args.FilePayload.Files, socketEntry.deviceFile)

	// Collect the addresses for the interface, enable forwarding,
	// and remove them from the host.
	link.Addresses = append(link.Addresses, s.addrs...)

	// Steal IP address from NIC.
	//if err := removeAddress(ifaceLink, addr.String()); err != nil {
	//	return fmt.Errorf("removing address %v from device %q: %v", iface.Name, addr, err)
	//}

	s.args.FDBasedLinks = append(s.args.FDBasedLinks, link)

	return nil
}

// CreateLinksAndRoutes creates links and routes in a network stack.  It should
// only be called once.
func (s setupStack) CreateLinksAndRoutes() error {
	if err := s.createInterfacesAndRoutes(); err != nil {
		return err
	}

	var nicID tcpip.NICID
	nicids := make(map[string]tcpip.NICID)

	// Collect routes from all links.
	var routes []tcpip.Route

	fdOffset := 0
	for _, link := range s.args.FDBasedLinks {
		nicID++
		nicids[link.Name] = nicID

		FDs := []int{}
		for j := 0; j < link.NumChannels; j++ {
			// Copy the underlying FD.
			oldFD := s.args.FilePayload.Files[fdOffset].Fd()
			newFD, err := syscall.Dup(int(oldFD))
			if err != nil {
				return fmt.Errorf("failed to dup FD %v: %v", oldFD, err)
			}
			FDs = append(FDs, newFD)
			fdOffset++
		}

		mac := tcpip.LinkAddress(link.LinkAddress)

		linkEP, err := fdbased.New(&fdbased.Options{
			FDs:                FDs,
			MTU:                uint32(link.MTU),
			EthernetHeader:     true,
			Address:            mac,
			PacketDispatchMode: fdbased.RecvMMsg,
			GSOMaxSize:         link.GSOMaxSize,
			SoftwareGSOEnabled: link.SoftwareGSOEnabled,
			TXChecksumOffload:  link.TXChecksumOffload,
			RXChecksumOffload:  link.RXChecksumOffload,
		})
		if err != nil {
			return err
		}

		log.Infof("Enabling interface %q with id %d on addresses %+v (%v) w/ %d channels", link.Name, nicID, link.Addresses, mac, link.NumChannels)

		// Enable support for AF_PACKET sockets to receive outgoing packets.
		linkEP = packetsocket.New(s.wrapper(linkEP))

		if err := s.createNICWithAddrs(nicID, link.Name, linkEP); err != nil {
			return err
		}

		// Collect the routes from this link.
		for _, r := range link.Routes {
			route, err := toTcpipRoute(r, nicID)
			if err != nil {
				return err
			}
			routes = append(routes, route)
		}
	}

	if !s.args.Defaultv4Gateway.Route.Empty() {
		nicID, ok := nicids[s.args.Defaultv4Gateway.Name]
		if !ok {
			return fmt.Errorf("invalid interface name %q for default route", s.args.Defaultv4Gateway.Name)
		}
		route, err := toTcpipRoute(s.args.Defaultv4Gateway.Route, nicID)
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}

	if !s.args.Defaultv6Gateway.Route.Empty() {
		nicID, ok := nicids[s.args.Defaultv6Gateway.Name]
		if !ok {
			return fmt.Errorf("invalid interface name %q for default route", s.args.Defaultv6Gateway.Name)
		}
		route, err := toTcpipRoute(s.args.Defaultv6Gateway.Route, nicID)
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}

	log.Infof("Setting routes %+v", routes)
	s.s.SetRouteTable(routes)
	return nil
}

// createNICWithAddrs creates a NIC in the network stack and adds the given
// addresses.
func (s setupStack) createNICWithAddrs(id tcpip.NICID, name string, ep stack.LinkEndpoint) error {
	opts := stack.NICOptions{Name: name}
	if err := s.s.CreateNICWithOptions(id, ep, opts); err != nil {
		return fmt.Errorf("CreateNICWithOptions(%d, _, %+v) failed: %v", id, opts, err)
	}

	// Always start with an arp address for the NIC.
	if err := s.s.AddAddress(id, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		return fmt.Errorf("AddAddress(%v, %v, %v) failed: %v", id, arp.ProtocolNumber, arp.ProtocolAddress, err)
	}

	for _, addr := range s.addrs {
		proto, tcpipAddr := ipToAddressAndProto(addr)
		if err := s.s.AddAddress(id, proto, tcpipAddr); err != nil {
			return fmt.Errorf("AddAddress(%v, %v, %v) failed: %v", id, proto, tcpipAddr, err)
		}
	}
	return nil
}

// routesForIface iterates over all routes for the given interface and converts
// them to boot.Routes. It also returns the a default v4/v6 route if found.
func routesForIface(iface net.Interface) ([]boot.Route, *boot.Route, *boot.Route, error) {
	link, err := netlink.LinkByIndex(iface.Index)
	if err != nil {
		return nil, nil, nil, err
	}
	rs, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting routes from %q: %v", iface.Name, err)
	}

	var defv4, defv6 *boot.Route
	var routes []boot.Route
	for _, r := range rs {
		// Is it a default route?
		if r.Dst == nil {
			if r.Gw == nil {
				return nil, nil, nil, fmt.Errorf("default route with no gateway %q: %+v", iface.Name, r)
			}
			// Create a catch all route to the gateway.
			switch len(r.Gw) {
			case header.IPv4AddressSize:
				if defv4 != nil {
					return nil, nil, nil, fmt.Errorf("more than one default route found %q, def: %+v, route: %+v", iface.Name, defv4, r)
				}
				defv4 = &boot.Route{
					Destination: net.IPNet{
						IP:   net.IPv4zero,
						Mask: net.IPMask(net.IPv4zero),
					},
					Gateway: r.Gw,
				}
			case header.IPv6AddressSize:
				if defv6 != nil {
					return nil, nil, nil, fmt.Errorf("more than one default route found %q, def: %+v, route: %+v", iface.Name, defv6, r)
				}

				defv6 = &boot.Route{
					Destination: net.IPNet{
						IP:   net.IPv6zero,
						Mask: net.IPMask(net.IPv6zero),
					},
					Gateway: r.Gw,
				}
			default:
				return nil, nil, nil, fmt.Errorf("unexpected address size for gateway: %+v for route: %+v", r.Gw, r)
			}
			continue
		}

		dst := *r.Dst
		dst.IP = dst.IP.Mask(dst.Mask)
		routes = append(routes, boot.Route{
			Destination: dst,
			Gateway:     r.Gw,
		})
	}
	return routes, defv4, defv6, nil
}

type socketEntry struct {
	deviceFile *os.File
}

// createSocket creates an underlying AF_PACKET socket and configures it for use by
// the sentry and returns an *os.File that wraps the underlying socket fd.
func createSocket(iface *net.Interface, ifaceLink netlink.Link) (*socketEntry, error) {
	// Create the socket.
	const protocol = 0x0300 // htons(ETH_P_ALL)
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, protocol)
	if err != nil {
		return nil, fmt.Errorf("unable to create raw socket: %v", err)
	}

	deviceFile := os.NewFile(uintptr(fd), "raw-device-fd")

	// Bind to the appropriate device.
	ll := syscall.SockaddrLinklayer{
		Protocol: protocol,
		Ifindex:  iface.Index,
		Hatype:   0, // No ARP type.
		//Pkttype:  syscall.PACKET_OTHERHOST,
	}
	if err := syscall.Bind(fd, &ll); err != nil {
		return nil, fmt.Errorf("unable to bind to %q: %v", iface.Name, err)
	}

	// Use SO_RCVBUFFORCE/SO_SNDBUFFORCE because on linux the receive/send buffer
	// for an AF_PACKET socket is capped by "net.core.rmem_max/wmem_max".
	// wmem_max/rmem_max default to a unusually low value of 208KB. This is too low
	// for gVisor to be able to receive packets at high throughputs without
	// incurring packet drops.
	const bufSize = 4 << 20 // 4MB.

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUFFORCE, bufSize); err != nil {
		return nil, fmt.Errorf("failed to increase socket rcv buffer to %d: %v", bufSize, err)
	}

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUFFORCE, bufSize); err != nil {
		return nil, fmt.Errorf("failed to increase socket snd buffer to %d: %v", bufSize, err)
	}

	return &socketEntry{deviceFile}, nil
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

func toTcpipRoute(r boot.Route, id tcpip.NICID) (tcpip.Route, error) {
	subnet, err := tcpip.NewSubnet(ipToAddress(r.Destination.IP), ipMaskToAddressMask(r.Destination.Mask))
	if err != nil {
		return tcpip.Route{}, err
	}
	return tcpip.Route{
		Destination: subnet,
		Gateway:     ipToAddress(r.Gateway),
		NIC:         id,
	}, nil
}

// ipToAddress converts IP to tcpip.Address, ignoring the protocol.
func ipToAddress(ip net.IP) tcpip.Address {
	_, addr := ipToAddressAndProto(ip)
	return addr
}

// ipMaskToAddressMask converts IPMask to tcpip.AddressMask, ignoring the
// protocol.
func ipMaskToAddressMask(ipMask net.IPMask) tcpip.AddressMask {
	return tcpip.AddressMask(ipToAddress(net.IP(ipMask)))
}

// removeAddress removes IP address from network device. It's equivalent to:
//   ip addr del <ipAndMask> dev <name>
func removeAddress(source netlink.Link, ipAndMask string) error {
	addr, err := netlink.ParseAddr(ipAndMask)
	if err != nil {
		return err
	}
	return netlink.AddrDel(source, addr)
}
