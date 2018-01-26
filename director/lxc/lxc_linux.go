// +build lxc

/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package lxc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"time"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"

	"golang.org/x/sync/syncmap"

	"gopkg.in/lxc/go-lxc.v2"
)

var (
	_ = director.Register("lxc", New)
)

func New(options ...func(director.Director) error) (director.Director, error) {
	d := &lxcDirector{
		eb:       pushers.MustDummy(),
		lxcCh:    make(chan interface{}),
		template: "honeytrap",
	}

	for _, optionFn := range options {
		optionFn(d)
	}

	d.cache = &syncmap.Map{} // map[string]*lxcContainer{}
	go d.HandleLxcCommands()
	return d, nil
}

type LxcStop struct{ c *lxc.Container }
type LxcStart struct{ c *lxc.Container }
type LxcFreeze struct{ c *lxc.Container }
type LxcUnfreeze struct{ c *lxc.Container }

type lxcDirector struct {
	template string
	eb       pushers.Channel
	cache    *syncmap.Map // map[string]*lxcContainer
	lxcCh    chan interface{}
}

func (d *lxcDirector) HandleLxcCommands() {
	for {
		x := <-d.lxcCh
		switch cmd := x; cmd.(type) {
		case LxcStop:
			err := x.(LxcStop).c.Stop()
			if err != nil {
				log.Errorf("Error Stopping container: %s, because %s", x.(LxcStop).c.Name(), err.Error())
			}
		case LxcStart:
			err := x.(LxcStart).c.Start()
			if err != nil {
				log.Errorf("Error Starting container: %s, because %s", x.(LxcStart).c.Name(), err.Error())
			}
		case LxcFreeze:
			err := x.(LxcFreeze).c.Freeze()
			if err != nil {
				log.Errorf("Error Freezing container: %s, because %s", x.(LxcFreeze).c.Name(), err.Error())
			}
		case LxcUnfreeze:
			err := x.(LxcUnfreeze).c.Unfreeze()
			if err != nil {
				log.Errorf("Error UnFreezing container: %s, because %s", x.(LxcUnfreeze).c.Name(), err.Error())
			}
		}
	}
}

func (d *lxcDirector) SetChannel(eb pushers.Channel) {
	d.eb = eb
}

func (d *lxcDirector) Dial(conn net.Conn) (net.Conn, error) {
	h := fnv.New32()

	remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	h.Write([]byte(remoteAddr))
	hash := h.Sum(nil)

	name := fmt.Sprintf("honeytrap-%s", hex.EncodeToString(hash))

	c, ok := d.cache.Load(name)
	// c := d.cache[name]
	if !ok {
		var err error

		c, err = d.newContainer(name, d.template)
		if err != nil {
			log.Errorf("Error creating container: %s", err.Error())
			return nil, err
		}

		// m.Lock()
		// d.cache[name] = c
		//m.Unlock()
		d.cache.Store(name, c)
	}

	if err := c.(*lxcContainer).ensureStarted(); err != nil {
		log.Errorf("Error creating container: %s", err.Error())
		return nil, err
	}

	if ta, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		connection, err := c.(*lxcContainer).Dial("tcp", ta.Port)
		return lxcContainerConn{Conn: connection, container: c.(*lxcContainer)}, err
	} else if ta, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		connection, err := c.(*lxcContainer).Dial("udp", ta.Port)
		return lxcContainerConn{Conn: connection, container: c.(*lxcContainer)}, err
	} else {
		return nil, errors.New("Unsupported protocol")
	}
}

type lxcContainer struct {
	c *lxc.Container

	d     *lxcDirector
	name  string
	eb    pushers.Channel
	lxcCh chan interface{}

	idle     time.Time
	ip       net.IP
	idevice  string
	template string
	Delays   Delays
}

// NewContainer returns a new LxcContainer from the provider.
func (d *lxcDirector) newContainer(name string, template string) (*lxcContainer, error) {
	c := lxcContainer{
		name:     name,
		template: template,
		eb:       d.eb,
		d:        d,
		lxcCh:    d.lxcCh,
		Delays: Delays{
			FreezeDelay:      Delay(15 * time.Minute),
			StopDelay:        Delay(30 * time.Minute),
			HousekeeperDelay: Delay(1 * time.Minute),
		},
	}

	if c2, err := lxc.NewContainer(c.name); err == nil {
		// TODO(nl5887): beautify
		c.c = c2
		return &c, nil
	}

	if err := c.clone(); err != nil {
		return nil, err
	}
	return &c, nil
}

// housekeeper handles the needed process of handling internal logic
// in maintaining the provided lxc.Container.
func (c *lxcContainer) housekeeper() {
	// container lifetime function
	log.Infof("Housekeeper (%s) started.", c.name)
	defer log.Infof("Housekeeper (%s) stopped.", c.name)

	for {
		time.Sleep(time.Duration(c.Delays.HousekeeperDelay))

		if c.isStopped() {
			continue
		}

		if time.Since(c.idle) > time.Duration(c.Delays.StopDelay) && c.isFrozen() {
			log.Debugf("LxcContainer %s: idle for %s, stopping container", c.name, time.Now().Sub(c.idle).String())
			c.lxcCh <- LxcStop{c.c}
			return
		} else if time.Since(c.idle) > time.Duration(c.Delays.FreezeDelay) && c.isRunning() {
			log.Debugf("LxcContainer %s: idle for %s, freezing container", c.name, time.Now().Sub(c.idle).String())
			c.lxcCh <- LxcFreeze{c.c}
		}
	}
}

