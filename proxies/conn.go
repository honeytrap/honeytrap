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

	Event pushers.Channel
}

// Write calls the internal connection write method and submits a method for such a data.
func (cw *ProxyConn) Write(p []byte) (int, error) {
	defer cw.Event.Send(DataWriteEvent(cw.Conn, p, map[string]interface{}{
		"container": cw.Container.Detail(),
	}))

	n, err := cw.Conn.Write(p)
	if err != nil {
		cw.Event.Send(ConnectionWriteErrorEvent(cw.Conn, err))
		return n, err
	}

	return n, nil
}

// Read calls the internal connection read method and submits a method for such a data.
func (cw *ProxyConn) Read(p []byte) (int, error) {
	var n int
	var err error

	defer cw.Event.Send(DataReadEvent(cw.Conn, p[:n], map[string]interface{}{
		"container": cw.Container.Detail(),
	}))

	n, err = cw.Conn.Read(p)
	if err != nil {
		cw.Event.Send(ConnectionReadErrorEvent(cw.Conn, err))
		return n, err
	}

	return n, nil
}

// RemoteHost returns the addr ip of the giving connection.
func (cw *ProxyConn) RemoteHost() string {
	host, _, _ := net.SplitHostPort(cw.RemoteAddr().String())
	return host
}

// Close closes the ProxyConn internal net.Conn.
func (cw *ProxyConn) Close() error {
	if cw.Server != nil {
		cw.Event.Send(ConnectionClosedEvent(cw.Server))

		cw.Server.Close()
	}

	if cw.Conn != nil {
		cw.Event.Send(ConnectionClosedEvent(cw.Conn))

		return cw.Conn.Close()
	}

	return nil
}
