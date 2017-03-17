package server

import (
	"net"

	_ "net/http/pprof"

	"github.com/BurntSushi/toml"
	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"

	proxies "github.com/honeytrap/honeytrap/proxies"
	_ "github.com/honeytrap/honeytrap/proxies/ssh"

	pushers "github.com/honeytrap/honeytrap/pushers"
	// _ "github.com/honeytrap/honeytrap/pushers/elasticsearch"
	// _ "github.com/honeytrap/honeytrap/pushers/honeytrap"
	_ "github.com/honeytrap/honeytrap/pushers/slack"

	utils "github.com/honeytrap/honeytrap/utils"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap")

type honeytrap struct {
	config   *config.Config
	director *director.Director
	pusher   *pushers.Pusher
}

type ServeFunc func() error

func New(conf *config.Config) *honeytrap {
	director := director.New(conf)
	pusher := pushers.New(conf)
	return &honeytrap{conf, director, pusher}
}

func (hc *honeytrap) startAgentServer() {
	// as := proxies.NewAgentServer(hc.director, hc.pusher, hc.config)
	// go as.ListenAndServe()
}

type ListenFunc func(string, *director.Director, *pushers.Pusher, *config.Config) (net.Listener, error)

type ListenerConfig struct {
	fn      ListenFunc
	address string
}

func (hc *honeytrap) startPusher() {
	hc.pusher.Start()
}

func (hc *honeytrap) startProxies() {
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

			service, err := serviceFn(st.Port, hc.director, hc.pusher, primitive)
			if err != nil {
				log.Errorf("Error in service: %s: %s", st.Service, err.Error())
				continue
			}

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

func (hc *honeytrap) Serve() {

	hc.startPusher()
	hc.startProxies()
	hc.startStatsServer()

	// hc.startAgentServer()
	//hc.startPing()
}
