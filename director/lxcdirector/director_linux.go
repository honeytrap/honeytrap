// +build linux

package lxcdirector

import (
	// #nosec
	"net"
	"sync"

	config "github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"

	lxc "github.com/honeytrap/golxc"
	"github.com/honeytrap/namecon"
)

// Director defines a struct which handles the management of registered containers.
type Director struct {
	containers map[string]director.Container
	m          sync.Mutex
	config     *config.Config
	provider   *LxcProvider
	namer      namecon.Namer
}

// New returns a new instance of a Director.
func New(conf *config.Config, events pushers.Events) *Director {
	// TODO: Need to replace this with Event API.
	// pusher := pushers.NewRecordPusher(conf)

	d := &Director{
		config:     conf,
		containers: map[string]director.Container{},
		provider:   NewLxcProvider(conf, events),
		namer:      namecon.NewNamerCon(conf.Template+"-%s", namecon.Basic{}),
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

func (d *Director) getName(c net.Conn) (string, error) {
	rhost, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return "", err
	}

	return d.namer.New(rhost), nil
}

// NewContainer returns a new providers.Container instance from the provided internal providers.
func (d *Director) NewContainer(name string) (director.Container, error) {
	return d.provider.NewContainer(
		name,
	)
}

// GetContainer returns a provider.Container instance from those already created on the director.
func (d *Director) GetContainer(c net.Conn) (director.Container, error) {
	d.m.Lock()
	defer d.m.Unlock()

	name, err := d.getName(c)
	if err != nil {
		return nil, err
	}

	log.Infof("Using container %s for addr: %s", name, c.RemoteAddr().String())

	if container, ok := d.containers[name]; ok {
		return container, nil
	}

	// TODO: ContainerConfig?
	container, err := d.NewContainer(name)
	d.containers[name] = container
	return container, err
}
