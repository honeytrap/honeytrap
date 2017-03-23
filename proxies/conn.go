package proxies

import (
	"net"

	providers "github.com/honeytrap/honeytrap/providers"
	pushers "github.com/honeytrap/honeytrap/pushers"
)

// ProxyConn defines a base decorator over a net.Conn for proxy purposes.
type ProxyConn struct {
	// Connection with host
	net.Conn
	// Connection to container
	Server net.Conn

	Container providers.Container

	Pusher *pushers.Pusher
	Event  *pushers.EventDelivery
}

// RemoteHost returns the addr ip of the giving connection.
func (cw *ProxyConn) RemoteHost() string {
	host, _, _ := net.SplitHostPort(cw.RemoteAddr().String())
	return host
}

// Close closes the ProxyConn internal net.Conn.
func (cw *ProxyConn) Close() error {
	if cw.Server != nil {

		ev := EventConnectionClosed(cw.RemoteAddr(), "ProxyConn.Conn", nil)
		cw.Event.Deliver(ev)
		cw.Server.Close()
	}

	if cw.Conn != nil {
		cw.Event.Deliver(EventConnectionClosed(cw.RemoteAddr(), "ProxyConn.Conn", nil))
		return cw.Conn.Close()
	}

	return nil
}
