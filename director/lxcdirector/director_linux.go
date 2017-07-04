// +build linux

package lxcdirector

import (
	// #nosec
	"errors"
	"net"
	"sync"

	"github.com/BurntSushi/toml"
	config "github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"

	lxc "github.com/honeytrap/golxc"
	"github.com/honeytrap/namecon"
)

const (
	// DirectorKey defines the key used to choose this giving director.
	DirectorKey = "lxc"
)

var (
	_ = director.RegisterDirector("lxc", NewWith)
)

// Director defines a struct which handles the management of registered containers.
type Director struct {
	lxconfig   LxcConfig
	config     *config.Config
	provider   *LxcProvider
	namer      namecon.Namer
	m          sync.Mutex
	containers map[string]director.Container
}

// NewWith defines a function to return a director.Director.
func NewWith(cnf *config.Config, meta toml.MetaData, data toml.Primitive, events pushers.Channel) (director.Director, error) {
	var lconfig LxcConfig

	if err := meta.PrimitiveDecode(data, &lconfig); err != nil {
		return nil, err
	}

	return New(cnf, lconfig, events), nil
}

// New returns a new instance of the Director.
func New(config *config.Config, xconfig LxcConfig, events pushers.Channel) *Director {
	// TODO: Need to replace this with Event API.
	// pusher := pushers.NewRecordPusher(conf)

	d := &Director{
		config:     config,
		lxconfig:   xconfig,
		provider:   NewLxcProvider(config, xconfig, events),
		containers: map[string]director.Container{},
		namer:      namecon.NewNamerCon(config.Template+"-%s", namecon.Basic{}),
	}

	d.registerContainers()

	// TODO: do we need this pusher?, use default pushers, PushData or something
	// go func() {
	// 	if err := pusher.Run(); err != nil {
	// 		log.Errorf("Error during Run call for pusher: %s", err.Error())
	// 	}
	// }()

	return d
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

func (d *Director) registerContainers() {
	// TODO: make this lxc independent
	for _, c := range lxc.Containers() {
		if c.State() == lxc.STOPPED {
			continue
		}

		name := c.Name()

		container, err := d.NewContainer(name)
		if err != nil {
			log.Errorf("Error during container registration: %s", err.Error())
			continue
		}

		d.containers[name] = container
	}
}

func (d *Director) getName(addr string) (string, error) {
	rhost, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}

	return d.namer.New(rhost), nil
}

// NewContainer returns a new providers.Container instance from the provided internal providers.
func (d *Director) NewContainer(addr string) (director.Container, error) {
	name, err := d.getName(addr)
	if err != nil {
		return nil, err
	}

	if container, ok := d.containers[name]; ok {
		return container, nil
	}

	log.Infof("Add new container %s for addr: %s", name, addr)

	// TODO: ContainerConfig?
	dl, err := d.provider.NewContainer(name)
	if err != nil {
		return nil, err
	}

	d.containers[name] = dl
	return dl, nil
}

// GetContainer returns a provider.Container instance from those already created on the director.
func (d *Director) GetContainer(c net.Conn) (director.Container, error) {
	d.m.Lock()
	defer d.m.Unlock()

	name, err := d.getName(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	log.Infof("Using container %s for addr: %s", name, c.RemoteAddr().String())

	if container, ok := d.containers[name]; ok {
		return container, nil
	}

	return nil, errors.New("Container not found")
}
