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
