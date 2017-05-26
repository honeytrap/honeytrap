package canary

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/honeytrap/honeytrap/canary/arp"
	"github.com/honeytrap/honeytrap/canary/ethernet"
	"github.com/honeytrap/honeytrap/canary/icmp"
	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/tcp"
	"github.com/honeytrap/honeytrap/canary/udp"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	MaxEpollEvents    = 2048
	DefaultBufferSize = 65535
)

const (
	EthernetTypeIPv4 = 0x0800
	EthernetTypeIPv6 = 0x86DD
	EthernetTypeARP  = 0x0806
)

type Canary struct {
	fd   int
	epfd int

	m sync.Mutex

	r *rand.Rand

	knockChan        chan SingleKnock
	networkInterface *net.Interface

	events pushers.Events
}

// knock?
type Knock struct {
	Start time.Time
	Last  time.Time

	SourceIP net.IP

	Knocks *UniqueSet
}

type SingleKnock struct {
	SourceIP net.IP

	DestinationPort uint16
}

// Taken from https://github.com/xiezhenye/harp/blob/master/src/arp/arp.go#L53
func htons(n uint16) uint16 {
	var (
		high uint16 = n >> 8
		ret  uint16 = n<<8 + high
	)

	return ret
}

func (c *Canary) handleUDP(iph *ipv4.Header, data []byte) error {
	hdr, err := udp.Unmarshal(data)
	if err != nil {
		return nil
	}

	_ = hdr
	// fmt.Printf("udp: src=%s, dst=%s, %s\n", iph.Src.String(), iph.Dst.String(), hdr)
	return nil
}

func (c *Canary) handleICMP(iph *ipv4.Header, data []byte) error {
	hdr, err := icmp.Parse(data)
	if err != nil {
		return err
	}

	fmt.Printf("icmp: src=%s, dst=%s, %s\n", iph.Src.String(), iph.Dst.String(), hdr)
	return nil
}

func (c *Canary) handleARP(data []byte) error {
	arp, err := arp.Parse(data)
	if err != nil {
		return err
	}

	// check if it is my hardware address or broadcast
	if bytes.Compare(arp.TargetMAC[:], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) == 0 {
		if arp.Opcode == 0x01 {
			fmt.Printf("ARP: Who has %s? tell %s.\n", net.IPv4(arp.TargetIP[0], arp.TargetIP[1], arp.TargetIP[2], arp.TargetIP[3]).String(), net.IPv4(arp.SenderIP[0], arp.SenderIP[1], arp.SenderIP[2], arp.SenderIP[3]).String())
		}
	} else if bytes.Compare(arp.TargetMAC[:], c.networkInterface.HardwareAddr) == 0 {
		if arp.Opcode == 0x01 {
			fmt.Printf("ARP: Who has %s? tell %s.\n", net.IPv4(arp.TargetIP[0], arp.TargetIP[1], arp.TargetIP[2], arp.TargetIP[3]).String(), net.IPv4(arp.SenderIP[0], arp.SenderIP[1], arp.SenderIP[2], arp.SenderIP[3]).String())
		} else {
		}
	} else {
		fmt.Println("ARP: not for me")
	}

	return nil
}

func (c *Canary) handleTCP(iph *ipv4.Header, data []byte) error {
	hdr, err := tcp.UnmarshalWithChecksum(data, iph.Dst, iph.Src)
	if err == tcp.ErrInvalidChecksum {
		// we are ignoring invalid checksums for now
	} else if err != nil {
		return err
	}

	if hdr.Source == 22 || hdr.Destination == 22 {
		return nil
	}

	myself := false

	if addrs, err := c.networkInterface.Addrs(); err != nil {
	} else {
		for _, addr := range addrs {
			myself = myself || addr.(*net.IPNet).Contains(iph.Dst)
		}
	}

	_ = myself

	if hdr.Ctrl&tcp.SYN != tcp.SYN {
		return nil
	} else if hdr.Ctrl&tcp.ACK != 0 {
		return nil
	}

	if iph.Dst.String() != "172.16.84.159" {
		return nil
	}

	c.knockChan <- SingleKnock{
		SourceIP:        iph.Src,
		DestinationPort: hdr.Destination,
	}

	return nil
}

func New(intf string, events pushers.Events) (*Canary, error) {
	if networkInterface, err := net.InterfaceByName(intf); err != nil {
		return nil, fmt.Errorf("The selected network interface %s does not exist.\n", intf)
	} else if fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL))); err != nil {
		return nil, fmt.Errorf("Could not create socket: %s", err.Error())
	} else if fd < 0 {
		return nil, fmt.Errorf("Socket error: return < 0")
	} else if epfd, err := syscall.EpollCreate1(0); err != nil {
		return nil, fmt.Errorf("epoll_create1: %s", err.Error())
	} else if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLL_NONBLOCK,
		Fd:     int32(fd),
	}); err != nil {
		return nil, fmt.Errorf("epollctl: %s", err.Error())
	} else {
		r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

		return &Canary{
			fd:               fd,
			epfd:             epfd,
			networkInterface: networkInterface,
			r:                r,
			knockChan:        make(chan SingleKnock, 100),
			events:           events,
		}, nil
	}
}

