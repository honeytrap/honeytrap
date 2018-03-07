/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/web"

	"github.com/honeytrap/honeytrap/director"
	_ "github.com/honeytrap/honeytrap/director/forward"
	_ "github.com/honeytrap/honeytrap/director/lxc"
	// _ "github.com/honeytrap/honeytrap/director/qemu"
	// Import your directors here.

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/eventbus"

	"github.com/honeytrap/honeytrap/services"
	_ "github.com/honeytrap/honeytrap/services/elasticsearch"
	_ "github.com/honeytrap/honeytrap/services/ethereum"
	_ "github.com/honeytrap/honeytrap/services/ftp"
	_ "github.com/honeytrap/honeytrap/services/ipp"
	_ "github.com/honeytrap/honeytrap/services/redis"
	_ "github.com/honeytrap/honeytrap/services/smtp"
	_ "github.com/honeytrap/honeytrap/services/ssh"
	_ "github.com/honeytrap/honeytrap/services/vnc"

	"github.com/honeytrap/honeytrap/listener"
	_ "github.com/honeytrap/honeytrap/listener/agent"
	_ "github.com/honeytrap/honeytrap/listener/canary"
	_ "github.com/honeytrap/honeytrap/listener/netstack"
	_ "github.com/honeytrap/honeytrap/listener/socket"
	_ "github.com/honeytrap/honeytrap/listener/tap"
	_ "github.com/honeytrap/honeytrap/listener/tun"

	// proxies

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/server/profiler"

	_ "github.com/honeytrap/honeytrap/pushers/console"       // Registers stdout backend.
	_ "github.com/honeytrap/honeytrap/pushers/elasticsearch" // Registers elasticsearch backend.
	_ "github.com/honeytrap/honeytrap/pushers/file"          // Registers file backend.
	_ "github.com/honeytrap/honeytrap/pushers/kafka"         // Registers kafka backend.
	_ "github.com/honeytrap/honeytrap/pushers/marija"        // Registers marija backend.
	_ "github.com/honeytrap/honeytrap/pushers/raven"         // Registers raven backend.
	_ "github.com/honeytrap/honeytrap/pushers/slack"         // Registers slack backend.
	_ "github.com/honeytrap/honeytrap/pushers/splunk"        // Registers splunk backend.

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap/server")

// Honeytrap defines a struct which coordinates the internal logic for the honeytrap
// container infrastructure.
type Honeytrap struct {
	config *config.Config

	profiler profiler.Profiler

	// TODO(nl5887): rename to bus, should we encapsulate this?
	bus *eventbus.EventBus

	director director.Director

	token string

	dataDir string

	ports map[int][]string // Maps a port to an array of service names

	services map[string]*ServiceMap // Maps a service name to a service implementation
}

