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

				c.events.Send(EventPortscan(k.SourceIP, k.DestinationIP, k.Last.Sub(k.Start), k.Count, ports))
			})
		}
	}
}
