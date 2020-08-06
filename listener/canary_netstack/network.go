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
	"gvisor.dev/gvisor/pkg/tcpip/transport/raw"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type LinkEndpointWrapper func(stack.LinkEndpoint) stack.LinkEndpoint

type RestoreDeviceFunc func() error

type NIC struct {
	id     tcpip.NICID
	iface  netlink.Link
	addrs  []netlink.Addr
	routes []tcpip.Route

	defv4 *netlink.Route // default ipv4 route to restore
	defv6 *netlink.Route // default ipv6 route to restore.
}

func SetupNetworkStack(s *stack.Stack, interfaces []string, wrap LinkEndpointWrapper) (*stack.Stack, RestoreDeviceFunc, error) {

	// Use defaults if no stack is given.
	if s == nil {
		s = newEmptyNetworkStack()
	}

	if wrap == nil {
		// set a default do nothing wrap func.
		wrap = func(ep stack.LinkEndpoint) stack.LinkEndpoint { return ep }
	}

	nics := make([]NIC, 0, len(interfaces))

	var defroute4, defroute6 *tcpip.Route
	for nicID, name := range interfaces {

		iface, err := netlink.LinkByName(name)
		if err != nil {
			log.Errorf("can't use interface: %s, %v", name, err)
			continue
		}
		if iface.Attrs().Flags&net.FlagUp == 0 {
			log.Infof("Skipping down interface: %+s", name)
			continue
		}

		addrs, err := getAddresses(iface)
		if err != nil {
			log.Errorf("skipping %s, failed to get addresses: %v", iface.Attrs().Name)
			continue
		}

		nic := NIC{
			id:    tcpip.NICID(nicID + 1), // let NIC IDs start at 1.
			iface: iface,
			addrs: addrs,
		}

		defv4, defv6, err := getRoutes(&nic)
		if err != nil {
			return nil, nil, err
		}

		if defv4 != nil {
			if defroute4 == nil {
				defroute4 = defv4
			} else {
				return nil, nil, errors.New("found more than one default route")
			}
		}
		if defv6 != nil {
			if defroute6 == nil {
				defroute6 = defv6
			} else {
				return nil, nil, errors.New("found more than one default route")
			}
		}

		err = createNIC(&nic, s, wrap)
		if err != nil {
			return nil, nil, err
		}

		nics = append(nics, nic)

		log.Debugf("using interface: %s, type: %s", name, iface.Type())
	}

	if len(nics) == 0 {
		return nil, nil, errors.New("no network interfaces to setup")
	}

	if defroute4 != nil {
		s.AddRoute(*defroute4)
	}
	if defroute6 != nil {
		s.AddRoute(*defroute6)
	}

	restoreAddrs := make(map[string][]string)
	restoreDefRoutes := make([]*netlink.Route, 2)

	// Steal IP addresses from NICs.
	for _, nic := range nics {
		for _, addr := range nic.addrs {
			if err := netlink.AddrDel(nic.iface, &addr); err != nil {
				log.Errorf("removing address %v from device %q: %v", nic.iface.Attrs().Name, addr, err)
				continue
			}
			restoreAddrs[nic.iface.Attrs().Name] = append(restoreAddrs[nic.iface.Attrs().Name], addr.String())
			log.Debugf("[%s] deleted address: %s", addr)
		}
		if nic.defv4 != nil {
			restoreDefRoutes[0] = nic.defv4
		}
		if nic.defv6 != nil {
			restoreDefRoutes[1] = nic.defv6
		}
	}

	restoreFn := func() error {
		log.Debug("Setting RestoreDeviceFunc")
		for link, addrs := range restoreAddrs {
			for _, addr := range addrs {
				if err := restoreAddress(link, addr); err != nil {
					log.Errorf("[%s] failed to restore %s: %v", link, addr, err)
					continue
				}
				log.Debugf("[%s] restored address: %s", link, addr)
			}
			// restore default route if any.
			if restoreDefRoutes[0] != nil {
				if err := netlink.RouteAdd(restoreDefRoutes[0]); err != nil {
					log.Errorf("[%s] failed to restore default ipv4 route: %v", link, err)
				}
			}
			if restoreDefRoutes[1] != nil {
				if err := netlink.RouteAdd(restoreDefRoutes[1]); err != nil {
					log.Errorf("[%s] failed to restore default ipv6 route: %v", link, err)
				}
			}
		}
		return nil
	}

	for _, nic := range nics {
		s.SetSpoofing(nic.id, true)
	}

	for _, nic := range nics {
		if !s.CheckNIC(nic.id) {
			return nil, nil, fmt.Errorf("check NIC \"%s\" failed", nic.iface.Attrs().Name)
		}
	}

	return s, RestoreDeviceFunc(restoreFn), nil
}

