// Package iodirector creates a director which simply generates new Containers that
// creates net.Conn connections which allows you to proxy data between these two
// endpoints.
package iodirector

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/process"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/namecon"
	logging "github.com/op/go-logging"
)

const (
	// DirectorKey defines the key used to choose this giving director.
	DirectorKey = "io"
)

var (
	dailTimeout = 5 * time.Second
	log         = logging.MustGetLogger("honeytrap:director:io")
	_           = director.RegisterDirector("io", NewWith)
)

// IOConfig defines the settings for the iodirector.
type IOConfig struct {
	ServiceAddr string                  `toml:"service_addr"`
	Commands    []process.Command       `toml:"commands"`
	Scripts     []process.ScriptProcess `toml:"scripts"`
}

// Director defines a central structure which creates/retrieves Container
// connections for the giving system.
type Director struct {
	config     *config.Config
	ioconfig   Config
	namer      namecon.Namer
	events     pushers.Channel
	m          sync.Mutex
	containers map[string]director.Container
}

// NewWith defines a function to return a director.Director.
func NewWith(cnf *Config, meta toml.MetaData, data toml.Primitive, events pushers.Channel) (director.Director, error) {
	var jconfig IOConfig

	if err := meta.PrimitiveDecode(data, &jconfig); err != nil {
		return nil, err
	}

	return New(cnf, jconfig, events), nil
}

// New returns a new instance of the Director.
func New(config *config.Config, ioc IOConfig, events pushers.Channel) *Director {
	return &Director{
		config:     config,
		ioconfig:   ioc,
		events:     events,
		containers: make(map[string]director.Container),
		namer:      namecon.NewNamerCon(config.Template+"-%s", namecon.Basic{}),
	}
}

// NewContainer returns a new Container generated from the director with the specified address.
func (d *Director) NewContainer(addr string) (director.Container, error) {
	log.Infof("IO : Creating new container : %s", addr)

	var err error
	var container director.Container

	name, err := d.getName(addr)
	if err != nil {
		log.Errorf("IO : Failed to make new container name : %+q", err)
		return nil, err
	}

	d.m.Lock()
	{
		var ok bool
		if container, ok = d.containers[name]; ok {
			d.m.Unlock()
			return container, nil
		}
	}
	d.m.Unlock()

	container = &IOContainer{
		targetAddr: addr,
		targetName: name,
		config:     d.config,
		meta:       d.ioconfig,
	}

	d.m.Lock()
	{
		d.containers[name] = container
	}
	d.m.Unlock()

	return container, nil
}

// ListContainers returns the giving list of containers details
// for all connected containers.
func (d *Director) ListContainers() []director.ContainerDetail {
	var details []director.ContainerDetail

	for _, item := range d.containers {
		details = append(details, item.Detail())
	}

	return details
}

// GetContainer returns a new Container using the provided net.Conn if already registered.
func (d *Director) GetContainer(conn net.Conn) (director.Container, error) {
	log.Infof("IO : Attempt to retrieve existing container : %+q", conn.RemoteAddr())

	var container director.Container

	name, err := d.getName(conn.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	d.m.Lock()
	{
		var ok bool
		if container, ok = d.containers[name]; ok {
			d.m.Unlock()
			return container, nil
		}
	}
	d.m.Unlock()

	return nil, errors.New("Container not found")
}

// getName returns a new name based on the provided address.
func (d *Director) getName(addr string) (string, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}

	return d.namer.New(host), nil
}

//=================================================================================

// IOContainer defines a core container structure which generates new net connections
// between stream endpoints.
type IOContainer struct {
	targetAddr string
	targetName string
	config     *config.Config
	meta       config.IOConfig
}

// Detail returns the ContainerDetail related to this giving container.
func (io *IOContainer) Detail() director.ContainerDetail {
	return director.ContainerDetail{
		Name:          io.targetName,
		ContainerAddr: io.meta.ServiceAddr,
		Meta: map[string]interface{}{
			"driver": "io",
		},
	}
}

// Dial connects to the giving address to provide proxying stream between
// both endpoints.
func (io *IOContainer) Dial(ctx context.Context, port string) (net.Conn, error) {
	log.Infof("IO : %q : Dial Connection : Remote : %+q", io.targetName, io.meta.ServiceAddr)

	// Execute all local commands.
	localScripts := process.SyncScripts{Scripts: io.meta.Scripts}
	localCommands := process.SyncProcess{Commands: io.meta.Commands}

	if err := localCommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	if err := localScripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", io.meta.ServiceAddr, dailTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Name returns the target address for this specific container.
func (io *IOContainer) Name() string {
	return io.targetName
}
