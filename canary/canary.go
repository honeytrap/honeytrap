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
	// MaxEpollEvents defines maximum number of poll events to retrieve at once
	MaxEpollEvents = 2048
	// DefaultBufferSize defines size of receive buffer
	DefaultBufferSize = 65535
)

const (
	// EthernetTypeIPv4 is the protocol number for IPv4 traffic
	EthernetTypeIPv4 = 0x0800
	// EthernetTypeIPv6 is the protocol number for IPv6 traffic
	EthernetTypeIPv6 = 0x86DD
	// EthernetTypeARP is the protocol number for ARP traffic
	EthernetTypeARP = 0x0806
)

// Protocol specifies the network protocol
type Protocol int

const (
	// ProtocolTCP specifies tcp protocol
	ProtocolTCP Protocol = iota
	// ProtocolUDP specifies udp protocol
	ProtocolUDP
	// ProtocolICMP specifies icmp protocol
	ProtocolICMP
)

// Canary contains the canary struct
type Canary struct {
	epfd int

	m sync.Mutex

	r *rand.Rand

	knockChan chan interface{}

	networkInterfaces []net.Interface

	events pushers.Channel
}

// KnockGroup groups multiple knocks
type KnockGroup struct {
	Start time.Time
	Last  time.Time

	SourceIP net.IP
	Protocol Protocol

	Count int

	Knocks *UniqueSet
}

// KnockGrouper defines the interface for NewGroup function
type KnockGrouper interface {
	NewGroup() *KnockGroup
}

// KnockUDPPort struct contain UDP port knock metadata
type KnockUDPPort struct {
	SourceIP        net.IP
	DestinationPort uint16
}

// NewGroup will return a new KnockGroup for UDP protocol
func (k KnockUDPPort) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:    time.Now(),
		SourceIP: k.SourceIP,
		Count:    0,
		Protocol: ProtocolUDP,
		Knocks: NewUniqueSet(func(v1, v2 interface{}) bool {
			if _, ok := v1.(KnockUDPPort); !ok {
				return false
			}
			if _, ok := v2.(KnockUDPPort); !ok {
				return false
			}

			k1, k2 := v1.(KnockUDPPort), v2.(KnockUDPPort)
			return k1.DestinationPort == k2.DestinationPort
		}),
	}
}

// KnockTCPPort struct contain TCP port knock metadata
type KnockTCPPort struct {
	SourceIP        net.IP
	DestinationPort uint16
}

// NewGroup will return a new KnockGroup for TCP protocol
func (k KnockTCPPort) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:    time.Now(),
		SourceIP: k.SourceIP,
		Protocol: ProtocolTCP,
		Count:    0,
		Knocks: NewUniqueSet(func(v1, v2 interface{}) bool {
			if _, ok := v1.(KnockTCPPort); !ok {
				return false
			}
			if _, ok := v2.(KnockTCPPort); !ok {
				return false
			}

			k1, k2 := v1.(KnockTCPPort), v2.(KnockTCPPort)
			return k1.DestinationPort == k2.DestinationPort
		}),
	}
}

// KnockICMP struct contain ICMP knock metadata
type KnockICMP struct {
	SourceIP net.IP
}

// NewGroup will return a new KnockGroup for ICMP protocol
func (k KnockICMP) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:    time.Now(),
		SourceIP: k.SourceIP,
		Count:    0,
		Protocol: ProtocolICMP,
		Knocks: NewUniqueSet(func(v1, v2 interface{}) bool {
			if _, ok := v1.(KnockICMP); !ok {
				return false
			}
			if _, ok := v2.(KnockICMP); !ok {
				return false
			}

			_, _ = v1.(KnockICMP), v2.(KnockICMP)
			return true
		}),
	}
}

// Taken from https://github.com/xiezhenye/harp/blob/master/src/arp/arp.go#L53
func htons(n uint16) uint16 {
	var (
		high = n >> 8
		ret  = n<<8 + high
	)

	return ret
}