func getRoutes(nic *NIC) (*tcpip.Route, *tcpip.Route, error) {
	rs, err := netlink.RouteList(nic.iface, netlink.FAMILY_ALL)
	if err != nil {
		return nil, nil, fmt.Errorf("getting routes from %q: %v", nic.iface.Attrs().Name, err)
	}

	for _, r := range rs {
		log.Debugf("[%s] route: %s", nic.iface.Attrs().Name, r.String())
	}

	var defv4, defv6 *tcpip.Route
	var routes []tcpip.Route
	for _, r := range rs {
		// Is it a default route?
		if r.Dst == nil {
			if r.Gw == nil {
				log.Errorf("default route with no gateway %q: %+v", nic.iface.Attrs().Name, r)
				continue
			}

			// Create a catch all route to the gateway.
			switch len(r.Gw) {
			case header.IPv4AddressSize:
				if defv4 != nil {
					return nil, nil, fmt.Errorf("more than one default route found %q, def: %+v, route: %+v", nic.iface.Attrs().Name, defv4, r)
				}
				defv4 = &tcpip.Route{
					Destination: header.IPv4EmptySubnet,
					Gateway:     tcpip.Address(r.Gw),
				}
				route4 := r
				nic.defv4 = &route4
			case header.IPv6AddressSize:
				if defv6 != nil {
					return nil, nil, fmt.Errorf("more than one default route found %q, def: %+v, route: %+v", nic.iface.Attrs().Name, defv6, r)
				}

				defv6 = &tcpip.Route{
					Destination: header.IPv6EmptySubnet,
					Gateway:     tcpip.Address(r.Gw),
				}
				route6 := r
				nic.defv6 = &route6
			default:
				return nil, nil, fmt.Errorf("unexpected address size for gateway: %+v for route: %+v", r.Gw, r)
			}
			continue
		}

		// sub, err := tcpip.NewSubnet(tcpip.Address(r.Dst.IP), tcpip.AddressMask(r.Dst.Mask))
		// if err != nil {
		// 	log.Errorf("NewSubnet: %v", err)
		// }

		// routes = append(routes, tcpip.Route{
		// 	Destination: sub,
		// 	Gateway:     tcpip.Address(r.Gw),
		// 	NIC:         nic.id,
		// })
	}

	nic.routes = routes

	return defv4, defv6, nil
}

func getAddresses(iface netlink.Link) ([]netlink.Addr, error) {
	addrs, err := netlink.AddrList(iface, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses for \"%s\": %v", iface.Attrs().Name, err)
	}
	return addrs, nil
}