// New returns a new instance of a Honeytrap struct.
// func New(conf *config.Config) *Honeytrap {
func New(options ...OptionFn) (*Honeytrap, error) {
	bus := eventbus.New()

	// Initialize all channels within the provided config.
	conf := &config.Default

	h := &Honeytrap{
		config:   conf,
		director: director.MustDummy(),
		bus:      bus,
		profiler: profiler.Dummy(),
	}

	for _, fn := range options {
		if err := fn(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (hc *Honeytrap) startAgentServer() {
	// as := proxies.NewAgentServer(hc.director, hc.pusher, hc.configig)
	// go as.ListenAndServe()
}

// EventServiceStarted will return a service started Event struct
func EventServiceStarted(service string) event.Event {
	return event.New(
		event.Category(service),
		event.ServiceSensor,
		event.ServiceStarted,
	)
}

// PrepareRun will prepare Honeytrap to run
func (hc *Honeytrap) PrepareRun() {
}

// Wraps a Servicer, adding some metadata
type ServiceMap struct {
	Service services.Servicer

	Name string
	Type string
}

/* Finds a service that can handle the given connection.
 * The service is picked (among those configured for the given port) as follows:
 *
 *     If there is only one service, pick it
 *     For each service (as sorted in the config file):
 *         - If it does not implement CanHandle, pick it
 *         - If it implements CanHandle, peek the connection and pass it to
 *           CanHandle. If it returns true, pick it
 */
func (hc *Honeytrap) findService(conn net.Conn) (*ServiceMap, net.Conn) {
	localAddr := conn.LocalAddr()
	var port int
	switch localAddr.(type) {
	case *net.TCPAddr:
		port = localAddr.(*net.TCPAddr).Port
	case *net.UDPAddr:
		port = localAddr.(*net.UDPAddr).Port
	default:
		log.Errorf("Unknown address type!")
	}
	serviceNames, ok := hc.ports[port]
	if !ok {
		// No logging, it will be reported by the caller instead
		// Todo(capacitorset): implement port "any"?
		return nil, nil
	}
	var serviceCandidates []*ServiceMap
	for _, serviceName := range serviceNames {
		// The second variable is unused, we already checked for the existence
		// of the service name when parsing the config
		serviceCandidates = append(serviceCandidates, hc.services[serviceName])
	}
	if len(serviceCandidates) == 1 {
		return serviceCandidates[0], conn
	}

	peekUninitialized := true
	var tConn net.Conn
	var pConn *peekConnection
	var n int
	buffer := make([]byte, 1024)
	for _, service := range serviceCandidates {
		ch, ok := service.Service.(services.CanHandlerer)
		if !ok {
			// Service does not implement CanHandle, assume it can handle the connection
			return service, conn
		}
		// Service implements CanHandle, initialize it if needed and run the checks
		if peekUninitialized {
			// wrap connection in a connection with deadlines
			tConn = TimeoutConn(conn, time.Second*30)
			pConn = PeekConnection(tConn)
			log.Debug("Peeking connection %s => %s", conn.RemoteAddr(), conn.LocalAddr())
			_n, err := pConn.Peek(buffer)
			n = _n // avoid silly "variable not used" warning
			if err == io.EOF {
				return nil, nil
			} else if err != nil {
				log.Errorf(color.RedString("Could not peek bytes: %s", err.Error()))
				return nil, nil
			}
			peekUninitialized = false
		}
		if ch.CanHandle(buffer[:n]) {
			// Service supports payload
			// TODO(capacitorset): pass peeked connection rather than the actual one
			return service, pConn
		}
	}
	// No service can handle the connection. Let the caller deal with it.
	return nil, nil
}

func (hc *Honeytrap) heartbeat() {
	beat := time.Tick(30 * time.Second)

	count := 0

	for range beat {
		hc.bus.Send(event.New(
			event.Sensor("honeytrap"),
			event.Category("heartbeat"),
			event.SeverityInfo,
			event.Custom("sequence", count),
		))

		count++
	}
}

// Addr, proto, port, error
func ToAddr(input string) (net.Addr, string, int, error) {
	parts := strings.Split(input, "/")

	if len(parts) != 2 {
		return nil, "", 0, fmt.Errorf("Error parsing port string")
	}

	proto := parts[0]
	portStr := parts[1]
	portInt16, err := strconv.ParseInt(portStr, 10, 16)
	if err != nil {
		return nil, "", 0, fmt.Errorf("Error parsing port value: %s", err.Error())
	}
	port := int(portInt16)
	switch proto {
	case "tcp":
		addr, err := net.ResolveTCPAddr("tcp", ":"+portStr)
		return addr, proto, port, err
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", ":"+portStr)
		return addr, proto, port, err
	default:
		return nil, "", 0, fmt.Errorf("Unknown protocol %s", proto)
	}
}

func IsTerminal(f *os.File) bool {
	if isatty.IsTerminal(f.Fd()) {
		return true
	} else if isatty.IsCygwinTerminal(f.Fd()) {
		return true
	}

	return false
}

// Run will start honeytrap
func (hc *Honeytrap) Run(ctx context.Context) {
	if IsTerminal(os.Stdout) {
		fmt.Println(color.YellowString(`
 _   _                       _____                %c
| | | | ___  _ __   ___ _   |_   _| __ __ _ _ __
| |_| |/ _ \| '_ \ / _ \ | | || || '__/ _' | '_ \
|  _  | (_) | | | |  __/ |_| || || | | (_| | |_) |
|_| |_|\___/|_| |_|\___|\__, ||_||_|  \__,_| .__/
                        |___/              |_|
`, 127855))
	}

	fmt.Println(color.YellowString("Honeytrap starting (%s)...", hc.token))
	fmt.Println(color.YellowString("Version: %s (%s)", cmd.Version, cmd.ShortCommitID))

	log.Debugf("Using datadir: %s", hc.dataDir)

	go hc.heartbeat()

	hc.profiler.Start()

	w, err := web.New(
		web.WithEventBus(hc.bus),
		web.WithDataDir(hc.dataDir),
		web.WithConfig(hc.config.Web),
	)
	if err != nil {
		log.Error("Error parsing configuration of web: %s", err.Error())
	}

	w.Start()

	channels := map[string]pushers.Channel{}
	// sane defaults!

	for key, s := range hc.config.Channels {
		x := struct {
			Type string `toml:"type"`
		}{}

		err := toml.PrimitiveDecode(s, &x)
		if err != nil {
			log.Error("Error parsing configuration of channel: %s", err.Error())
			continue
		}

		if x.Type == "" {
			log.Error("Error parsing configuration of channel %s: type not set", key)
			continue
		}

		if channelFunc, ok := pushers.Get(x.Type); !ok {
			log.Error("Channel %s not supported on platform (%s)", x.Type, key)
		} else if d, err := channelFunc(
			pushers.WithConfig(s),
		); err != nil {
			log.Fatalf("Error initializing channel %s(%s): %s", key, x.Type, err)
		} else {
			channels[key] = d
		}
	}

	for _, s := range hc.config.Filters {
		x := struct {
			Channels   []string `toml:"channel"`
			Services   []string `toml:"services"`
			Categories []string `toml:"categories"`
		}{}

		err := toml.PrimitiveDecode(s, &x)
		if err != nil {
			log.Error("Error parsing configuration of filter: %s", err.Error())
			continue
		}

		for _, name := range x.Channels {
			channel, ok := channels[name]
			if !ok {
				log.Error("Could not find channel %s for filter", name)
				continue
			}

			channel = pushers.TokenChannel(channel, hc.token)

			if len(x.Categories) != 0 {
				channel = pushers.FilterChannel(channel, pushers.RegexFilterFunc("category", x.Categories))
			}

			if len(x.Services) != 0 {
				channel = pushers.FilterChannel(channel, pushers.RegexFilterFunc("service", x.Services))
			}

			if err := hc.bus.Subscribe(channel); err != nil {
				log.Error("Could not add channel %s to bus: %s", name, err.Error())
			}
		}
	}

	// initialize directors
	directors := map[string]director.Director{}
	availableDirectorNames := director.GetAvailableDirectorNames()

	for key, s := range hc.config.Directors {
		x := struct {
			Type string `toml:"type"`
		}{}

		err := toml.PrimitiveDecode(s, &x)
		if err != nil {
			log.Error("Error parsing configuration of director: %s", err.Error())
			continue
		}

		if x.Type == "" {
			log.Error("Error parsing configuration of service %s: type not set", key)
			continue
		}

		if directorFunc, ok := director.Get(x.Type); !ok {
			log.Error("Director type=%s not supported on platform (director=%s). Available directors: %s", x.Type, key, strings.Join(availableDirectorNames, ", "))
		} else if d, err := directorFunc(
			director.WithChannel(hc.bus),
			director.WithConfig(s),
		); err != nil {
			log.Fatalf("Error initializing director %s(%s): %s", key, x.Type, err)
		} else {
			directors[key] = d
		}
	}

	// initialize listener
	x := struct {
		Type string `toml:"type"`
	}{}

	if err := toml.PrimitiveDecode(hc.config.Listener, &x); err != nil {
		log.Error("Error parsing configuration of listener: %s", err.Error())
		return
	}

	if x.Type == "" {
		fmt.Println(color.RedString("Listener not set"))
	}

	var enabledDirectorNames []string
	for key := range directors {
		enabledDirectorNames = append(enabledDirectorNames, key)
	}

	hc.services = make(map[string]*ServiceMap)
	// same for proxies
	for key, s := range hc.config.Services {
		x := struct {
			Type     string `toml:"type"`
			Director string `toml:"director"`
			Port     string `toml:"port"`
		}{}

		if err := toml.PrimitiveDecode(s, &x); err != nil {
			log.Error("Error parsing configuration of service %s: %s", key, err.Error())
			continue
		}

		if x.Port != "" {
			log.Error("Ports in services are deprecated, add services to ports instead")
			continue
		}

		// individual configuration per service
		options := []services.ServicerFunc{
			services.WithChannel(hc.bus),
			services.WithConfig(s),
		}

		if x.Director == "" {
		} else if d, ok := directors[x.Director]; ok {
			options = append(options, services.WithDirector(d))
		} else {
			log.Error(color.RedString("Could not find director=%s for service=%s. Enabled directors: %s", x.Director, key, strings.Join(enabledDirectorNames, ", ")))
			continue
		}

		fn, ok := services.Get(x.Type)
		if !ok {
			log.Error(color.RedString("Could not find type %s for service %s", x.Type, key))
			continue
		}

		service := fn(options...)
		hc.services[key] = &ServiceMap{
			Service: service,
			Name:    key,
			Type:    x.Type,
		}
		log.Infof("Configured service %s (%s)", x.Type, key)
	}

	listenerFunc, ok := listener.Get(x.Type)
	if !ok {
		fmt.Println(color.RedString("Listener %s not support on platform", x.Type))
		return
	}

	l, err := listenerFunc(
		listener.WithChannel(hc.bus),
		listener.WithConfig(hc.config.Listener),
	)
	if err != nil {
		log.Fatalf("Error initializing listener %s: %s", x.Type, err)
	}

	hc.ports = make(map[int][]string)
	for _, s := range hc.config.Ports {
		x := struct {
			Port     string   `toml:"port"`
			Services []string `toml:"services"`
		}{}

		if err := toml.PrimitiveDecode(s, &x); err != nil {
			log.Error("Error parsing configuration of generic ports: %s", err.Error())
			continue
		}

		addr, _, port, err := ToAddr(x.Port)
		if err != nil {
			log.Error("Error parsing port string: %s", err.Error())
			continue
		}
		if addr == nil {
			log.Error("Failed to bind (addr is nil)")
			continue
		}

		if _, ok := hc.ports[port]; ok {
			log.Error("Port %d was already defined, ignoring the newer definition", port)
			continue
		}
		// check for the existence of the named services
		for _, service := range x.Services {
			found := false
			for name := range hc.services {
				if service == name {
					found = true
					break
				}
			}
			if !found {
				log.Error("Unknown service '%s' in ports", service)
			}
		}
		hc.ports[port] = x.Services

		a, ok := l.(listener.AddAddresser)
		if !ok {
			log.Error("Listener error")
			continue
		}
		a.AddAddress(addr)

		log.Infof("Configured port %s/%s", addr.Network(), addr.String())
	}

	if err := l.Start(ctx); err != nil {
		fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
		return
	}

	incoming := make(chan net.Conn)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}

			incoming <- conn
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-incoming:
			go hc.handle(conn)
		}
	}
}

