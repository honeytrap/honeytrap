package proxies

import (
	"net"

	"github.com/honeytrap/honeytrap/director"
	pushers "github.com/honeytrap/honeytrap/pushers"
)

// ProxyConn defines a base decorator over a net.Conn for proxy purposes.
type ProxyConn struct {
	// Connection with host
	net.Conn
	// Connection to container
	Server net.Conn

	Container director.Container

	Pusher *pushers.Pusher
	Event  pushers.Events
}

// RemoteHost returns the addr ip of the giving connection.
func (cw *ProxyConn) RemoteHost() string {
	host, _, _ := net.SplitHostPort(cw.RemoteAddr().String())
	return host
}

// Close closes the ProxyConn internal net.Conn.
func (cw *ProxyConn) Close() error {
	if cw.Server != nil {

		cw.Event.Deliver(EventConnectionClosed(cw.RemoteAddr(), cw.LocalAddr(), "ProxyConn.Conn", nil, nil))
		cw.Server.Close()
	}

	if cw.Conn != nil {
		cw.Event.Deliver(EventConnectionClosed(cw.RemoteAddr(), cw.LocalAddr(), "ProxyConn.Conn", nil, nil))
		return cw.Conn.Close()
	}

	return nil
}
