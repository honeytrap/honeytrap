// +build linux

package providers

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/sniffer"

	lxc "github.com/honeytrap/golxc"
)

/*
TODO: enable providers registration
var (
	_ = director.RegisterProvider("lxc", NewLxcContainer)
)*/

type LxcConfig struct {
	Template string
}

type LxcProvider struct {
	config *config.Config
	pusher *pushers.RecordPush
}

func NewLxcProvider(pusher *pushers.RecordPush, config *config.Config) Provider {
	return &LxcProvider{config, pusher}
}

type LxcContainer struct {
	ip       string
	name     string
	template string
	idevice  string
	config   *config.Config
	idle     time.Time
	c        *lxc.Container
	m        sync.Mutex
	sf       *sniffer.Sniffer
	provider *LxcProvider
}

func (lp *LxcProvider) NewContainer(name string) (Container, error) {
	c := LxcContainer{
		config:   lp.config,
		name:     name,
		provider: lp,
		idle:     time.Now(),
		sf:       sniffer.New(lp.config.NetFilter),
	}

	var err error
	if c.c, err = lxc.NewContainer(c.name); err != nil {
		return nil, err
	}

	go c.housekeeper()

	return &c, nil
}

func (c *LxcContainer) clone() error {
	log.Debugf("Creating new container %s from template %s", c.name, c.config.Template)

	var c1 *lxc.Container
	var err error
	if c1, err = lxc.NewContainer(c.config.Template); err != nil {
		return err
	}

	defer lxc.Release(c1)

	// http://developerblog.redhat.com/2014/09/30/overview-storage-scalability-docker/
	// TODO: use overlayfs / make it configurable
	if err = c1.Clone(c.name, lxc.CloneOptions{Backend: lxc.Aufs, Snapshot: true, KeepName: true}); err != nil {
		return err
	}

	if c.c, err = lxc.NewContainer(c.name); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.console", "none"); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.tty", "0"); err != nil {
		return err
	}

	if err := c.c.SetConfigItem("lxc.cgroup.devices.deny", "c 5:1 rwm"); err != nil {
		return err
	}

	return nil
}

func (c *LxcContainer) start() error {
	log.Infof("Starting container")

	c.idle = time.Now()

	if !c.c.Defined() {
		if err := c.clone(); err != nil {
			log.Error(err.Error())
			return err
		}
	}

	// run independent of our process
	c.c.WantDaemonize(true)

	if err := c.c.Start(); err != nil {
		return err
	}

	if err := c.settle(); err != nil {
		return err
	}

	if err := c.sf.Start(c.idevice); err != nil {
		log.Errorf("Error occured while attaching sniffer for %s to %s ", c.name, c.idevice, err)
	}

	return nil
}

func (c *LxcContainer) housekeeper() {
	// container lifetime function
	log.Infof("Housekeeper (%s) started.", c.name)

	for {
		time.Sleep(time.Duration(c.config.Delays.HousekeeperDelay))

		if c.isStopped() {
			continue
		}

		log.Debugf("LxcContainer %s: idle for %s with current state %s", c.name, time.Now().Sub(c.idle).String(), c.c.State().String())

		if time.Since(c.idle) > time.Duration(c.config.Delays.StopDelay) && c.isFrozen() {
			// stop
			c.stop()
		} else if time.Since(c.idle) > time.Duration(c.config.Delays.FreezeDelay) && c.isRunning() {
			// freeze
			c.freeze()
		}
	}

	log.Infof("Housekeeper (%s) stopped.", c.name)
}

func (c *LxcContainer) isRunning() bool {
	return c.c.State() == lxc.RUNNING
}

func (c *LxcContainer) isStopped() bool {
	return c.c.State() == lxc.STOPPED
}

func (c *LxcContainer) isFrozen() bool {
	return c.c.State() == lxc.FROZEN
}

func (c *LxcContainer) unfreeze() error {
	log.Infof("Unfreezing container: %s", c.name)

	if err := c.c.Unfreeze(); err != nil {
		return err
	}

	if err := c.settle(); err != nil {
		return err
	}

	if err := c.sf.Start(c.idevice); err != nil {
		log.Errorf("Error occured while attaching sniffer for %s to %s ", c.name, c.idevice, err)
	}

	return nil
}

func (c *LxcContainer) settle() error {
	log.Infof("Waiting for container to settle %s\n", c.name)

	if !c.c.Wait(lxc.RUNNING, 30) {
		return fmt.Errorf("LxcContainer still not running %s\n", c.name)
	}

	var retries int = 0
	for {
		ip, err := c.c.IPAddress("eth0")
		if err == nil {
			log.Debugf("Got ip: %s", ip)
			c.ip = ip[0]
			break
		}

		if retries < 50 {
			log.Debugf("Waiting for ip to settle %s (%s)\n", c.name, err.Error())
			time.Sleep(time.Millisecond * 200)
			retries++
			continue
		}

		return fmt.Errorf("Could not get an IP address.")
	}

	var isets []string
	netws := c.c.ConfigItem("lxc.network")
	for ind := range netws {
		itypes := c.c.RunningConfigItem(fmt.Sprintf("lxc.network.%d.type", ind))
		if itypes == nil {
			continue
		}

		if itypes[0] == "veth" {
			isets = c.c.RunningConfigItem(fmt.Sprintf("lxc.network.%d.veth.pair", ind))
			break
		} else {
			isets = c.c.RunningConfigItem(fmt.Sprintf("lxc.network.%d.link", ind))
			break
		}
	}

	if len(isets) == 0 {
		return fmt.Errorf("Could not get an network device.")
	}

	c.idevice = isets[0]

	log.Debugf("Using network device %s to %s", c.idevice, c.name)

	return nil
}

