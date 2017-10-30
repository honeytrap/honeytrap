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
	"net"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/honeytrap/honeytrap/config"

	"github.com/honeytrap/honeytrap/director"
	_ "github.com/honeytrap/honeytrap/director/forward"
	_ "github.com/honeytrap/honeytrap/director/lxc"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/eventbus"

	"github.com/honeytrap/honeytrap/services"
	_ "github.com/honeytrap/honeytrap/services/vnc"

	"github.com/honeytrap/honeytrap/listener"
	_ "github.com/honeytrap/honeytrap/listener/canary"
	_ "github.com/honeytrap/honeytrap/listener/socket"
	_ "github.com/honeytrap/honeytrap/listener/tap"
	_ "github.com/honeytrap/honeytrap/listener/tun"

	// proxies

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/server/profiler"

	web "github.com/honeytrap/honeytrap/web"

	_ "github.com/honeytrap/honeytrap/pushers/console"       // Registers stdout backend.
	_ "github.com/honeytrap/honeytrap/pushers/elasticsearch" // Registers elasticsearch backend.
	_ "github.com/honeytrap/honeytrap/pushers/file"          // Registers file backend.
	_ "github.com/honeytrap/honeytrap/pushers/kafka"         // Registers kafka backend.
	_ "github.com/honeytrap/honeytrap/pushers/slack"         // Registers slack backend.

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

	matchers []ServiceMap
}

// New returns a new instance of a Honeytrap struct.
// func New(conf *config.Config) *Honeytrap {
func New(options ...OptionFn) *Honeytrap {
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
		fn(h)
	}

	return h
}

func (hc *Honeytrap) startAgentServer() {
	// as := proxies.NewAgentServer(hc.director, hc.pusher, hc.configig)
	// go as.ListenAndServe()
}

// EventServiceStarted will return a service started Event struct
func EventServiceStarted(service string, primitive toml.Primitive) event.Event {
	return event.New(
		event.Category(service),
		event.ServiceSensor,
		event.ServiceStarted,
	)
}

// PrepareRun will prepare Honeytrap to run
func (hc *Honeytrap) PrepareRun() {
}

type ServiceMap struct {
	Matcher func(net.Addr) bool

	Service services.Servicer

	Name string
	Type string
}

func (hc *Honeytrap) heartbeat() {
	beat := time.Tick(30 * time.Second)

	count := 0

	for {
		select {
		case <-beat:
			hc.bus.Send(event.New(
				event.Sensor("honeytrap"),
				event.Category("heartbeat"),
				event.SeverityInfo,
				event.Custom("sequence", count),
			))

			count++
		}
	}
}

// Run will start honeytrap
func (hc *Honeytrap) Run(ctx context.Context) {
	fmt.Println(color.YellowString("Honeytrap starting..."))

	go hc.heartbeat()

	hc.profiler.Start()

	w := web.New(
		web.WithEventBus(hc.bus),
	)

	go w.ListenAndServe()

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
			log.Error("Director %s not supported on platform (%s)", x.Type, key)
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

	err := toml.PrimitiveDecode(hc.config.Listener, &x)
	if err != nil {
		log.Error("Error parsing configuration of listener: %s", err.Error())
		return
	}

	if x.Type == "" {
		fmt.Println(color.RedString("Listener not set"))
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

	// same for proxies
	for key, s := range hc.config.Services {
		x := struct {
			Type     string `toml:"type"`
			Director string `toml:"director"`
			Port     string `toml:"port"`
		}{}

		err := toml.PrimitiveDecode(s, &x)
		if err != nil {
			log.Error("Error parsing configuration of service %s(%s): %s", x.Type, key, err.Error())
			continue
		}

		if x.Type == "" {
			log.Error("Error parsing configuration of service %s: type not set", key)
			continue
		}

		// individual configuration per service
		options := []services.ServicerFunc{
			services.WithChannel(hc.bus),
			services.WithConfig(s),
		}

		fn, ok := services.Get(x.Type)
		if !ok {
			log.Error(color.RedString("Could not find type %s for service %s", x.Type, key))
			continue
		}

		if x.Director == "" {
		} else if d, ok := directors[x.Director]; ok {
			options = append(options, services.WithDirector(d))
		} else {
			log.Error(color.RedString("Could not find director=%s for service=%s\n", x.Director, key))
			continue
		}

		service := fn(options...)

		parts := strings.Split(x.Port, "/")
		if len(parts) == 2 {
			log.Infof("Mapping port %s(%s) to service %s (%s)\n", parts[1], strings.ToLower(parts[0]), x.Type, key)

			// add address to listener and create mapping between
			// port and service
			if strings.ToLower(parts[0]) == "tcp" {
				if a, ok := l.(listener.AddAddresser); ok {
					addr, _ := net.ResolveTCPAddr("tcp", ":"+parts[1])
					a.AddAddress(addr)

					fn := func(port int) func(net.Addr) bool {
						return func(a net.Addr) bool {
							if ta, ok := a.(*net.TCPAddr); ok {
								return ta.Port == port
							}

							return false
						}
					}

					hc.matchers = append(hc.matchers, ServiceMap{
						Name:    key,
						Type:    x.Type,
						Matcher: fn(addr.Port),
						Service: service,
					})
				}
			} else if strings.ToLower(parts[0]) == "udp" {
				if a, ok := l.(listener.AddAddresser); ok {
					addr, _ := net.ResolveUDPAddr("udp", ":"+parts[1])
					a.AddAddress(addr)

					fn := func(port int) func(net.Addr) bool {

						return func(a net.Addr) bool {
							if ta, ok := a.(*net.UDPAddr); ok {
								return ta.Port == port
							}

							return false
						}
					}

					hc.matchers = append(hc.matchers, ServiceMap{
						Name:    key,
						Type:    x.Type,
						Matcher: fn(addr.Port),
						Service: service,
					})
				}
			}
		}
	}

	if err := l.Start(); err != nil {
		fmt.Println(color.RedString("Error starting listener: %s", err.Error()))
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
			hc.handle(conn)
		}
	}
}

func (hc *Honeytrap) handle(conn net.Conn) {
	for _, sm := range hc.matchers {
		if !sm.Matcher(conn.LocalAddr()) {
			continue
		}

		log.Debug("Handling connection for %s => %s %s(%s)", conn.RemoteAddr(), conn.LocalAddr(), sm.Name, sm.Type)

		go func(service services.Servicer) {
			err := service.Handle(conn)
			if err != nil {
				fmt.Println(color.RedString(err.Error()))
			}
		}(sm.Service)
	}
}

// Stop will stop Honeytrap
func (hc *Honeytrap) Stop() {
	hc.profiler.Stop()

	fmt.Println(color.YellowString("Honeytrap stopped."))
}
