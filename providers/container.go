package providers

import "net"

type Container interface {
	Dial(string) (net.Conn, error)
	Name() string
}
