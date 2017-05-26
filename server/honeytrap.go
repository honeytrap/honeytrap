package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	_ "net/http/pprof"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	web "github.com/honeytrap/honeytrap-web"

	"github.com/honeytrap/honeytrap/canary"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/director/cowriedirector"
	"github.com/honeytrap/honeytrap/director/iodirector"
	"github.com/honeytrap/honeytrap/director/lxcdirector"

	proxies "github.com/honeytrap/honeytrap/proxies"
	_ "github.com/honeytrap/honeytrap/proxies/ssh" // TODO: Add comment

	"github.com/honeytrap/honeytrap/pushers/message"

	pushers "github.com/honeytrap/honeytrap/pushers"
	_ "github.com/honeytrap/honeytrap/pushers/backends/elasticsearch" // Registers elasticsearch backend.
	_ "github.com/honeytrap/honeytrap/pushers/backends/fschannel"     // Registers file backend.
	_ "github.com/honeytrap/honeytrap/pushers/backends/honeytrap"     // Registers honeytrap backend.
	_ "github.com/honeytrap/honeytrap/pushers/backends/slack"         // Registers slack backend.

	utils "github.com/honeytrap/honeytrap/utils"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("Honeytrap")

// Honeytrap defines a struct which coordinates the internal logic for the honeytrap
// container infrastructure.
type Honeytrap struct {
	config    *config.Config
	pusher    *pushers.Pusher
	events    pushers.Events
	honeycast *Honeycast
	director  director.Director
	manager   *director.ContainerConnections
}

// ServeFunc defines the function called to handle internal server details.
type ServeFunc func() error

// New returns a new instance of a Honeytrap struct.
func New(conf *config.Config) *Honeytrap {
	pusher := pushers.New(conf)
	pushChannel := pushers.NewProxyPusher(pusher)

	bus := pushers.NewEventBus()
	channels := pushers.ChannelStream{pushChannel, bus}
	events := pushers.NewTokenedEventDelivery(conf.Token, channels)

	var dr director.Director

	switch conf.Director {
	case cowriedirector.DirectorKey:
		dr = cowriedirector.New(conf, events)
	case iodirector.DirectorKey:
		dr = iodirector.New(conf, events)
	case lxcdirector.DirectorKey:
		dr = lxcdirector.New(conf, events)
	default:
		panic(fmt.Sprintf("Unknown director type: %q", conf.Director))
	}

	manager := director.NewContainerConnections()

	honeycast := NewHoneycast(conf, manager, dr, HoneycastAssets(&assetfs.AssetFS{
		Asset:     web.Asset,
		AssetDir:  web.AssetDir,
		AssetInfo: web.AssetInfo,
		Prefix:    web.Prefix,
	}))

	bus.Subscribe(honeycast)

	return &Honeytrap{
		config:    conf,
		director:  dr,
		pusher:    pusher,
		events:    events,
		honeycast: honeycast,
		manager:   manager,
	}
}

func (hc *Honeytrap) startAgentServer() {
	// as := proxies.NewAgentServer(hc.director, hc.pusher, hc.configig)
	// go as.ListenAndServe()
}

// ListenFunc defines a function type which returns a net.Listener specific for the
// use of its argument and for the reception of net connections.
type ListenFunc func(string, director.Director, *pushers.Pusher, *pushers.EventDelivery, *config.Config) (net.Listener, error)

// ListenerConfig defines a struct for holding configuration fields for a Listener
// builder.
type ListenerConfig struct {
	fn      ListenFunc
	address string
}

func (hc *Honeytrap) startPusher() {
	hc.pusher.Start()
}

// EventServiceStarted will return a service started Event struct
func EventServiceStarted(service string, primitive toml.Primitive) message.Event {
	return message.Event{
		Sensor:   service,
		Category: "Services",
		Type:     message.ServiceStarted,
		Details: map[string]interface{}{
			"primitive": primitive,
		},
	}
}

