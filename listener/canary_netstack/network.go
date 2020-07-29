package nscanary

import (
	"errors"
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

type RestoreDeviceFunc func() error

type setupStack struct {
	s         *stack.Stack
	ifaces    []*net.Interface
	args      *boot.CreateLinksAndRoutesArgs
	wrapper   LinkEndpointWrapper
	restoreFn RestoreDeviceFunc
}

func SetupNetworkStack(s *stack.Stack, iface []*net.Interface, wrap LinkEndpointWrapper) (*stack.Stack, RestoreDeviceFunc, error) {
	if len(iface) == 0 {
		return nil, nil, errors.New("no network interfaces to setup")
	}

	// Use defaults if no value given.
	if s == nil {
		s = newEmptyNetworkStack()
	}

	if wrap == nil {
		// set a default wrap func.
		wrap = func(ep stack.LinkEndpoint) stack.LinkEndpoint { return ep }
	}

	setup := setupStack{
		s:       s,
		ifaces:  iface,
		wrapper: wrap,
		args:    &boot.CreateLinksAndRoutesArgs{},
	}

	err := setup.CreateLinksAndRoutes()
	if err != nil {
		return nil, nil, err
	}

	if setup.restoreFn == nil {
		return nil, nil, errors.New("restore network device not set")
	}

	//setup.s.SetSpoofing(1, true)

	return setup.s, setup.restoreFn, nil
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

func (s *setupStack) createInterfacesAndRoutes() error {
	// Restore device addresses on listener exit.
	// map[netlink.Link] = []*net.IPNet.String()
	restoreMap := make(map[netlink.Link][]string)

	// Restore Routes on the host when honeytrap exits.
	restoreRoutes := make([]netlink.Route, 0, 10)

	defer func() {
		log.Debug("Setting RestoreDeviceFunc")
		restoreFn := func() error {
			for link, addrs := range restoreMap {
				for _, addr := range addrs {
					if err := restoreAddress(link, addr); err != nil {
						return err
					}
				}
			}
			for _, route := range restoreRoutes {
				if err := netlink.RouteAdd(&route); err != nil {
					log.Debugf("failed to add route: %v", err)
				}
			}
			return nil
		}
		s.restoreFn = RestoreDeviceFunc(restoreFn)
	}()

	for _, iface := range s.ifaces {
		if iface.Flags&net.FlagUp == 0 {
			log.Infof("Skipping down interface: %+v", iface)
			continue
		}

		// We skip loopback devices.
		if iface.Flags&net.FlagLoopback != 0 {
			log.Infof("Skipping loopback interface: %+v", iface)
			continue
		}

		allAddrs, err := iface.Addrs()
		if err != nil {
			return fmt.Errorf("fetching interface addresses for %q: %v", iface.Name, err)
		}
		if multi, err := iface.MulticastAddrs(); err != nil {
			log.Debugf("MulticastAddrs: %v", err)
		} else {
			log.Debugf("Multicast Addresses: %+v", multi)
		}

		var ipAddrs []*net.IPNet
		for _, ifaddr := range allAddrs {
			ipNet, ok := ifaddr.(*net.IPNet)
			if !ok {
				return fmt.Errorf("address is not IPNet: %+v", ifaddr)
			}
			ipAddrs = append(ipAddrs, ipNet)
		}
		if len(ipAddrs) == 0 {
			log.Warningf("No usable IP addresses found for interface %q, skipping", iface.Name)
			continue
		}

		// Scrape the routes before removing the address, since that
		// will remove the routes as well.
		routes, defv4, defv6, err := routesForIface(*iface)
		if err != nil {
			return fmt.Errorf("getting routes for interface %q: %v", iface.Name, err)
		}
		if defv4 != nil {
			if !s.args.Defaultv4Gateway.Route.Empty() {
				return fmt.Errorf("more than one default route found, interface: %v, route: %v, default route: %+v", iface.Name, defv4, s.args.Defaultv4Gateway)
			}
			s.args.Defaultv4Gateway.Route = *defv4
			s.args.Defaultv4Gateway.Name = iface.Name
		}

		if defv6 != nil {
			if !s.args.Defaultv6Gateway.Route.Empty() {
				return fmt.Errorf("more than one default route found, interface: %v, route: %v, default route: %+v", iface.Name, defv6, s.args.Defaultv6Gateway)
			}
			s.args.Defaultv6Gateway.Route = *defv6
			s.args.Defaultv6Gateway.Name = iface.Name
		}

		link := boot.FDBasedLink{
			Name:        iface.Name,
			MTU:         iface.MTU,
			Routes:      routes,
			NumChannels: 1,
		}

		// Get the link for the interface.
		ifaceLink, err := netlink.LinkByName(iface.Name)
		if err != nil {
			return fmt.Errorf("getting link for interface %q: %v", iface.Name, err)
		}
		link.LinkAddress = ifaceLink.Attrs().HardwareAddr

		log.Debugf("Setting up network channels")
		// Create the socket for the device.
		for i := 0; i < link.NumChannels; i++ {
			log.Debugf("Creating Channel %d", i)
			socketEntry, err := createSocket(iface, ifaceLink)
			if err != nil {
				return fmt.Errorf("failed to createSocket for %s : %v", iface.Name, err)
			}
			s.args.FilePayload.Files = append(s.args.FilePayload.Files, socketEntry.deviceFile)
		}

		// Collect all routes in the system to restore them later.
		// Do this before the addresses get removed.
		rs, err := netlink.RouteList(ifaceLink, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("failed to get routes: %v", err)
		}

		restoreRoutes = append(restoreRoutes, rs...)

		// Collect the addresses for the interface, enable forwarding,
		// and remove them from the host.
		for _, addr := range ipAddrs {
			link.Addresses = append(link.Addresses, addr.IP)

			// Save addresses to restore.
			restoreMap[ifaceLink] = append(restoreMap[ifaceLink], addr.String())

			// Steal IP address from NIC.
			if err := removeAddress(ifaceLink, addr.String()); err != nil {
				return fmt.Errorf("removing address %v from device %q: %v", iface.Name, addr, err)
			}
		}

		s.args.FDBasedLinks = append(s.args.FDBasedLinks, link)
	}

	log.Debugf("Setting up network, config: %+v", s.args)
	return nil
}

// CreateLinksAndRoutes creates links and routes in a network stack.  It should
// only be called once.
func (s *setupStack) CreateLinksAndRoutes() error {
	if err := s.createInterfacesAndRoutes(); err != nil {
		return err
	}

	wantFDs := 0
	for _, l := range s.args.FDBasedLinks {
		wantFDs += l.NumChannels
	}
	if got := len(s.args.FilePayload.Files); got != wantFDs {
		return fmt.Errorf("args.FilePayload.Files has %d FD's but we need %d entries based on FDBasedLinks", got, wantFDs)
	}

	var nicID tcpip.NICID
	nicids := make(map[string]tcpip.NICID)

	// Collect routes from all links.
	var routes []tcpip.Route

	// Loopback normally appear before other interfaces.
	// Don't do loopback NIC.

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
		// Use wrapper on endpoint.
		linkEP = packetsocket.New(s.wrapper(linkEP))

		if err := s.createNICWithAddrs(nicID, link.Name, linkEP, link.Addresses); err != nil {
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
func (s setupStack) createNICWithAddrs(id tcpip.NICID, name string, ep stack.LinkEndpoint, addrs []net.IP) error {
	opts := stack.NICOptions{Name: name}
	if err := s.s.CreateNICWithOptions(id, ep, opts); err != nil {
		return fmt.Errorf("CreateNICWithOptions(%d, _, %+v) failed: %v", id, opts, err)
	}

	// Always start with an arp address for the NIC.
	if err := s.s.AddAddress(id, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		return fmt.Errorf("AddAddress(%v, %v, %v) failed: %v", id, arp.ProtocolNumber, arp.ProtocolAddress, err)
	}

	for _, addr := range addrs {
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
	log.Debugf("Netlink Routes: %+v", rs)

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

// restoreAddress restores IP address on network device. It's equivalent to:
//   ip addr add <ipAndMask> dev <name>
func restoreAddress(source netlink.Link, ipAndMask string) error {
	addr, err := netlink.ParseAddr(ipAndMask)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(source, addr)
}
