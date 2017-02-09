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

type ProxyFn func(address string, d *director.Director, p *pushers.Pusher, c toml.Primitive) (net.Listener, error)

var proxies = map[string]ProxyFn{}

func Register(name string, fn ProxyFn) ProxyFn {
	proxies[name] = fn
	return fn
}

func Get(service string) (ProxyFn, bool) {
	fn, ok := proxies[service]
	return fn, ok
}

type ProxyConfig struct {
	pusher *pushers.Pusher
	Config *config.Config
}

type Proxyer interface {
	Proxy() error
}