func (hc *Honeytrap) startProxies() {
	for _, primitive := range hc.config.Services {
		st := struct {
			Service string `toml:"service"`
			Port    string `toml:"port"`
		}{}

		if err := toml.PrimitiveDecode(primitive, &st); err != nil {
			log.Errorf("Error in service configuration: %s", err.Error())
			continue
		}

		if serviceFn, ok := proxies.Get(st.Service); ok {
			log.Debugf("Listener starting: %s", st.Port)

			service, err := serviceFn(st.Port, hc.manager, hc.director, hc.pusher, hc.events, primitive)
			if err != nil {
				log.Errorf("Error in service: %s: %s", st.Service, err.Error())

				hc.events.Deliver(message.Event{
					Sensor: st.Service,
					Type:   message.ServiceStarted,
					Details: map[string]interface{}{
						"primitive": primitive,
						"error":     err.Error(),
					},
				})

				continue
			}

			hc.events.Deliver(message.Event{
				Sensor: st.Service,
				Type:   message.ServiceStarted,
				Details: map[string]interface{}{
					"primitive": primitive,
				},
			})

			/*
				if err := toml.PrimitiveDecode(primitive, &service); err != nil {
					log.Errorf("Error in configuration for service: %s: %s", st.Service, err.Error())
					continue
				}
			*/

			/*
				l, err := net.Listen("tcp", st.Address)
				if err != nil {
					return nil, err
				}

				{
					&ProxyListener{l, d,
						p,
						c,
					},
				ProxyListener()
			*/

			go func(listener net.Listener) {
				/*
					l, err := net.Listen("tcp", address)
					if err != nil {
						return nil, err
					}
				*/

				// or just listener.Listen()
				defer listener.Close()

				for {
					conn, err := listener.Accept()
					if err != nil {
						log.Error(err.Error())
						continue
					}

					go func(conn net.Conn) {
						defer utils.RecoverHandler()

						defer func() {
							// TODO: add idle disconnect timeout? or should we add that to proxy conn self
							log.Info("Connection closed.")
							conn.Close()
						}()

						conn.(proxies.Proxyer).Proxy()
					}(conn)
				}
			}(service)
		}
	}
	/*
		listeners := []ListenerConfig{
			{
				fn:      proxies.ListenHTTP,
				address: hc.config.Proxies.HTTP.Port,
			},
			{
				fn:      proxies.ListenSMTP,
				address: hc.config.Proxies.SMTP.Port,
			},
			{
				fn:      proxies.ListenSSH,
				address: hc.config.Proxies.SSH.Port,
			},
			{
				fn:      proxies.ListenSIP,
				address: hc.config.Proxies.SIP.Port,
			},
		}

		for _, listener := range listeners {
		}
	*/
}

// startStatsServer starts the http server for handling request.
func (hc *Honeytrap) startStatsServer() {
	log.Infof("Stats server Listening on port: %s", hc.config.Web.Port)

	// if hc.config.Web.Path != "" {
	// 	log.Debug("Using static file path: ", hc.config.Web.Path)
	//
	// 	// check local css first
	// 	// TODO: What is this for and why are we assigning here.
	// 	// staticHandler = http.FileServer(http.Dir(hc.config.Web.Path))
	// }

	fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap server started, listening on address %s.", hc.config.Web.Port)))

	defer func() {
		fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap server stopped.")))
	}()

	if err := http.ListenAndServe(hc.config.Web.Port, hc.honeycast); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (hc *Honeytrap) startCanary() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	c, err := canary.New(ifaces, hc.events)
	if err != nil {
		return err
	}

	go c.Run()

	return nil
}

// Serve initializes and starts the internal logic for the Honeytrap instance.
func (hc *Honeytrap) Serve() {

	hc.startCanary()
	hc.startPusher()
	hc.startProxies()
	hc.startStatsServer()

	// hc.startAgentServer()
	//hc.startPing()
}