func (s *Canary) Close() {
	syscall.Close(s.epfd)
	syscall.Close(s.fd)
}

func (s *Canary) knockDetector() {
	knocks := NewUniqueSet(func(v1, v2 interface{}) bool {
		k1, k2 := v1.(*Knock), v2.(*Knock)
		return bytes.Compare(k1.SourceIP, k2.SourceIP) == 0
	})

	for {
		select {
		case sk := <-s.knockChan:
			knock := knocks.Add(&Knock{
				Start:    time.Now(),
				SourceIP: sk.SourceIP,
				Knocks: NewUniqueSet(func(v1, v2 interface{}) bool {
					k1, k2 := v1.(uint16), v2.(uint16)
					return k1 == k2
				}),
			}).(*Knock)

			knock.Last = time.Now()

			knock.Knocks.Add(sk.DestinationPort)

		case <-time.After(time.Second * 5):
			// TODO: make time configurable

			now := time.Now()

			knocks.Each(func(i int, v interface{}) {
				k := v.(*Knock)

				if k.Last.Add(time.Second * 5).After(now) {
					return
				}

				defer knocks.Remove(k)

				ports := make([]string, k.Knocks.Count())

				k.Knocks.Each(func(i int, v interface{}) {
					ports[i] = fmt.Sprintf("tcp/%d", v.(uint16))
				})

				s.events.Deliver(EventPortscan(k.SourceIP, k.Last.Sub(k.Start), ports))
			})
		}
	}
}

// EventPortscan will return a portscan event struct
func EventPortscan(sourceIP net.IP, duration time.Duration, ports []string) message.Event {
	return message.Event{
		Sensor:   "Canary",
		Category: "Portscan",
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("Port touch(es) detected from %s with duration %+v: %s\n", sourceIP, duration, strings.Join(ports, ", ")),
		},
	}
}

func (s *Canary) send(data []byte) error {
	s.m.Lock()
	defer s.m.Unlock()

	return nil
}

func (s *Canary) transmit(fd int32) {
	// os specific transmitter
	// protocol implementation specific

	// simple HTTP

	// record all kind of challenge responses
	// os fingerprint

	s.m.Lock()
	defer s.m.Unlock()
}

func (s *Canary) Run() error {
	go s.knockDetector()

	var (
		events [MaxEpollEvents]syscall.EpollEvent
		buffer [DefaultBufferSize]byte
	)

	for {
		nevents, err := syscall.EpollWait(s.epfd, events[:], -1)
		if err != nil {
			break
		}

		for ev := 0; ev < nevents; ev++ {
			if events[ev].Events&syscall.EPOLLIN == syscall.EPOLLIN {
				if n, _, err := syscall.Recvfrom(int(events[ev].Fd), buffer[:], 0); err != nil {
					fmt.Printf("Could not receive from descriptor: %s\n", err.Error())
					return err
				} else if n == 0 {
					// no packets received
				} else if eh, err := ethernet.Parse(buffer[:n]); err != nil {
				} else if eh.Type == EthernetTypeARP {
					s.handleARP(eh.Payload[:])
				} else if eh.Type == EthernetTypeIPv4 {
					if iph, err := ipv4.Parse(eh.Payload[:]); err != nil {
						fmt.Printf("Error parsing ip header: %s\n", err.Error())
					} else {
						data := iph.Payload[:]

						switch iph.Protocol {
						case 1 /* icmp */ :
							s.handleICMP(iph, data)
						case 6 /* tcp */ :
							s.handleTCP(iph, data)
						case 17 /* udp */ :
							s.handleUDP(iph, data)
						default:
							fmt.Printf("Unknown protocol: %x\n", iph.Protocol)
						}
					}
				}
			}

			if events[ev].Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
				s.transmit(events[ev].Fd)

				// disable epollout again
				syscall.EpollCtl(s.epfd, syscall.EPOLL_CTL_MOD, s.fd, &syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(s.fd),
				})
			}

			if events[ev].Events&syscall.EPOLLERR == syscall.EPOLLERR {
				if v, err := syscall.GetsockoptInt(int(events[ev].Fd), syscall.SOL_SOCKET, syscall.SO_ERROR); err != nil {
					fmt.Println("Error", err)
				} else {
					fmt.Println("Error val", v)
				}
			}
		}
	}

	return nil
}
