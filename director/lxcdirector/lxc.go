// +build linux

package lxcdirector

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/process"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/sniffer"
	"github.com/honeytrap/honeytrap/utils/files"
	logging "github.com/op/go-logging"

	lxc "github.com/honeytrap/golxc"
)

var log = logging.MustGetLogger("honeytrap:lxcdirector")

/*
TODO: enable providers registration
var (
	_ = director.RegisterProvider("lxc", NewLxcContainer)
)*/

// LxcConfig defines a struct to provide configuration fields for a LxcProvider.
type LxcConfig struct {
	Template string
}

// LxcProvider defines a struct which loads the needed configuration for handling
// lxc based containers.
type LxcProvider struct {
	config         *config.Config
	events         pushers.Channel
	globalCommands process.SyncProcess
	globalScripts  process.SyncScripts
}

// NewLxcProvider returns a new instance of a LxcProvider as a Provider.
func NewLxcProvider(config *config.Config, events pushers.Channel) *LxcProvider {
	return &LxcProvider{
		config:         config,
		events:         events,
		globalScripts:  process.SyncScripts{Scripts: config.Directors.Scripts},
		globalCommands: process.SyncProcess{Commands: config.Directors.Commands},
	}
}

// NewContainer returns a new LxcContainer from the provider.
func (lp *LxcProvider) NewContainer(name string) (director.Container, error) {
	c := LxcContainer{
		provider:  lp,
		name:      name,
		config:    lp.config,
		idle:      time.Now(),
		gscripts:  lp.globalScripts,
		gcommands: lp.globalCommands,
		meta:      lp.config.Directors.LXC,
		sf:        sniffer.New(lp.config.NetFilter),
	}

	var err error
	if c.c, err = lxc.NewContainer(c.name); err != nil {
		return nil, err
	}

	go c.housekeeper()

	return &c, nil
}

// LxcContainer defines a struct to encapsulated a lxc.Container.
type LxcContainer struct {
	ip        string
	name      string
	template  string
	idevice   string
	idle      time.Time
	config    *config.Config
	meta      config.LxcConfig
	c         *lxc.Container
	m         sync.Mutex
	sf        *sniffer.Sniffer
	provider  *LxcProvider
	gcommands process.SyncProcess
	gscripts  process.SyncScripts
}

// Detail returns the ContainerDetail related to this giving container.
func (c *LxcContainer) Detail() director.ContainerDetail {
	return director.ContainerDetail{
		Name:          c.name,
		ContainerAddr: c.ip,
		Meta: map[string]interface{}{
			"template": c.template,
			"idle":     c.idle,
			"driver":   "lxc",
			"idevice":  c.idevice,
		},
	}
}

