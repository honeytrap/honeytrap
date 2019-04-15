// +build lxc
// +build linux

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
package lxc

import (
	"fmt"
	"net"
	"time"
)

// lxcContainerConn defines a custom connection type which proxies the data
// for the container.
type lxcContainerConn struct {
	net.Conn
	container *lxcContainer
}

// Read reads the giving set of data from the container connection to the
// byte slice.
func (c lxcContainerConn) Read(b []byte) (n int, err error) {
	c.container.stillActive()
	return c.Conn.Read(b)
}

// Write writes the data into byte slice from the container.
func (c lxcContainerConn) Write(b []byte) (n int, err error) {
	c.container.stillActive()
	return c.Conn.Write(b)
}

// stillActive returns an error if the containerr is not still active
func (c *lxcContainer) stillActive() error {
	if c.isStopped() {
		return fmt.Errorf("lxccontainer not running %s", c.name)
	}
	if c.isFrozen() {
		return c.unfreeze()
	}
	if !c.isRunning() {
		return fmt.Errorf("lxccontainer in unknown state %s:%s", c.name, c.c.State())
	}
	c.idle = time.Now()
	return nil
}
