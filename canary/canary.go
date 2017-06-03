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

	"github.com/glycerine/rbuf"
	"github.com/honeytrap/honeytrap/canary/arp"
	"github.com/honeytrap/honeytrap/canary/ethernet"
	"github.com/honeytrap/honeytrap/canary/icmp"
	"github.com/honeytrap/honeytrap/canary/ipv4"
	"github.com/honeytrap/honeytrap/canary/tcp"
	"github.com/honeytrap/honeytrap/canary/udp"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/event"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("canary")

// first dns
// ntp
// send reset?
// udp check connect or answer
// parameters
// clean up old states
// check ring buffer
// use sockets and io.Reader
// parameters: ports to include exclude/ filter (or do we want to filter the events)
// interface to listen on

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
	rt RouteTable

	ac ARPCache

	epfd int

	m sync.Mutex

	r *rand.Rand

	knockChan chan interface{}

	networkInterfaces []net.Interface

	events pushers.Channel

	descriptors map[string]int32

	buffer *rbuf.FixedSizeRingBuf

	stateTable StateTable
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

	// check if we have udp listeners on specified port, and answer otherwise
	// parse udp
	// we should check if the received packet is a response or request
	// detect if our interface initiated or portscan

	handlers := map[uint16]func(*ipv4.Header, *udp.Header) error{
		53:   c.DecodeDNS,
		123:  c.DecodeNTP,
		1900: c.DecodeSSDP,
		5060: c.DecodeSIP,
		161:  c.DecodeSNMP,
		162:  c.DecodeSNMPTrap,
	}

	if fn, ok := handlers[hdr.Destination]; !ok {
		// default handler
		c.knockChan <- KnockUDPPort{
			SourceIP:        iph.Src,
			DestinationPort: hdr.Destination,
		}

		// do we only want to detect scans? Or also detect payloads?
		c.events.Send(EventUDP(iph.Src, iph.Dst, hdr.Source, hdr.Destination, hdr.Payload))
	} else if err := fn(iph, hdr); err != nil {
		fmt.Printf("Could not decode udp packet: %s", err)
	}

	return nil
}

