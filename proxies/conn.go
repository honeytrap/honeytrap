package proxies

import (
	"net"

	providers "github.com/honeytrap/honeytrap/providers"
	pushers "github.com/honeytrap/honeytrap/pushers"
)

type ProxyConn struct {
	// Connection with host
	net.Conn
	// Connection to container
	Server net.Conn

	Container providers.Container

	Pusher *pushers.Pusher
}

func (cw *ProxyConn) RemoteHost() string {
	host, _, _ := net.SplitHostPort(cw.RemoteAddr().String())
	return host
}

func (cw *ProxyConn) Close() error {
	if cw.Server != nil {
		cw.Server.Close()
	}
	if cw.Conn != nil {
		return cw.Conn.Close()
	}

	return nil
}