// clone attempts to clone the underline lxc.Container.
func (c *LxcContainer) clone() error {
	log.Debugf("Creating new container %s from template %s", c.name, c.config.Template)

	var c1 *lxc.Container
	var err error
	if c1, err = lxc.NewContainer(c.config.Template); err != nil {
		return err
	}

	defer lxc.Release(c1)

	c.provider.events.Send(director.ContainerClonedEvent(c, c.name, c.template, c.ip))

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

// start begins the call to the lxc.Container.
func (c *LxcContainer) start() error {
	log.Infof("Starting container")

	c.idle = time.Now()

	if !c.c.Defined() {
		if err := c.clone(); err != nil {
			log.Error(err.Error())

			return err
		}
	}

	c.provider.events.Send(director.ContainerStartedEvent(c, map[string]interface{}{
		"name":     c.name,
		"template": c.template,
		"ip":       c.ip,
	}))

	// run independent of our process
	c.c.WantDaemonize(true)

	if err := c.c.Start(); err != nil {
		return err
	}

	if err := c.settle(); err != nil {
		return err
	}

	if err := c.sf.Start(c.idevice); err != nil {
		log.Errorf("Error occured while attaching sniffer for %s to %s: %s", c.name, c.idevice, err.Error())
	}

	return nil
}

// housekeeper handls the needed process of handling internal logic
// in maintaining the provided lxc.Container.
func (c *LxcContainer) housekeeper() {
	// container lifetime function
	log.Infof("Housekeeper (%s) started.", c.name)
	defer log.Infof("Housekeeper (%s) stopped.", c.name)

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
}

// isRunning returns true/false if the container is in running state.
func (c *LxcContainer) isRunning() bool {
	return c.c.State() == lxc.RUNNING
}

// isStopped returns true/false if the container is in stopped state.
func (c *LxcContainer) isStopped() bool {
	return c.c.State() == lxc.STOPPED
}

// isFrozen returns true/false if the container is in frozen state.
func (c *LxcContainer) isFrozen() bool {
	return c.c.State() == lxc.FROZEN
}

// unfreeze sets the internal container into an unfrozen state.
func (c *LxcContainer) unfreeze() error {
	log.Infof("Unfreezing container: %s", c.name)

	if err := c.c.Unfreeze(); err != nil {
		return err
	}

	c.provider.events.Send(director.ContainerUnfrozenEvent(c, map[string]interface{}{
		"name":     c.name,
		"template": c.template,
		"ip":       c.ip,
	}))

	if err := c.settle(); err != nil {
		return err
	}

	if err := c.sf.Start(c.idevice); err != nil {
		log.Errorf("Error occured while attaching sniffer for %s to %s: %s", c.name, c.idevice, err.Error())
	}

	return nil
}

// settle runs the process to take the container into a proper running state.
func (c *LxcContainer) settle() error {
	log.Infof("Waiting for container to settle %s", c.name)

	if !c.c.Wait(lxc.RUNNING, 30) {
		return fmt.Errorf("lxccontainer still not running %s", c.name)
	}

	retries := 0

	for {
		ip, err := c.c.IPAddress("eth0")
		if err == nil {
			log.Debugf("Got ip: %s", ip)
			c.ip = ip[0]
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
		return fmt.Errorf("could not get an network device")
	}

	c.idevice = isets[0]

	log.Debugf("Using network device %s to %s", c.idevice, c.name)

	return nil
}

// ensureStated validates that the internal container has started.
func (c *LxcContainer) ensureStarted() error {
	if c.isFrozen() {
		return c.unfreeze()
	}

	if c.isStopped() {
		return c.start()
	}

	return nil
}

// Device returns the network device connected to the container.
func (c *LxcContainer) Device() (string, error) {
	if c.idevice == "" {
		return "", fmt.Errorf("Unable to get device")
	}
	return c.idevice, nil
}

// Name returns the underling Name of the container.
func (c *LxcContainer) Name() string {
	return (c.name)
}

// lxcContainerConn defines a custom connection type which proxies the data
// for the container.
type lxcContainerConn struct {
	net.Conn
	container *LxcContainer
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
func (c *LxcContainer) stillActive() error {
	if err := c.ensureStarted(); err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.idle = time.Now()
	return nil
}

// deltaUp retrieves the underline delta changes for the containers
// filesystem and writes it up to the underline host system.
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

	err = files.TarWalker(deltaPath, w)
	if err != nil {
		return "", err
	}

	return fo.Name(), nil
}

// checkpoint retrieves the current state of the container and
// compresses it through the checkpoint API exposed by the lxc.Container.
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

	err = files.TarWalker(tmpdir, w)
	if err != nil {
		return "", err
	}

	return fo.Name(), nil
}

// freeze sets the container into a freeze state.
func (c *LxcContainer) freeze() error {
	if !c.isRunning() {
		// not running
		return nil
	}

	c.provider.events.Send(director.ContainerFrozenEvent(c, map[string]interface{}{
		"name":     c.name,
		"template": c.template,
		"ip":       c.ip,
	}))

	// should actually first checkpoint, stop sniffer and freeze, then tar
	for {
		log.Debugf("Checkpointing container: %s", c.name)
		chpnt, err := c.checkpoint()
		if err != nil {
			log.Errorf("Checkpoint failed: %s", err.Error())
			break
		}

		chp, err := files.NewFileCloser(chpnt)
		if err != nil {
			log.Errorf("Unable to find checkpoint file closer: %s", err)
			break
		}

		defer func() {
			if cerr := os.Remove(chp.Name()); err != nil {
				log.Errorf("Error deleting file (%s): %s", chp.Name(), cerr.Error())
			}
		}()

		buff, err := ioutil.ReadAll(chp)
		if err != nil {
			log.Errorf("Could not read checkpoint: %s", err)
			break
		}

		endpoint := fmt.Sprintf("http://api.honeytrap.io/v1/container/%s/checkpoint", c.name)

		c.provider.events.Send(director.ContainerCheckpointEvent(c, buff, map[string]interface{}{
			"name":     c.name,
			"template": c.template,
			"ip":       c.ip,
			"endpoint": endpoint,
		}))

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
			log.Errorf("Could not read packets: %s", err)
			break
		}

		buff, err := ioutil.ReadAll(r)
		if err != nil {
			log.Errorf("Could not read packets: %s", err)
			break
		}

		log.Debugf("Pushing packets")

		c.provider.events.Send(director.ContainerPcappedEvent(c, buff, map[string]interface{}{
			"name":     c.name,
			"template": c.template,
			"ip":       c.ip,
			"endpoint": endpoint,
		}))
		break
	}

	for {
		log.Debugf("Tarring container: %s", c.name)
		delta, err := c.deltaUp()
		if err != nil {
			log.Errorf("Could not tar: %s", err)
			break
		}

		r, err := files.NewFileCloser(delta)
		if err != nil {
			log.Error(err.Error())
			break
		}

		defer func() {
			if cerr := os.Remove(r.Name()); cerr != nil {
				log.Errorf("Error deleting file (%s): %s", r.Name(), cerr.Error())
			}
		}()

		buff, err := ioutil.ReadAll(r)
		if err != nil {
			log.Errorf("Error reading file: %s", err.Error())
			break
		}

		endpoint := fmt.Sprintf("http://api.honeytrap.io/v1/container/%s/data", c.name)

		c.provider.events.Send(director.ContainerTarredEvent(c, buff, map[string]interface{}{
			"name":     c.name,
			"template": c.template,
			"ip":       c.ip,
			"endpoint": endpoint,
		}))
		break
	}

	return nil
}