// handleICMP will handle tcp packets
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
	if bytes.Compare(arp.TargetMAC[:], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) == 0 {
		if arp.Opcode == 0x01 {
			// fmt.Printf("ARP: Who has %s? tell %s.", net.IPv4(arp.TargetIP[0], arp.TargetIP[1], arp.TargetIP[2], arp.TargetIP[3]).String(), net.IPv4(arp.SenderIP[0], arp.SenderIP[1], arp.SenderIP[2], arp.SenderIP[3]).String())
		}
		return nil
	}

	for _, networkInterface := range c.networkInterfaces {
		if bytes.Compare(arp.TargetMAC[:], networkInterface.HardwareAddr) == 0 {
			if arp.Opcode == 0x01 {
				// fmt.Printf("ARP: Who has %s? tell %s.", net.IPv4(arp.TargetIP[0], arp.TargetIP[1], arp.TargetIP[2], arp.TargetIP[3]).String(), net.IPv4(arp.SenderIP[0], arp.SenderIP[1], arp.SenderIP[2], arp.SenderIP[3]).String())
			} else {
			}
		} else {
			// 			fmt.Println("ARP: not for me")
		}
	}

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

	if !c.isMe(iph.Dst) {
		return nil
	}

	if hdr.Source == 22 || hdr.Destination == 22 {
		return nil
	}

	state := c.stateTable.Get(iph.Src, iph.Dst, hdr.Source, hdr.Destination)
	if state != nil {
	} else if hdr.HasFlag(tcp.SYN) && !hdr.HasFlag(tcp.ACK) {
		// no state found
		state = NewState(iph.Src, hdr.Source, iph.Dst, hdr.Destination)
		c.stateTable.Add(state)
	} else {
		// no existing state found, returning
		return nil // ErrNoExistingStateFound()
	}

	// USE PUSH?
	// ACK THE PAYLOAD ALWAYS

	// ACK EACH SYN, PSH, FIN AND RST
	switch {
	case hdr.HasFlag(tcp.SYN):
		fallthrough
	case hdr.HasFlag(tcp.RST):
		fallthrough
	case hdr.HasFlag(tcp.FIN):
		fallthrough
	case hdr.HasFlag(tcp.PSH):
		c.ack(state, iph, hdr)
	}

	if hdr.Ctrl&tcp.SYN == tcp.SYN {
		c.knockChan <- KnockTCPPort{
			SourceIP:        iph.Src,
			DestinationPort: hdr.Destination,
		}
	} else if hdr.Ctrl&tcp.PSH == tcp.PSH {
		handlers := map[uint16]func(*ipv4.Header, *tcp.Header) error{
			80: c.DecodeHTTP,
		}

		if fn, ok := handlers[hdr.Destination]; !ok {
			c.events.Send(EventTCPPayload(iph.Src, hdr.Destination, string(hdr.Payload)))
		} else if err := fn(iph, hdr); err != nil {
			fmt.Printf("Could not decode tcp packet: %s", err)
		}

		return nil
	} else {
		// FIN / RST
		return nil
	}

	// check if we have tcp listeners on specified port, and answer otherwise
	return nil
}
func (c *Canary) ack(state *State, iph *ipv4.Header, tcph *tcp.Header) error {
	seqNum := tcph.SeqNum + uint32(len(tcph.Payload))
	flags := tcp.Flag(tcp.ACK)

	if tcph.HasFlag(tcp.SYN) {
		seqNum++
		flags |= tcp.Flag(tcp.SYN)
	} else if tcph.HasFlag(tcp.RST) {
		seqNum++
		flags |= tcp.Flag(tcp.RST)
	} else if tcph.HasFlag(tcp.FIN) {
		seqNum++
		flags |= tcp.Flag(tcp.FIN)
	}

	// TODO: keep state....
	// SeqNum
	// ID

	th := &tcp.Header{
		Source:      tcph.Destination,
		Destination: tcph.Source,
		SeqNum:      state.SendNext,
		AckNum:      seqNum,
		Reserved:    0,
		ECN:         0,
		Ctrl:        flags,
		Window:      65531,
		Checksum:    0,
		Urgent:      0,
		Options:     []tcp.Option{},
		Payload:     []byte{},
	}

	data1, err := th.Marshal()
	if err != nil {
		return err
	}

	// ack the received packet
	iph2 := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		Flags:    0,
		FragOff:  0,
		TTL:      128,
		Src:      iph.Dst,
		Dst:      iph.Src,
		ID:       int(state.ID), // state.ID() which will increment automatically
		Protocol: 6,
		TotalLen: 20 + len(data1),
	}

	data, err := iph2.Marshal()
	if err != nil {
		return err
	}

	state.ID++
	// we don't have to increate sendNext for ACK
	// state.SendNext++

	if tcph.HasFlag(tcp.SYN) {
		state.SendNext++
	} else if tcph.HasFlag(tcp.RST) {
		state.SendNext++
	} else if tcph.HasFlag(tcp.FIN) {
		state.SendNext++
	}

	updateTCPChecksum(iph2, data1)

	data = append(data, data1...)

	// Src := net.IPv4(data1[12], data1[13], data1[14], data1[15])
	Dst := net.IPv4(data[16], data[17], data[18], data[19])

	ef := ethernet.EthernetFrame{
		Source:      c.networkInterfaces[0].HardwareAddr,
		Destination: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Type:        0x0800,
	}

	intf := ""

	ae := c.ac.Get(Dst)
	if ae != nil {
		ef.Destination = ae.HardwareAddress
		intf = ae.Interface
	} else {
		// TODO(make function)
		for _, route := range c.rt {
			// find shortest route
			if !route.Destination.Contains(Dst) {
				continue
			}

			a := c.ac.Get(route.Gateway)

			intf = a.Interface
			ef.Destination = a.HardwareAddress
			break
		}

	}

	data2, err := ef.Marshal()
	if err != nil {
		fmt.Println("Error marshalling ethernet frame: ", err)
	}

	csum := uint32(0)

	// calculate correct ip header length here.
	length := 20 // len(data1) - 1

	// calculate options?

	/*
		i := length

		for {
			if i > len(data) {
				break
			}

			if data[i] == 0x00 {
				break
			}

			fmt.Println("Got option")

			length += int(data[i+1])
			i += int(data[i+1])
		}
	*/

	for i := 0; i < length; i += 2 {
		if i == 10 {
			continue
		}

		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}

	for {
		// Break when sum is less or equals to 0xFFFF
		if csum <= 65535 {
			break
		}
		// Add carry to the sum
		csum = (csum >> 16) + uint32(uint16(csum))
	}

	csum = uint32(^uint16(csum))

	data[10] = uint8((csum >> 8) & 0xFF)
	data[11] = uint8(csum & 0xFF)

	data = append(data2, data...)

	c.send(intf, data)

	return nil
}

