// Package iodirector creates a director which simply generates new Containers that
// creates net.Conn connections which allows you to proxy data between these two
// endpoints.
package iodirector

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
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
)

// Director defines a central structure which creates/retrieves Container
// connections for the giving system.
type Director struct {
	config     *config.Config
	namer      namecon.Namer
	events     pushers.Events
	m          sync.Mutex
	containers map[string]director.Container
}

// New returns a new instance of the Director.
func New(config *config.Config, events pushers.Events) *Director {
	return &Director{
		config:     config,
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
		meta:       d.config.Directors.IOConfig,
		targetName: name,
	}

	d.m.Lock()
	{
		d.containers[name] = container
	}
	d.m.Unlock()

	return container, nil
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
	meta       config.IOConfig
	targetName string
}

// Dial connects to the giving address to provide proxying stream between
// both endpoints.
func (io *IOContainer) Dial() (net.Conn, error) {
	log.Infof("IO : %q : Dial Connection : Remote : %+q", io.targetName, io.meta.ServiceAddr)

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