func (c *LxcContainer) ensureStarted() error {
	if c.isFrozen() {
		return c.unfreeze()
	}

	if c.isStopped() {
		return c.start()
	}

	return nil
}

func (c *LxcContainer) Device() (string, error) {
	if c.idevice == "" {
		return "", fmt.Errorf("Unable to get device")
	}
	return c.idevice, nil
}

func (c *LxcContainer) Name() string {
	return (c.name)
}

type lxcContainerConn struct {
	net.Conn
	container *LxcContainer
}

func (c lxcContainerConn) Read(b []byte) (n int, err error) {
	c.container.stillActive()
	return c.Conn.Read(b)
}

func (c lxcContainerConn) Write(b []byte) (n int, err error) {
	c.container.stillActive()
	return c.Conn.Write(b)
}

func (c *LxcContainer) stillActive() error {
	if err := c.ensureStarted(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.idle = time.Now()
	return nil
}

func (c *LxcContainer) deltaUp() (string, error) {
	fo, err := ioutil.TempFile("", c.name)
	if err != nil {
		return "", err
	}

	defer fo.Close()

	w := gzip.NewWriter(fo)
	defer w.Close()

	rootfs := c.c.ConfigItem("lxc.rootfs")[0]
	deltaPath := strings.Split(rootfs, ":")[2]

	err = TarWalker(deltaPath, w)
	if err != nil {
		return "", err
	}

	return fo.Name(), nil
}

func (c *LxcContainer) checkpoint() (string, error) {
	tmpdir, err := ioutil.TempDir("", c.name)
	if err != nil {
		return "", err
	}

	defer func() {
		os.RemoveAll(tmpdir)
	}()

	cx := lxc.CheckpointOptions{
		Directory: tmpdir,
		Stop:      false,
		Verbose:   true,
	}

	err = c.c.Checkpoint(cx)
	if err != nil {
		return "", err
	}

	fo, err := ioutil.TempFile("", c.name)
	if err != nil {
		return "", err
	}
	defer fo.Close()

	w := gzip.NewWriter(fo)
	defer w.Close()

	err = TarWalker(tmpdir, w)
	if err != nil {
		return "", err
	}

	return fo.Name(), nil
}

func (c *LxcContainer) freeze() error {
	if !c.isRunning() {
		// not running
		return nil
	}

	// should actually first checkpoint, stop sniffer and freeze, then tar

	for {
		log.Debug("Checkpointing container: %s", c.name)

		chpnt, err := c.checkpoint()
		if err != nil {
			log.Error("Checkpoint failed: %s", err.Error())
			break
		}

		chp, err := NewFileCloser(chpnt)
		if err != nil {
			log.Error("Unable to find checkpoint file closer: %s", err)
			break
		}

		defer func() {
			if err := os.Remove(chp.Name()); err != nil {
				log.Error("Error deleting file (%s):", chp.Name(), err.Error())
			}
		}()

		buff, err := ioutil.ReadAll(chp)
		if err != nil {
			log.Error("Could not read checkpoint: %s", err)
			break
		}

		endpoint := fmt.Sprintf("http://api.honeytrap.io/v1/container/%s/checkpoint", c.name)
		c.provider.pusher.Push(endpoint, buff)
		break
	}

	log.Debugf("Freezing container: %s", c.name)
	if err := c.c.Freeze(); err != nil {
		return err
	}

	for {
		log.Debugf("Sending packets.")
		endpoint := fmt.Sprintf("http://api.honeytrap.io/v1/container/%s/packets", c.name)

		r, err := c.sf.Stop()
		if err != nil {
			log.Errorf("Could not read packets", err)
			break
		}

		buff, err := ioutil.ReadAll(r)
		if err != nil {
			log.Errorf("Could not read packets", err)
			break
		}

		log.Debugf("Pushing packets")
		c.provider.pusher.Push(endpoint, buff)
		break
	}

	for {
		log.Debugf("Tarring container: %s", c.name)
		delta, err := c.deltaUp()
		if err != nil {
			log.Error("Could not tar: %s", err)
			break
		}

		r, err := NewFileCloser(delta)
		if err != nil {
			log.Error(err.Error())
			break
		}

		defer func() {
			if err := os.Remove(r.Name()); err != nil {
				log.Error("Error deleting file (%s):", r.Name(), err.Error())
			}
		}()

		buff, err := ioutil.ReadAll(r)
		if err != nil {
			log.Error(err.Error())
			break
		}

		endpoint := fmt.Sprintf("http://api.honeytrap.io/v1/container/%s/data", c.name)
		c.provider.pusher.Push(endpoint, buff)
		break
	}

	return nil
}

func (c *LxcContainer) stop() error {
	if c.isStopped() {
		// already stopped
		return nil
	}

	log.Infof("LxcContainer (%s) stopping (ip: %s)", c.name, c.ip)

	if err := c.c.Stop(); err != nil {
		return err
	}

	return nil
}

func (c *LxcContainer) CleanUp() error {
	return nil
}

func (c *LxcContainer) Dial(port string) (net.Conn, error) {
	if err := c.ensureStarted(); err != nil {
		return nil, err
	}

	if err := c.settle(); err != nil {
		return nil, err
	}

	var conn net.Conn
	var err error

	retries := 0
	for {
		// TODO(nl5887): remove
		port = "22"

		conn, err = net.Dial("tcp", net.JoinHostPort(c.ip, port))
		if err == nil {
			break
		}

		if retries < 50 {
			log.Debug("Waiting for container to be fully started %s (%s)\n", c.name, err.Error())
			time.Sleep(time.Millisecond * 200)
			retries++
			continue
		}

		return nil, fmt.Errorf("Could not connect to container.")
	}

	return lxcContainerConn{Conn: conn, container: c}, err
}
