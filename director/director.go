package director

import (
	"context"
	"net"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:director")

// Director defines an interface which exposes an interface to allow structures that
// implement this interface allow us to control containers which they provide.
type Director interface {
	NewContainer(string) (Container, error)
	GetContainer(net.Conn) (Container, error)
}

// Container defines a type which exposes methods for connecting to a container.
type Container interface {
	Dial(context.Context) (net.Conn, error)
	Name() string
}