// createNIC creates a NIC in the network stack and adds the NICs
// addresses and routes to it.
func createNIC(nic *NIC, s *stack.Stack, wrapFn LinkEndpointWrapper) error {
	// Create the raw socket.
	fd, err := createSocket(nic.iface)
	if err != nil {
		return err
	}

	// Copy the underlying FD.
	oldFD := fd.deviceFile.Fd()
	newFD, err := syscall.Dup(int(oldFD))
	if err != nil {
		return fmt.Errorf("failed to dup FD %v: %v", oldFD, err)
	}
	FDs := []int{newFD}

	mac := tcpip.LinkAddress(nic.iface.Attrs().HardwareAddr)

	linkEP, err := fdbased.New(&fdbased.Options{
		FDs:                FDs,
		MTU:                uint32(nic.iface.Attrs().MTU),
		EthernetHeader:     true,
		Address:            mac,
		PacketDispatchMode: fdbased.PacketMMap,
		GSOMaxSize:         0,
		SoftwareGSOEnabled: false,
		TXChecksumOffload:  false,
		RXChecksumOffload:  false,
	})
	if err != nil {
		return err
	}
	fmt.Printf("linkEP = %+v\n", linkEP)

	// Enable support for AF_PACKET sockets to receive outgoing packets.
	// Use wrapper on endpoint.
	linkEP = packetsocket.New(wrapFn(linkEP))

	opts := stack.NICOptions{Name: nic.iface.Attrs().Name}
	if err := s.CreateNICWithOptions(nic.id, linkEP, opts); err != nil {
		return fmt.Errorf("CreateNICWithOptions(%d, _, %+v) failed: %v", nic.id, opts, err)
	}

	// Always start with an arp address for the NIC.
	if err := s.AddAddress(nic.id, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		return fmt.Errorf("AddAddress(%v, %v, %v) failed: %v", nic.id, arp.ProtocolNumber, arp.ProtocolAddress, err)
	}

	for _, addr := range nic.addrs {
		proto, tcpipAddr := ipToAddressAndProto(addr.IP)
		if err := s.AddAddress(nic.id, proto, tcpipAddr); err != nil {
			log.Errorf("AddAddress(%v, %v, %v) failed: %v", nic.id, proto, tcpipAddr, err)
		}
	}
	log.Debug("NIC addresses added to the stack")

	// Add the NICs routes to the stack,
	for _, r := range nic.routes {
		s.AddRoute(r)
	}
	log.Debug("NICs routes added to the stack")

	return nil
}

type socketEntry struct {
	deviceFile *os.File
}

// createSocket creates an underlying AF_PACKET socket and configures it for use by
// the sentry and returns an *os.File that wraps the underlying socket fd.
func createSocket(iface netlink.Link) (*socketEntry, error) {
	const protocol = 0x0300 // htons(ETH_P_ALL)

	// Create the socket.
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, protocol)
	if err != nil {
		return nil, fmt.Errorf("unable to create raw socket: %v", err)
	}

	deviceFile := os.NewFile(uintptr(fd), "raw-device-fd")

	// Bind to the appropriate device.
	ll := syscall.SockaddrLinklayer{
		Protocol: protocol,
		Ifindex:  iface.Attrs().Index,
		Hatype:   0, // No ARP type.
		// Pkttype:  syscall.PACKET_OTHERHOST, // maybe need PACKET_HOST ???
	}
	if err := syscall.Bind(fd, &ll); err != nil {
		return nil, fmt.Errorf("unable to bind to %q: %v", iface.Attrs().Name, err)
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

func newEmptyNetworkStack() *stack.Stack {
	netProtos := []stack.NetworkProtocol{
		ipv4.NewProtocol(),
		ipv6.NewProtocol(),
		arp.NewProtocol(),
	}
	transProtos := []stack.TransportProtocol{
		tcp.NewProtocol(),
		udp.NewProtocol(),
		//icmp.NewProtocol4(),
		//icmp.NewProtocol6(),
	}
	s := stack.New(stack.Options{
		NetworkProtocols:   netProtos,
		TransportProtocols: transProtos,
		HandleLocal:        true,

		//TODO (jerry): Do we still need this??
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

// ipMaskToAddressMask converts IPMask to tcpip.AddressMask, ignoring the
// protocol.
func ipMaskToAddressMask(ipMask net.IPMask) tcpip.AddressMask {
	return tcpip.AddressMask(ipToAddress(net.IP(ipMask)))
}

// restoreAddress restores IP address on network device. It's equivalent to:
//   ip addr add <ipAndMask> dev <name>
func restoreAddress(iface string, ipAndMask string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(ipAndMask)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, addr)
}
