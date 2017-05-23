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
	ListContainers() []ContainerDetail
}

// ContainerDetail defines a struct which is used to detail specific container meta-data.
type ContainerDetail struct {
	Name          string                 `json:"name"`
	ContainerAddr string                 `json:"container_addr"`
	Meta          map[string]interface{} `json:"meta"`
}

// Container defines a type which exposes methods for connecting to a container.
type Container interface {
	Name() string
	Dial(context.Context) (net.Conn, error)
	Detail() ContainerDetail
}