// stop stops the container and shuts it down.
func (c *LxcContainer) stop() error {
	if c.isStopped() {
		// already stopped
		return nil
	}

	log.Infof("LxcContainer (%s) stopping (ip: %s)", c.name, c.ip)

	if err := c.c.Stop(); err != nil {
		return err
	}

	c.provider.events.Send(director.ContainerStoppedEvent(c, map[string]interface{}{
		"name":     c.name,
		"template": c.template,
		"ip":       c.ip,
	}))

	return nil
}

// CleanUp attempts to run certain process to cleanup
// the state of the uinternal container.
func (c *LxcContainer) CleanUp() error {
	return nil
}

// Dial attempts to connect to the internal network of the
// internal container.
func (c *LxcContainer) Dial(ctx context.Context, port string) (net.Conn, error) {
	if err := c.ensureStarted(); err != nil {
		return nil, err
	}

	if err := c.settle(); err != nil {
		return nil, err
	}

	// Execute all global commands.
	// TODO: Move context to be supplied by caller and not set in code
	if err := c.gcommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	if err := c.gscripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	// Execute all local commands.
	localScripts := process.SyncScripts{Scripts: c.meta.Scripts}
	localCommands := process.SyncProcess{Commands: c.meta.Commands}

	if err := localCommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	if err := localScripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	var conn net.Conn
	var err error

	retries := 0
	for {
		// TODO(nl5887): remove
		port := "22"

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

		return nil, fmt.Errorf("could not connect to container")
	}

	return lxcContainerConn{Conn: conn, container: c}, err
}