// handleUDP will handle udp packets
func (c *Canary) handleUDP(iph *ipv4.Header, data []byte) error {
	hdr, err := udp.Unmarshal(data)
	if err != nil {
		return nil
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	// detect if our interface initiated or portscan
	c.knockChan <- KnockUDPPort{
		SourceIP:        iph.Src,
		DestinationPort: hdr.Destination,
	}

	return nil
}

// handleTCP will handle tcp packets
func (c *Canary) handleICMP(iph *ipv4.Header, data []byte) error {
	_, err := icmp.Parse(data)
	if err != nil {
		return err
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	c.knockChan <- KnockICMP{
		SourceIP: iph.Src,
	}
	return nil
}

// handleARP will handle arp packets
func (c *Canary) handleARP(data []byte) error {
	arp, err := arp.Parse(data)
	if err != nil {
		return err
	}

	_ = arp

	// check if it is my hardware address or broadcast
	/*
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
	*/

	return nil
}

// isMe returns if the ip is one of our interfaces addresses
func (c *Canary) isMe(ip net.IP) bool {
	for _, intf := range c.networkInterfaces {
		addrs, _ := intf.Addrs()

		for _, addr := range addrs {
			if ip4net, ok := addr.(*net.IPNet); !ok {
			} else if ip4net.IP.Equal(ip) {
				return true
			}
		}
	}

	return false
}

// handleTCP will handle tcp packets
func (c *Canary) handleTCP(iph *ipv4.Header, data []byte) error {
	hdr, err := tcp.UnmarshalWithChecksum(data, iph.Dst, iph.Src)
	if err == tcp.ErrInvalidChecksum {
		// we are ignoring invalid checksums for now
	} else if err != nil {
		return err
	}

	if hdr.Ctrl&tcp.SYN != tcp.SYN {
		return nil
	} else if hdr.Ctrl&tcp.ACK != 0 {
		return nil
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	c.knockChan <- KnockTCPPort{
		SourceIP:        iph.Src,
		DestinationPort: hdr.Destination,
	}

	return nil
}

// New will return a Canary for specified interfaces. Events will be delivered through
// events
func New(interfaces []net.Interface, events pushers.Channel) (*Canary, error) {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("epoll_create1: %s", err.Error())
	}

	networkInterfaces := []net.Interface{}

	for _, intf := range interfaces {
		if fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL))); err != nil {
			return nil, fmt.Errorf("Could not create socket: %s", err.Error())
		} else if fd < 0 {
			return nil, fmt.Errorf("Socket error: return < 0")
		} else if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLL_NONBLOCK,
			Fd:     int32(fd),
		}); err != nil {
			return nil, fmt.Errorf("epollctl: %s", err.Error())
		}

		networkInterfaces = append(networkInterfaces, intf)
	}

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	return &Canary{
		epfd:              epfd,
		networkInterfaces: networkInterfaces,
		r:                 r,
		knockChan:         make(chan interface{}, 100),
		events:            events,
	}, nil
}

// Close will close the canary
func (c *Canary) Close() {
	syscall.Close(c.epfd)
}

func (c *Canary) knockDetector() {
	knocks := NewUniqueSet(func(v1, v2 interface{}) bool {
		k1, k2 := v1.(*KnockGroup), v2.(*KnockGroup)
		if k1.Protocol != k2.Protocol {
			return false
		}

		return bytes.Compare(k1.SourceIP, k2.SourceIP) == 0
	})

	for {
		select {
		case sk := <-c.knockChan:
			grouper := sk.(KnockGrouper)
			knock := knocks.Add(grouper.NewGroup()).(*KnockGroup)

			knock.Count++
			knock.Last = time.Now()

			knock.Knocks.Add(sk)

		case <-time.After(time.Second * 5):
			// TODO: make time configurable

			now := time.Now()

			knocks.Each(func(i int, v interface{}) {
				k := v.(*KnockGroup)

				if k.Count > 100 {
				} else if k.Last.Add(time.Second * 5).After(now) {
					return
				}

				defer knocks.Remove(k)

				ports := make([]string, k.Knocks.Count())

				k.Knocks.Each(func(i int, v interface{}) {
					if k, ok := v.(KnockTCPPort); ok {
						ports[i] = fmt.Sprintf("tcp/%d", k.DestinationPort)
					} else if k, ok := v.(KnockUDPPort); ok {
						ports[i] = fmt.Sprintf("udp/%d", k.DestinationPort)
					} else if _, ok := v.(KnockICMP); ok {
						ports[i] = fmt.Sprintf("icmp")
					}
				})

				c.events.Send(EventPortscan(k.SourceIP, k.Last.Sub(k.Start), k.Count, ports))
			})
		}
	}
}

// EventPortscan will return a portscan event struct
func EventPortscan(sourceIP net.IP, duration time.Duration, count int, ports []string) message.Event {
	return message.Event{
		Sensor:   "Canary",
		Category: "Portscan",
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"message": fmt.Sprintf("Port %d touch(es) detected from %s with duration %+v: %s\n", count, sourceIP, duration, strings.Join(ports, ", ")),
		},
	}
}

// send will queue a packet for sending
func (c *Canary) send(data []byte) error {
	c.m.Lock()
	defer c.m.Unlock()

	return nil
}

// send will queue a packet for sending
func (c *Canary) transmit(fd int32) {
	// os specific transmitter
	// protocol implementation specific

	// simple HTTP

	// record all kind of challenge responses
	// os fingerprint

	c.m.Lock()
	defer c.m.Unlock()
}

// Run will start Canary
func (c *Canary) Run() error {
	go c.knockDetector()

	var (
		events [MaxEpollEvents]syscall.EpollEvent
		buffer [DefaultBufferSize]byte
	)

	for {
		nevents, err := syscall.EpollWait(c.epfd, events[:], -1)
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
					c.handleARP(eh.Payload[:])
				} else if eh.Type == EthernetTypeIPv4 {
					if iph, err := ipv4.Parse(eh.Payload[:]); err != nil {
						fmt.Printf("Error parsing ip header: %s\n", err.Error())
					} else {
						data := iph.Payload[:]

						switch iph.Protocol {
						case 1 /* icmp */ :
							c.handleICMP(iph, data)
						case 6 /* tcp */ :
							c.handleTCP(iph, data)
						case 17 /* udp */ :
							c.handleUDP(iph, data)
						default:
							fmt.Printf("Unknown protocol: %x\n", iph.Protocol)
						}
					}
				}
			}

			if events[ev].Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
				// should we use the network interface fd, or just events[ev]Fd?
				c.transmit(events[ev].Fd)

				// disable epollout again
				syscall.EpollCtl(c.epfd, syscall.EPOLL_CTL_MOD, int(events[ev].Fd), &syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(events[ev].Fd),
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
