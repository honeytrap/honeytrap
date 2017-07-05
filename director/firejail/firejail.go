package firejail

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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

	fireJailScript = `#!/bin/sh
firejail %s`
)

var (
	dailTimeout = 5 * time.Second
	log         = logging.MustGetLogger("honeytrap:director:firejail")
	_           = director.RegisterDirector("firejail", NewWith)
)

// JailConfig defines a structure for the execution of a command policy for the generation
// of a given firejail instance.
type JailConfig struct {
	Options       map[string]string       `toml:"options"`
	Envs          map[string]string       `toml:"envs"`
	App           string                  `toml:"app"`
	Name          string                  `toml:"name"`
	DefaultPort   string                  `toml:"default_port"`
	Profile       string                  `toml:"profile"`
	IgnoreProfile bool                    `toml:"ignore_profile"`
	IPAddr        string                  `toml:"ip_addr"`
	Net           string                  `toml:"net"`
	Commands      []process.Command       `toml:"commands"`
	Scripts       []process.ScriptProcess `toml:"scripts"`
}

// Director defines a central structure which creates/retrieves Container
// connections for the giving system.
type Director struct {
	config         *config.Config
	jailConfig     JailConfig
	namer          namecon.Namer
	events         pushers.Channel
	globalCommands process.SyncProcess
	globalScripts  process.SyncScripts
	m              sync.Mutex
	containers     map[string]director.Container
}

// NewWith defines a function to return a director.Director.
func NewWith(cnf *config.Config, meta toml.MetaData, data toml.Primitive, events pushers.Channel) (director.Director, error) {
	var jconfig JailConfig

	if err := meta.PrimitiveDecode(data, &jconfig); err != nil {
		return nil, err
	}

	director := New(cnf, jconfig, events)

	return director, nil
}

// New returns a new instance of the Director.
func New(config *config.Config, jailconfig JailConfig, events pushers.Channel) *Director {
	return &Director{
		config:     config,
		jailConfig: jailconfig,
		events:     events,
		containers: make(map[string]director.Container),
		namer:      namecon.NewNamerCon("firejail-%s", namecon.Basic{}),
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
		targetName:   name,
		config:       d.config,
		meta:         d.jailConfig,
		gscripts:     d.globalScripts,
		gcommands:    d.globalCommands,
		targetScript: fmt.Sprintf("%s-firejail.sh", name),
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
		log.Errorf("Jail : Failed to retrieve existing container : %+q : %+q", conn.RemoteAddr(), err)
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

	return d.NewContainer(conn.RemoteAddr().String())
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
	targetName   string
	targetScript string
	config       *config.Config
	gcommands    process.SyncProcess
	gscripts     process.SyncScripts
	meta         JailConfig
}

// Detail returns the ContainerDetail related to this giving container.
func (io *JailContainer) Detail() director.ContainerDetail {
	return director.ContainerDetail{
		Name:          io.targetName,
		ContainerAddr: fmt.Sprintf("%s:%s", io.meta.IPAddr, io.meta.DefaultPort),
		Meta: map[string]interface{}{
			"driver": DirectorKey,
		},
	}
}

// writeScript will run the giving file script with the contents of the firejail script.
func (io *JailContainer) writeScript() error {
	args, err := toArgs(io.meta)
	if err != nil {
		log.Error("Jail : %q : Failed to write script : %q", err, io.targetScript)
		return err
	}

	file, err := os.Create(io.targetScript)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := fmt.Fprintf(file, fireJailScript, strings.Join(args, " ")); err != nil {
		return err
	}

	return nil
}

// Dial connects to the giving address to provide proxying stream between
// both endpoints.
func (io *JailContainer) Dial(ctx context.Context, port string) (net.Conn, error) {
	log.Infof("Jail : %q : Dial Connection : Remote : %q", io.targetName, io.meta.App)

	if port == "0" {
		port = io.meta.DefaultPort
	}

	// Attempt to write a script file for execution.
	if err := io.writeScript(); err != nil {
		log.Errorf("Jail : %q : Dial Connection : Failed : %q", io.targetName, err)
		return nil, err
	}

	var command process.Command
	command.Name = "/bin/sh"
	command.Level = process.RedAlert
	command.Args = []string{io.targetScript}

	log.Infof("Jail : %q : Dial Connection : Executing Command : Command{Name: %q, Args: %+q}", io.targetName, command.Name, command.Args)

	// Run command associated with firejail to bootup
	if err := command.Run(ctx, os.Stdout, os.Stderr); err != nil {
		log.Errorf("Jail : %q : Dial Connection : Failed : %q", io.targetName, err)
		return nil, err
	}

	// Execute all local commands.
	localScripts := process.SyncScripts{Scripts: io.meta.Scripts}
	localCommands := process.SyncProcess{Commands: io.meta.Commands}

	if err := localCommands.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		log.Errorf("Jail : %q : Dial Connection : Failed : %q", io.targetName, err)
		return nil, err
	}

	if err := localScripts.Exec(ctx, os.Stdout, os.Stderr); err != nil {
		log.Errorf("Jail : %q : Dial Connection : Failed : %q", io.targetName, err)
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", io.meta.IPAddr, port), dailTimeout)
	if err != nil {
		log.Errorf("Jail : %q : Dial Connection : Failed : %q", io.targetName, err)
		return nil, err
	}

	return conn, nil
}

// Name returns the target address for this specific container.
func (io *JailContainer) Name() string {
	return io.targetName
}

//===================================================================================================================

func toArgs(jc JailConfig) ([]string, error) {
	if jc.Name == "" {
		return nil, errors.New("Name can not be empty in JailConfig")
	}

	if jc.App == "" {
		return nil, errors.New("App can not be empty in JailConfig")
	}

	var args []string

	if jc.Name != "" {
		// args = append(args, "name", jc.Name)
		args = append(args, fmt.Sprintf("--name=%s", jc.Name))
	}

	if !jc.IgnoreProfile {
		_, ok := jc.Options["profile"]
		if jc.Profile == "" && !ok {
			args = append(args, fmt.Sprintf("--profile=%s", "noprofile"))
		} else {
			if jc.Profile != "" && !ok {
				args = append(args, fmt.Sprintf("--profile=%s", jc.Profile))
			}
		}
	}

	_, ok := jc.Options["net"]
	if jc.Net != "" && !ok {
		// args = append(args, "net", jc.Net)
		args = append(args, fmt.Sprintf("--net=%s", jc.Net))
	}

	if jc.IPAddr != "" {
		args = append(args, fmt.Sprintf("--ip=%q", jc.IPAddr))
	} else if ip, ok := jc.Options["ip"]; ok {
		args = append(args, fmt.Sprintf("--ip=%q", ip))
	} else {
		addr := director.GetHostAddr("")

		if ip, _, err := net.SplitHostPort(addr); err == nil {
			args = append(args, fmt.Sprintf("--ip=%q", ip))
		}
	}

	for name, value := range jc.Envs {
		// args = append(args, "env", fmt.Sprintf("%s=%s", name, value))
		args = append(args, fmt.Sprintf("--env %s=%s", name, value))
	}

	for name, value := range jc.Options {
		// args = append(args, name, value)
		args = append(args, fmt.Sprintf("--%s=%s", name, value))
	}

	// Add the appname.
	args = append(args, jc.App)

	return args, nil
}
