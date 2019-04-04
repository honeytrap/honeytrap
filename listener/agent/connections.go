// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package agent

import (
	"net"
	"sync"
)

type Connections struct {
	m     sync.Mutex
	conns []*agentConnection
}

func (c *Connections) Add(ac *agentConnection) {
	c.m.Lock()
	defer c.m.Unlock()

	c.conns = append(c.conns, ac)
}

func (c *Connections) Each(fn func(ac *agentConnection)) {
	c.m.Lock()
	defer c.m.Unlock()

	for index := 0; index < len(c.conns); index++ {
		fn(c.conns[index])
	}
}

func (c *Connections) Delete(ac *agentConnection) {
	c.m.Lock()
	defer c.m.Unlock()

	for index := 0; index < len(c.conns); index++ {
		if c.conns[index] != ac {
			continue
		}

		c.conns = append(c.conns[:index], c.conns[index+1:]...)
		return
	}
}

func (c *Connections) Get(laddr net.Addr, raddr net.Addr) *agentConnection {
	c.m.Lock()
	defer c.m.Unlock()

	for _, conn := range c.conns {
		if conn.Laddr.String() != laddr.String() {
			continue
		}

		if conn.Raddr.String() != raddr.String() {
			continue
		}

		return conn
	}

	return nil
}