// Count occurrences in s of any bytes in t.
func countAnyByte(s string, t string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			n++
		}
	}
	return n
}

// Split s at any bytes in t.
func splitAtBytes(s string, t string) []string {
	a := make([]string, 1+countAnyByte(s, t))
	n := 0
	last := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			if last < i {
				a[n] = s[last:i]
				n++
			}
			last = i + 1
		}
	}
	if last < len(s) {
		a[n] = s[last:]
		n++
	}
	return a[0:n]
}

// New will return a Canary for specified interfaces. Events will be delivered through
// events
func New(interfaces []net.Interface, events pushers.Channel) (*Canary, error) {
	rt, err := parseRouteTable("/proc/net/route")
	if err != nil {
		return nil, fmt.Errorf("Could not parse route table: %s", err.Error())
	}

	ac, err := parseARPCache("/proc/net/arp")
	if err != nil {
		return nil, fmt.Errorf("Could not parse arp cache: %s", err.Error())
	}

	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("epoll_create1: %s", err.Error())
	}

	networkInterfaces := []net.Interface{}
	descriptors := map[string]int32{}

	for _, intf := range interfaces {
		if intf.Name != "ens160" && intf.Name != "eth0" && intf.Name != "ens3" {
			continue
		}

		fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		if err != nil {
			return nil, fmt.Errorf("Could not create socket: %s", err.Error())
		}

		if fd < 0 {
			return nil, fmt.Errorf("Socket error: return < 0")
		}

		if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLL_NONBLOCK,
			Fd:     int32(fd),
		}); err != nil {
			return nil, fmt.Errorf("epollctl: %s", err.Error())
		}

		descriptors[intf.Name] = int32(fd)
		networkInterfaces = append(networkInterfaces, intf)
	}

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	return &Canary{
		ac:                ac,
		rt:                rt,
		epfd:              epfd,
		descriptors:       descriptors,
		networkInterfaces: networkInterfaces,
		r:                 r,
		knockChan:         make(chan interface{}, 100),
		events:            events,

		buffer: rbuf.NewFixedSizeRingBuf(65535),
	}, nil
}

// Close will close the canary
func (c *Canary) Close() {
	syscall.Close(c.epfd)
}

var (
	// EventCategoryPortscan contains events for ssdp traffic
	EventCategoryPortscan = event.Category("portscan")
)

// EventPortscan will return a portscan event struct
func EventPortscan(sourceIP net.IP, duration time.Duration, count int, ports []string) event.Event {
	// TODO: do something different with message
	return event.New(
		CanaryOptions,
		EventCategoryPortscan,
		event.ServiceStarted,
		event.Custom("source-ip", sourceIP.String()),
		event.Custom("portscan.ports", ports),
		event.Custom("portscan.duration", duration),
		event.Custom("message", fmt.Sprintf("Port %d touch(es) detected from %s with duration %+v: %s", count, sourceIP, duration, strings.Join(ports, ", "))),
	)
}

