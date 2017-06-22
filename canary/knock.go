package canary

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/pushers/event"
)

var (
	// EventCategoryPortscan contains events for ssdp traffic
	EventCategoryPortscan = event.Category("portscan")
)

// KnockGroup groups multiple knocks
type KnockGroup struct {
	Start time.Time
	Last  time.Time

	SourceIP      net.IP
	DestinationIP net.IP

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
	DestinationIP   net.IP
	DestinationPort uint16
}

// NewGroup will return a new KnockGroup for UDP protocol
func (k KnockUDPPort) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:         time.Now(),
		SourceIP:      k.SourceIP,
		DestinationIP: k.DestinationIP,
		Count:         0,
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
	DestinationIP   net.IP
	DestinationPort uint16
}

// NewGroup will return a new KnockGroup for TCP protocol
func (k KnockTCPPort) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:         time.Now(),
		SourceIP:      k.SourceIP,
		DestinationIP: k.DestinationIP,
		Protocol:      ProtocolTCP,
		Count:         0,
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
	SourceIP      net.IP
	DestinationIP net.IP
}

// NewGroup will return a new KnockGroup for ICMP protocol
func (k KnockICMP) NewGroup() *KnockGroup {
	return &KnockGroup{
		Start:         time.Now(),
		SourceIP:      k.SourceIP,
		DestinationIP: k.DestinationIP,
		Count:         0,
		Protocol:      ProtocolICMP,
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

// EventPortscan will return a portscan event struct
func EventPortscan(src, dst net.IP, duration time.Duration, count int, ports []string) event.Event {
	// TODO: do something different with message
	return event.New(
		CanaryOptions,
		EventCategoryPortscan,
		event.ServiceStarted,
		event.SourceIP(src),
		event.DestinationIP(dst),
		event.Custom("portscan.ports", ports),
		event.Custom("portscan.duration", duration),
		event.Message(fmt.Sprintf("Port %d touch(es) detected from %s with duration %+v: %s", count, src, duration, strings.Join(ports, ", "))),
	)
}

func (c *Canary) knockDetector() {
	knocks := NewUniqueSet(func(v1, v2 interface{}) bool {
		k1, k2 := v1.(*KnockGroup), v2.(*KnockGroup)
		if k1.Protocol != k2.Protocol {
			return false
		}

		if bytes.Compare(k1.SourceIP, k2.SourceIP) != 0 {
			return false
		}

		return bytes.Compare(k1.DestinationIP, k2.DestinationIP) == 0
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
			now := time.Now()

			knocks.Each(func(i int, v interface{}) {
				k := v.(*KnockGroup)

				// TODO(): make duration configurable
				if k.Count > 100 {
					// we'll also bail out at a specific count
					// to prevent ddos
				} else if k.Last.Add(time.Second * 5).After(now) {
					return
				}

				// we have two timeouts, one to send notifications,
				// one to remove the knock. This will detect portscans
				// with a longer interval

				// TODO(): make duration configurable
				if k.Last.Add(time.Second * 60).After(now) {
					defer knocks.Remove(k)
				}

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

				c.events.Send(EventPortscan(k.SourceIP, k.DestinationIP, k.Last.Sub(k.Start), k.Count, ports))
			})
		}
	}
}
