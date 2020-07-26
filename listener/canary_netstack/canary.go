// Package nscanary provides canary listener using gvisors netstack.
// https://github.com/google/gvisor/tree/master/pkg/tcpip
package nscanary

//
// config.toml
//  listener = "canary-netstack"
//  interfaces = ["iface"]
//
//  # interface addresses to use, ipv4/ipv6.
//  interface-addrs=["1.2.3.4", "ff80::1"]
//
//  # exclude_log_protos sets the used protos for logging (optional) (default: all)
//  # recognized options for protos: ["ip4", "ip6", "arp", "udp", "tcp", "icmp"]
//  exclude_log_protos = []
//

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var log = logging.MustGetLogger("listeners/netstack-canary")

var (
	_                    = listener.Register("netstack-canary", New)
	EventCategoryUnknown = event.Category("unknown")
	SensorCanary         = event.Sensor("canary")

	CanaryOptions = event.NewWith(
		SensorCanary,
	)

	//IPv6loopback               = net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	//IPv6interfacelocalallnodes = net.IP{0xff, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	//IPv6linklocalallnodes      = net.IP{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	//IPv6linklocalallrouters    = net.IP{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02}
)

type Config struct {
	IfaceAddrs       []string `toml:"interface_addrs"`
	Interfaces       []string `toml:"interfaces"`
	ExcludeLogProtos []string `toml:"exclude_log_protos"`
}

type Canary struct {
	Config

	listenAddrs []net.Addr
	interfaces  []net.Interface
	events      pushers.Channel
	nconn       chan net.Conn
	knockChan   chan KnockGrouper
	tlsConf     TLS

	stack *stack.Stack
}

//AddAddress implements listener.AddAddresser
func (c *Canary) AddAddress(a net.Addr) {
	c.listenAddrs = append(c.listenAddrs, a)
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	c := &Canary{
		events:    pushers.MustDummy(),
		knockChan: make(chan KnockGrouper),
		tlsConf:   NewTLSConf("", ""),
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// set log flags.
	ExcludeLogProtos(c.ExcludeLogProtos)

	if len(c.Interfaces) == 0 {
		return nil, fmt.Errorf("no interface defined")
	}

	iface, err := net.InterfaceByName(c.Interfaces[0])
	if err != nil {
		return nil, err
	}

	if iface.Flags&net.FlagUp == 0 {
		return nil, fmt.Errorf("interface is down: %+v", iface)
	}

	var addrs []net.IP
	for _, addr := range c.IfaceAddrs {
		if ip := net.ParseIP(addr); ip != nil {
			addrs = append(addrs, ip)
		}
	}

	eventWrapper := func(lower stack.LinkEndpoint) stack.LinkEndpoint {
		return WrapLinkEndpoint(lower, c.events, c.knockChan)
	}

	s, err := SetupNetworkStack(nil, iface, addrs, eventWrapper)
	if err != nil {
		return nil, err
	}

	fmt.Printf("s.AllAdresses() = %+v\n", s.AllAddresses())
	if main, err := s.GetMainNICAddress(1, ipv4.ProtocolNumber); err != nil {
		fmt.Printf("err = %+v\n", err)
	} else {
		fmt.Printf("main = %+v\n", main)
	}
	if main, err := s.GetMainNICAddress(1, ipv6.ProtocolNumber); err != nil {
		fmt.Printf("err = %+v\n", err)
	} else {
		fmt.Printf("main = %+v\n", main)
	}
	fmt.Printf("routes:  = %+v\n", s.GetRouteTable())

	c.stack = s

	log.Infof("canary started using network interface: %+v", iface)

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

	go RunKnockDetector(ctx, c.knockChan, c.events)

	fmt.Printf("c.listenAddrs = %+v\n", c.listenAddrs)

	for _, address := range c.listenAddrs {
		full := tcpip.FullAddress{
			NIC: 1,
		}

		if a, ok := address.(*net.TCPAddr); ok {
			proto, addr := ipToAddressAndProto(a.IP)

			full.Addr = addr
			full.Port = uint16(a.Port)

			//l, err := ListenTCP(c.stack, full, netproto)
			l, err := gonet.ListenTCP(c.stack, full, proto)
			if err != nil {
				log.Errorf("Error starting listener: %v", err)
				continue
			}

			log.Infof("Listener started: tcp/%s:%d", full.Addr.String(), full.Port)

			isTLS := false

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}

					//conn, isTLS, err := l.Accept()
					conn, err := l.Accept()
					if err != nil {
						log.Errorf("Error accepting connection: %s", err.Error())
						continue
					}

					if isTLS {
						tlsConn, err := c.tlsConf.Handshake(conn, c.events)
						if err != nil {
							log.Debugf("tls connection: %v", err)
							continue
						}
						c.nconn <- tlsConn
						continue
					}

					c.nconn <- conn
				}
			}()
		} else if _, ok := address.(*net.UDPAddr); ok {
			proto, addr := ipToAddressAndProto(a.IP)

			full.Addr = addr
			full.Port = uint16(a.Port)

			l, err := gonet.DialUDP(c.stack, &full, nil, proto)
			if err != nil {
				log.Errorf("Error starting listener: %v", err)
				continue
			}

			ul := UDPConn{l}

			log.Infof("Listener started: udp/%s", address)

			go func() {
				select {
				case <-ctx.Done():
					return
				default:
				}

				for {
					var buf [65535]byte

					n, raddr, err := ul.ReadFromUDP(buf[:])
					if err != nil {
						log.Error("Error reading udp:", err.Error())
						continue
					}

					c.nconn <- &listener.DummyUDPConn{
						Buffer: buf[:n],
						Laddr:  l.LocalAddr(),
						Raddr:  raddr,
						Fn:     ul.WriteToUDP,
					}
				}
			}()
		}
	}

	return nil
}

func htons(n uint16) uint16 {
	var (
		high = n >> 8
		ret  = n<<8 + high
	)
	return ret
}

//fileDescriptor opens a raw socket and binds it to network interface with name 'link'
//returns the socket file descriptor, on error fd=0.
func fileDescriptor(link string, linkIndex int) (int, error) {

	var fd int
	var err error

	if strings.HasPrefix(link, "tun") {
		fd, err = tun.Open(link)
		if err != nil {
			return 0, fmt.Errorf("could not open tun interface: %s", err.Error())
		}
	} else if strings.HasPrefix(link, "tap") {
		fd, err = tun.OpenTAP(link)
		if err != nil {
			return 0, fmt.Errorf("could not open tap interface: %s", err.Error())
		}
	} else {
		//fd, err = unix.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		fd, err = unix.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0)
		if err != nil {
			return 0, fmt.Errorf("could not create socket: %s", err.Error())
		}

		if fd < 0 {
			return 0, fmt.Errorf("socket error: fd < 0")
		}
		if err := unix.SetNonblock(fd, true); err != nil {
			return 0, err
		}

		ll := unix.SockaddrLinklayer{
			Protocol: htons(syscall.ETH_P_ALL),
			Ifindex:  linkIndex,
		}

		if err := unix.Bind(fd, &ll); err != nil {
			return 0, fmt.Errorf("unable to bind to %q: %v", link, err)
		}
	}
	return fd, nil
}
