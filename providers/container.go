package providers

import "net"

// Container defines a type which exposes methods for connecting to a container.
type Container interface {
	Dial(string) (net.Conn, error)
	Name() string
}
