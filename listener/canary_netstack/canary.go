// Package nscanary provides canary listener using gvisors netstack.
// https://github.com/google/gvisor/tree/master/pkg/tcpip
package nscanary

//
// config.toml
//  listener = "canary-netstack"
//  interfaces = ["iface"]
//
//  # exclude_log_protos sets the used protos for logging (optional) (default: all)
//  # recognized options for protos: ["ip4", "ip6", "arp", "udp", "tcp", "icmp"]
//  exclude_log_protos = [] (default)
//  # no_tls true: checks connection for tls and does tls handshake if so.
//  no_tls=false (default)
//

import (
	"context"
	"errors"
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
	//IfaceAddrs       []string `toml:"interface_addrs"`
	Interfaces       []string `toml:"interfaces"`
	ExcludeLogProtos []string `toml:"exclude_log_protos"`
	NoTLS            bool     `toml:"no_tls"`
	CertificateFile  string   `toml:"certificate_file"`
	KeyFile          string   `toml:"key_file"`
}

type Canary struct {
	Config

	listenAddrs []net.Addr
	interfaces  []net.Interface
	events      pushers.Channel
	nconn       chan net.Conn
	knockChan   chan KnockGrouper
	tlsConf     TLS

	stack            *stack.Stack
	restoreInterface RestoreDeviceFunc
}

//AddAddress implements listener.AddAddresser
func (c *Canary) AddAddress(a net.Addr) {
	c.listenAddrs = append(c.listenAddrs, a)
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	c := &Canary{
		events:    pushers.MustDummy(),
		nconn:     make(chan net.Conn),
		knockChan: make(chan KnockGrouper),
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.CertificateFile == "" || c.KeyFile == "" {
		c.NoTLS = true
	}

	if !c.NoTLS {
		c.tlsConf = NewTLSConf(c.CertificateFile, c.KeyFile)
	}

	// Set event creation flags.
	ExcludeLogProtos(c.ExcludeLogProtos)

	ifaces := []*net.Interface{}
	for _, name := range c.Interfaces {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			log.Errorf("can't use interface: %s, %v", name, err)
			continue
		}
		ifaces = append(ifaces, iface)
		log.Infof("using interface: %s", name)
	}

	// Create events on NIC traffic.
	eventWrapper := func(lower stack.LinkEndpoint) stack.LinkEndpoint {
		return WrapLinkEndpoint(lower, c.events, c.knockChan)
	}

	s, restoreFunc, err := SetupNetworkStack(nil, ifaces, eventWrapper)
	if err != nil {
		return nil, err
	}
	c.restoreInterface = restoreFunc

	fmt.Printf("s.NICInfo() = %+v\n", s.NICInfo())

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
	defer func() {
		log.Debug("restoring network device(s)...")
		err := c.restoreInterface()
		if err != nil {
			log.Debugf("restore network devices has error: %v", err)
		} else {
			log.Debug("restoring network device(s) successfull")
		}
	}()

	if !c.stack.CheckNIC(1) {
		return errors.New("check failed on NIC 1")
	}

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

			l, err := gonet.ListenTCP(c.stack, full, proto)
			if err != nil {
				log.Errorf("Error starting listener: %v", err)
				continue
			}

			log.Infof("Listener started: tcp/%s:%d", full.Addr.String(), full.Port)

			go func() {
				for {
					select {
					case <-ctx.Done():
						log.Debug("closing TCP listener")
						return
					default:
					}

					conn, err := l.Accept()
					if err != nil {
						log.Errorf("Error accepting connection: %s", err.Error())
						continue
					}

					if !c.NoTLS {
						mconn, err := c.tlsConf.MaybeTLS(conn, c.events)
						if err != nil {
							log.Errorf("failed maybe tls connection: %v", err)
							continue
						}
						conn = mconn
					}

					log.Debug("Accepted a connection")

					c.nconn <- conn

					log.Debug("connection is in channel")
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
