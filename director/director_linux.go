// +build linux

package director

import (
	"crypto/md5"
	"fmt"
	"net"
	"sync"

	config "github.com/honeytrap/honeytrap/config"
	providers "github.com/honeytrap/honeytrap/providers"
	pushers "github.com/honeytrap/honeytrap/pushers"

	lxc "github.com/honeytrap/golxc"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:director")

type Director struct {
	containers map[string]providers.Container
	m          sync.Mutex
	config     *config.Config
	provider   providers.Provider
}

func New(conf *config.Config) *Director {
	pusher := pushers.NewRecordPusher(conf)

	d := &Director{
		containers: map[string]providers.Container{},
		config:     conf,
		provider:   providers.NewLxcProvider(pusher, conf),
	}

	d.registerContainers()

	// TODO: do we need this pusher?, use default pushers, PushData or something
	go pusher.Run()

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
			log.Error("Error during container registration: %s", err.Error())
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

	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprintf("%s%s", rhost, d.config.Token)))
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func (d *Director) NewContainer(name string) (providers.Container, error) {
	return d.provider.NewContainer(
		name,
	)
}

func (d *Director) GetContainer(c net.Conn) (providers.Container, error) {
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
