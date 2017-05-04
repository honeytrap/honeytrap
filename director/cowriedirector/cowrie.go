// Package cowriedirector creates a director which simply generates new Containers that
// creates net.Conn connectCowriens which allows you to proxy data between these two
// endpoints.
package cowriedirector

import (
	"errors"
	"fmt"
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
	DirectorKey = "cowrie"
)

var (
	dailTimeout = 5 * time.Second
	log         = logging.MustGetLogger("honeytrap:director:cowrie")
)

// Director defines a central structure which creates/retrieves Container
// connectCowriens for the giving system.
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
	log.Infof("Cowrie : Creating new container : %s", addr)

	var err error
	var container director.Container

	name, err := d.getName(addr)
	if err != nil {
		log.Errorf("Cowrie : Failed to make new container name : %+q", err)
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

	container = &CowrieContainer{
		meta:       d.config.Directors.Cowrie,
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
	log.Infof("Cowrie : Attempt to retrieve existing container : %+q", conn.RemoteAddr())

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

// CowrieContainer defines a core container structure which generates new net connectCowriens
// between stream endpoints.
type CowrieContainer struct {
	meta       config.CowrieConfig
	targetName string
}

// Dial connects to the giving address to provide proxying stream between
// both endpoints.
func (c *CowrieContainer) Dial() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%s", c.meta.SSHAddr, c.meta.SSHPort)

	log.Infof("Cowrie : %q : Dial Connection : Remote : %+q", c.targetName, addr)

	// TODO(alex): Do we need to do more here?
	// We know we are dealing with ssh connections:
	// Do we need some checks or conditions which must be met first before
	// attempting to connect?
	conn, err := net.DialTimeout("tcp", addr, dailTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Name returns the target address for this specific container.
func (c *CowrieContainer) Name() string {
	return c.targetName
}