func TimeoutConn(conn net.Conn, duration time.Duration) net.Conn {
	return &timeoutConn{
		conn,
		time.Duration(duration),
		time.Duration(duration),
	}
}

type timeoutConn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *timeoutConn) Read(b []byte) (int, error) {
	err := c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *timeoutConn) Write(b []byte) (int, error) {
	err := c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func (hc *Honeytrap) handle(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Errorf("Error: %s", err)
			log.Errorf("Stack of %d bytes: %s\n", count, string(trace))
			return
		}
	}()

	defer conn.Close()

	defer func() {
		if r := recover(); r != nil {
			message := event.Message("%+v", r)
			if err, ok := r.(error); ok {
				message = event.Message("%+v", err)
			}

			hc.bus.Send(event.New(
				event.SeverityFatal,
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Stack(),
				message,
			))
		}
	}()

	log.Debug("Accepted connection for %s => %s", conn.RemoteAddr(), conn.LocalAddr())
	defer log.Debug("Disconnected connection for %s => %s", conn.RemoteAddr(), conn.LocalAddr())

	/* conn is the original connection. newConn can be either the same
	 * connection, or a wrapper in the form of a PeekConnection.
	 */
	sm, newConn := hc.findService(conn)
	if sm == nil {
		log.Debug("No suitable handler for %s => %s", conn.RemoteAddr(), conn.LocalAddr())
		return
	}

	log.Debug("Handling connection for %s => %s %s(%s)", conn.RemoteAddr(), conn.LocalAddr(), sm.Name, sm.Type)

	ctx := context.Background()
	if err := sm.Service.Handle(ctx, newConn); err != nil {
		log.Errorf(color.RedString("Error handling service: %s: %s", sm.Name, err.Error()))
	}
}

// Stop will stop Honeytrap
func (hc *Honeytrap) Stop() {
	hc.profiler.Stop()

	fmt.Println(color.YellowString("Honeytrap stopped."))
}
