package firejail

import (
	"context"
	"errors"
	"fmt"
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
	DirectorKey = "firejail"
)

var (
	dailTimeout = 5 * time.Second
	log         = logging.MustGetLogger("honeytrap:director:firejail")
	_           = director.RegisterDirector("firejail", NewWith)
)

// JailConfig defines a structure for the execution of a command policy for the generation
// of a given firejail instance.
type JailConfig struct {
	Options      map[string]string       `toml:"options"`
	Envs         map[string]string       `toml:"envs"`
	Namespace    string                  `toml:"namespace"`
	App          string                  `toml:"app"`
	Profile      string                  `toml:"profile"`
	GatewayAddr  string                  `toml:"gateway_addr"`
	IPAddr       string                  `toml:"ip_addr"`
	DNSAddr      string                  `toml:"dns_addr"`
	Hostname     string                  `toml:"hostname"`
	NetInterface string                  `toml:"net"`
	Commands     []process.Command       `toml:"commands"`
	Scripts      []process.ScriptProcess `toml:"scripts"`
}

// Director defines a central structure which creates/retrieves Container
// connections for the giving system.
type Director struct {
	config         *config.Config
	namer          namecon.Namer
	events         pushers.Channel
	globalCommands process.SyncProcess
	globalScripts  process.SyncScripts
	m              sync.Mutex
	containers     map[string]director.Container
}

// NewWith defines a function to return a director.Director.
func NewWith(cnf *Config, meta toml.MetaData, data toml.Primitive, events pushers.Channel) (director.Director, error) {
	var jconfig JailConfig

	if err := meta.PrimitiveDecode(data, &jconfig); err != nil {
		return nil, err
	}

	return New(cnf, jconfig, events), nil
}

// New returns a new instance of the Director.
func New(config *config.Config, jailconfig JailConfig, events pushers.Channel) *Director {
	return &Director{
		config:         config,
		jailConfig:     jailconfig,
		events:         events,
		containers:     make(map[string]director.Container),
		globalScripts:  process.SyncScripts{Scripts: config.Directors.Scripts},
		globalCommands: process.SyncProcess{Commands: config.Directors.Commands},
		namer:          namecon.NewNamerCon(config.Template+"-%s", namecon.Basic{}),
	}
}

// NewContainer returns a new Container generated from the director with the specified address.
func (d *Director) NewContainer(addr string) (director.Container, error) {
	log.Infof("Jail : Creating new container : %s", addr)

	var err error
	var container director.Container

	name, err := d.getName(addr)
	if err != nil {
		log.Errorf("Jail : Failed to make new container name : %+q", err)
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

	container = &JailContainer{
		config:    d.config,
		gscripts:  d.globalScripts,
		gcommands: d.globalCommands,
		meta:      d.jailConfig,
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
	log.Infof("Jail : Attempt to retrieve existing container : %+q", conn.RemoteAddr())

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

// JailContainer defines a core container structure which generates new net connections
// between stream endpoints.
type JailContainer struct {
	config    *config.Config
	gcommands process.SyncProcess
	gscripts  process.SyncScripts
	meta      JailConfig
}

// Detail returns the ContainerDetail related to this giving container.
func (io *JailContainer) Detail() director.ContainerDetail {
	return director.ContainerDetail{
		Name:          io.targetName,
		ContainerAddr: io.meta.ServiceAddr,
		Meta: map[string]interface{}{
			"driver": DirectorKey,
		},
	}
}

// Dial connects to the giving address to provide proxying stream between
// both endpoints.
func (io *JailContainer) Dial(ctx context.Context, port string) (net.Conn, error) {
	log.Infof("Jail : %q : Dial Connection : Remote : %+q", io.targetName, io.meta.ServiceAddr)

	// Execute all global commands.
	// TODO: Move context to be supplied by caller and not set in code
	if err := io.gcommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	if err := io.gscripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	command, err := toCommand(io.meta)
	if err != nil {
		return nil, err
	}

	command.Async = true

	// Run command associated with firejail to bootup
	if err := command.Run(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	// Execute all local commands.
	localScripts := process.SyncScripts{Scripts: io.meta.Scripts}
	localCommands := process.SyncProcess{Commands: io.meta.Commands}

	if err := localCommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	if err := localScripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", io.meta.IPAddr, port), dailTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Name returns the target address for this specific container.
func (io *JailContainer) Name() string {
	return io.targetName
}

//===================================================================================================================

// toCommand returns the process.Command best associated with the given JailCommand
// which returns a process.Command which executes then eeded firejail call to start up
// the desired chrooted instance.
func toCommand(jc config.FireJailConfig) (process.Command, error) {
	var proc process.Command

	if jc.App == "" {
		return proc, errors.New("App can not be empty in FireJailConfig")
	}

	proc.Name = "firejail"

	var args []string

	_, ok := jc.Options["profile"]
	if jc.Profile != "" && !ok {
		args = append(args, fmt.Sprintf("profile=%s", jc.Profile))
	}

	if jc.IPAddr != "" {
		args = append(args, fmt.Sprintf("ip=%s", jc.IPAddr))
	} else if ip, ok := jc.Options["ip"]; ok {
		args = append(args, fmt.Sprintf("ip=%s", ip))
	} else {
		args = append(args, fmt.Sprintf("ip=%s", director.GetHostAddr()))
	}

	_, ok := jc.Options["dns"]
	if jc.DNSAddr != "" && !ok {
		args = append(args, fmt.Sprintf("dns=%s", jc.DNSAddr))
	}

	_, ok := jc.Options["hostname"]
	if jc.Hostname != "" && !ok {
		args = append(args, fmt.Sprintf("hostname=%s", jc.Hostname))
	}

	_, ok := jc.Options["net"]
	if jc.NetInterface != "" && !ok {
		args = append(args, fmt.Sprintf("net=%s", jc.NetInterface))
	}

	_, ok := jc.Options["defaultgw"]
	if jc.GatewayAddr != "" && !ok {
		args = append(args, fmt.Sprintf("defaultgw=%s", jc.GatewayAddr))
	}

	if jc.Name != "" {
		args = append(args, fmt.Sprintf("name=%s", jc.Name))
	}

	for name, value := range jc.Envs {
		args = append(args, fmt.Sprintf("env %s=%s", name, value))
	}

	for name, value := range jc.Options {
		args = append(args, fmt.Sprintf("%s=%s", name, value))
	}

	// Add the appname.
	args = append(args, jc.App)

	proc.Args = args
	proc.Level = process.CriticalLevel

	return proc
}
