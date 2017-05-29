package proxies

import (
	"net"

	"github.com/op/go-logging"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	pushers "github.com/honeytrap/honeytrap/pushers"

	"github.com/BurntSushi/toml"
)

var log = logging.MustGetLogger("honeytrap:proxy")

// ProxyFn defines a function which delivers the needed listener for using the
// underline proxy connection.
type ProxyFn func(address string, m *director.ContainerConnections, d director.Director, p *pushers.Pusher, el pushers.Channel, c toml.Primitive) (net.Listener, error)

var proxies = map[string]ProxyFn{}

// Register hands the giving ProxyFn into the map with the giving name as key.
func Register(name string, fn ProxyFn) ProxyFn {
	proxies[name] = fn
	return fn
}

// Get returns the service function if it exists with the service name.
func Get(service string) (ProxyFn, bool) {
	fn, ok := proxies[service]
	return fn, ok
}

// ProxyConfig defines the configuration object delivered to a proxy creator.
type ProxyConfig struct {
	pusher *pushers.Pusher

	Config *config.Config
}

// Proxyer defines a interface which exposes a method to begin the internal
// proxy operation of it's implementer.
type Proxyer interface {
	Proxy() error
}
