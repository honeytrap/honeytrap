package canary

import (
	"bytes"
	"fmt"
	"time"
)

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