// send will queue a packet for sending
func (c *Canary) send(intf string, data []byte) error {
	c.m.Lock()
	defer c.m.Unlock()

	// add to send queue
	// ring buffer?

	// how does the buffer play nice with retransmits

	c.buffer.Write(data)

	// enable poll out
	// find interface

	for _, intf := range c.networkInterfaces {
		// send network frame
		// find gateway
		//

		_ = intf
	}

	fd := c.descriptors[intf]

	err := syscall.EpollCtl(c.epfd, syscall.EPOLL_CTL_MOD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		Fd:     int32(fd),
	})

	return err
}

func updateTCPChecksum(iph *ipv4.Header, data []byte) {
	length := len(data)

	csum := uint32(0)

	csum += (uint32(iph.Src[12]) + uint32(iph.Src[14])) << 8
	csum += uint32(iph.Src[13]) + uint32(iph.Src[15])
	csum += (uint32(iph.Dst[12]) + uint32(iph.Dst[14])) << 8
	csum += uint32(iph.Dst[13]) + uint32(iph.Dst[15])

	csum += uint32(6)
	csum += uint32(length) & 0xffff
	csum += uint32(length) >> 16

	length = len(data) - 1

	// calculate correct ip header length here.
	for i := 0; i < length; i += 2 {
		if i == 16 {
			continue
		}

		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}

	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}

	csum = uint32(^uint16(csum + (csum >> 16)))

	data[16] = uint8((csum >> 8) & 0xFF)
	data[17] = uint8(csum & 0xFF)
}

// send will queue a packet for sending
func (c *Canary) transmit(fd int32) error {
	// os specific transmitter
	// protocol implementation specific

	// simple HTTP

	// record all kind of challenge responses
	// os fingerprint

	c.m.Lock()
	defer c.m.Unlock()

	buffer := make([]byte, 65535)
	n, err := c.buffer.Read(buffer)
	if err != nil {
		fmt.Println("BLA", err)
	}

	to := &syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  c.networkInterfaces[0].Index,
	}

	err = syscall.Sendto((int(fd)), buffer[:n], 0, to)
	if err != nil {
		panic(err)
	}

	return nil
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
					fmt.Printf("Could not receive from descriptor: %s", err.Error())
					return err
				} else if n == 0 {
					// no packets received
				} else if eh, err := ethernet.Parse(buffer[:n]); err != nil {
				} else if eh.Type == EthernetTypeARP {
					c.handleARP(eh.Payload[:])
				} else if eh.Type == EthernetTypeIPv4 {
					if iph, err := ipv4.Parse(eh.Payload[:]); err != nil {
						log.Debugf("Error parsing ip header: %s", err.Error())
					} else {
						data := iph.Payload[:]

						switch iph.Protocol {
						case 1 /* icmp */ :
							c.handleICMP(iph, data)
						case 2 /* IGMP */ :

						case 6 /* tcp */ :
							// what interface?
							c.handleTCP(iph, data)
						case 17 /* udp */ :
							c.handleUDP(iph, data)
						default:
							log.Debugf("Ignoring protocol: %x", iph.Protocol)
						}
					}
				}
			}

			if events[ev].Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
				c.transmit(events[ev].Fd)

				// disable epollout again
				syscall.EpollCtl(c.epfd, syscall.EPOLL_CTL_MOD, int(events[ev].Fd), &syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(events[ev].Fd),
				})
			}

			if events[ev].Events&syscall.EPOLLERR == syscall.EPOLLERR {
				if v, err := syscall.GetsockoptInt(int(events[ev].Fd), syscall.SOL_SOCKET, syscall.SO_ERROR); err != nil {
					log.Errorf("Error while retrieving polling error: %s", err)
				} else {
					log.Errorf("Polling error: %#q", v)
				}
			}
		}
	}

	return nil
}