// clone attempts to clone the underline lxc.Container.
func (c *lxcContainer) clone() error {
	log.Debugf("Creating new container %s from template %s", c.name, c.d.template)

	c1, err := lxc.NewContainer(c.template)
	if err != nil {
		return err
	}

	defer lxc.Release(c1)

	// http://developerblog.redhat.com/2014/09/30/overview-storage-scalability-docker/
	// TODO(nl5887): use overlayfs / make it configurable
	if err = c1.Clone(c.name, lxc.CloneOptions{
		Backend:  lxc.Aufs,
		Snapshot: true,
		KeepName: true,
	}); err != nil {
		return err
	}

	if c.c, err = lxc.NewContainer(c.name); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.console.path", "none"); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.tty.max", "0"); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.cgroup.devices.deny", "c 5:1 rwm"); err != nil {
		return err
	}

	c.d.eb.Send(ContainerClonedEvent(c.name, c.template))

	return nil
}

// start begins the call to the lxc.Container.
func (c *lxcContainer) start() error {
	log.Infof("Starting container %s", c.c.Name())

	c.idle = time.Now()

	if !c.c.Defined() {
		if err := c.clone(); err != nil {
			log.Error(err.Error())
			return err
		}
	}

	c.d.eb.Send(ContainerStartedEvent(c.name))

	// run independent of our process
	c.c.WantDaemonize(true)

	c.lxcCh <- LxcStart{c.c}
	// Housekeeper only runs in Running containers, so start it always
	go c.housekeeper()

	if err := c.settle(); err != nil {
		return err
	}

	/*
		if err := c.sf.Start(c.idevice); err != nil {
			log.Errorf("Error occurred while attaching sniffer for %s to %s: %s", c.name, c.idevice, err.Error())
		}
	*/

	return nil
}

// unfreeze sets the internal container into an unfrozen state.
func (c *lxcContainer) unfreeze() error {
	log.Infof("Unfreezing container: %s", c.name)

	c.lxcCh <- LxcUnfreeze{c.c}

	if err := c.settle(); err != nil {
		return err
	}

	c.d.eb.Send(ContainerUnfrozenEvent(c.name, c.ip))

	/*
		if err := c.sf.Start(c.idevice); err != nil {
			log.Errorf("Error occurred while attaching sniffer for %s to %s: %s", c.name, c.idevice, err.Error())
		}
	*/

	return nil
}

// settle runs the process to take the container into a proper running state.
func (c *lxcContainer) settle() error {
	log.Infof("Waiting for container %s to settle, current state=%s", c.name, c.c.State())

	if !c.c.Wait(lxc.RUNNING, 30*time.Second) {
		return fmt.Errorf("lxccontainer still not running %s", c.name)
	}

	retries := 0

	for {
		ip, err := c.c.IPAddress("eth0")
		if err == nil {
			log.Debugf("Got ip: %s", ip[0])
			c.ip = net.ParseIP(ip[0])
			break
		}

		if retries < 50 {
			log.Debugf("Waiting for ip to settle %s (%s)", c.name, err.Error())
			time.Sleep(time.Millisecond * 200)
			retries++
			continue
		}

		return fmt.Errorf("Could not get an IP address")
	}

	var isets []string
	netws := c.c.ConfigItem("lxc.net")
	for ind := range netws {
		itypes := c.c.RunningConfigItem(fmt.Sprintf("lxc.net.0.%d.type", ind))
		if itypes == nil {
			continue
		}

		if itypes[0] == "veth" {
			isets = c.c.RunningConfigItem(fmt.Sprintf("lxc.net.0.%d.veth.pair", ind))
			break
		} else {
			isets = c.c.RunningConfigItem(fmt.Sprintf("lxc.net.0.%d.link", ind))
			break
		}
	}

	if len(isets) == 0 {
		return fmt.Errorf("could not get an network device")
	}

	c.idevice = isets[0]

	log.Debugf("Using network device %s to %s", c.idevice, c.name)

	c.idle = time.Now()

	return nil
}

func (c *lxcContainer) ensureStarted() error {
	if c.isFrozen() {
		return c.unfreeze()
	}

	if c.isStopped() {
		return c.start()
	}

	// settle will fill the container with ip address and interface
	return c.settle()
}

// isRunning returns true/false if the container is in running state.
func (c *lxcContainer) isRunning() bool {
	return c.c.State() == lxc.RUNNING
}

// isStopped returns true/false if the container is in stopped state.
func (c *lxcContainer) isStopped() bool {
	return c.c.State() == lxc.STOPPED
}

// isFrozen returns true/false if the container is in frozen state.
func (c *lxcContainer) isFrozen() bool {
	return c.c.State() == lxc.FROZEN
}

// Dial attempts to connect to the internal network of the
// internal container.
func (c *lxcContainer) Dial(network string, port int) (net.Conn, error) {
	host := net.JoinHostPort(c.ip.String(), fmt.Sprintf("%d", port))
	retries := 0
	for {
		conn, err := net.Dial(network, host)
		if err == nil {
			return conn, nil
		}

		if retries < 50 {
			log.Debug("Waiting for container to be fully started %s (%s)\n", c.name, err.Error())
			time.Sleep(time.Millisecond * 200)
			retries++
			continue
		}

		return nil, fmt.Errorf("could not connect to container")
	}
}
