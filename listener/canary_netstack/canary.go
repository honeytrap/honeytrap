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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
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
	BlockPorts         []string `toml:"block_ports"`
	BlockSourceIP      []string `toml:"block_source_ip"`
	BlockDestinationIP []string `toml:"block_destination_ip"`

	Addrs           []string `toml:"addresses"`
	Interfaces      []string `toml:"interfaces"` // name of network interface (ip link)
	CertificateFile string   `toml:"certificate_file"`
	KeyFile         string   `toml:"key_file"`
}

type Canary struct {
	Config

	serviceAddrs []net.Addr
	events       pushers.Channel
	nconn        chan net.Conn
	knockChan    chan KnockGrouper

	stack          *stack.Stack
	restoreDevices RestoreDeviceFunc

	tls TLS
}

//AddAddress implements listener.AddAddresser
func (c *Canary) AddAddress(a net.Addr) {
	c.serviceAddrs = append(c.serviceAddrs, a)
}

func (c *Canary) AddTLSConfig(port uint16, config *tls.Config) {
	c.tls.AddConfig(port, config)
}

func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	c := &Canary{
		events:    pushers.MustDummy(),
		nconn:     make(chan net.Conn),
		knockChan: make(chan KnockGrouper),
		tls:       make(TLS),
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	o := SniffAndFilterOpts{
		EventChan:          c.events,
		KnockChan:          c.knockChan,
		BlockPorts:         c.BlockPorts,
		BlockSourceIP:      c.BlockSourceIP,
		BlockDestinationIP: c.BlockDestinationIP,
	}

	// Create events on NIC traffic.
	eventWrapper := func(lower stack.LinkEndpoint) stack.LinkEndpoint {
		return WrapLinkEndpoint(lower, o)
	}

	s, restoreFunc, err := SetupNetworkStack(nil, c.Interfaces, eventWrapper)
	if err != nil {
		return nil, err
	}
	c.restoreDevices = restoreFunc

	log.Debugf("s.NICInfo() = %+v\n", s.NICInfo())

	c.stack = s

	// // set the default tls.Config
	// if tconf, err := tlsconf(c.CertificateFile, c.KeyFile); err != nil {
	// 	log.Debugf("No default TLS config set: %v", err)
	// } else {
	// 	c.tls = &TLS{}
	// 	c.tls.AddConfig(0, tconf)
	// }

	return c, nil
}

func (c *Canary) Accept() (net.Conn, error) {
	conn := <-c.nconn
	return conn, nil
}

func (c *Canary) SetChannel(ch pushers.Channel) {
	c.events = ch
}

func (c *Canary) Close() error {
	log.Debugf("s.NICInfo() = %+v\n", c.stack.NICInfo())

	c.stack.Close()

	// log.Debug("restoring network device(s)...")
	// err := c.restoreDevices()
	// if err != nil {
	// 	return fmt.Errorf("restore network devices has error: %v", err)
	// }

	// log.Debug("restoring network device(s) successfull")
	return nil
}

func (c *Canary) Start(ctx context.Context) error {
	defer func() {
		log.Debug("restoring network device(s)...")
		err := c.restoreDevices()
		if err != nil {
			log.Errorf("restore network devices has error: %v", err)
		}

		log.Debug("restoring network device(s) successfull")
	}()

	go RunKnockDetector(ctx, c.knockChan, c.events)

	canHandle := func(addr2 net.Addr) bool {
		for _, addr := range c.serviceAddrs {
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
		log.Debugf("no service defined for: %s", addr2.String())

		return false
	}

	tcpForwarder := tcp.NewForwarder(c.stack, 30000, 5000, func(r *tcp.ForwarderRequest) {
		// got syn
		id := r.ID()

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

		conn, terr := c.tls.MaybeTLS(gonet.NewTCPConn(&wq, ep), id.LocalPort, c.events)
		if terr != nil {
			log.Errorf("maybe tls: %s", terr)
			r.Complete(false)
		}

		// check for ports to ignore
		if !canHandle(conn.LocalAddr()) {
			conn.Close()
			r.Complete(false)
			return
		}

		c.nconn <- conn
	})

	c.stack.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

	udpForwarder := NewUDPForwarder(c.stack, func(fr *UDPForwarderRequest) {
		id := fr.ID()

		if !canHandle(
			&net.UDPAddr{
				IP:   net.IP(id.LocalAddress),
				Port: int(id.LocalPort),
			}) {
			return
		}

		c.nconn <- &listener.DummyUDPConn{
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

	c.stack.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
	/*

		for _, address := range c.serviceAddrs {
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

						//TODO (jerry): Not thread safe, need to copy buf.
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
	*/

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

func tlsconf(certFile, keyFile string) (*tls.Config, error) {
	var pemCert, pemKey bytes.Buffer

	file, err := os.Open(certFile)
	if err != nil {
		return nil, fmt.Errorf("open(%s): %v", certFile, err)
	}
	io.Copy(&pemCert, file)
	file.Close()

	file, err = os.Open(keyFile)
	if err != nil {
		return nil, fmt.Errorf("open(%s): %v", keyFile, err)
	}
	io.Copy(&pemKey, file)
	file.Close()

	cert, err := tls.X509KeyPair(pemCert.Bytes(), pemKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed setting TLS config: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
