package proxies

import (
	"context"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"
)

// ProxyListener defines a struct which holds a giving net.Listener.
type ProxyListener struct {
	net.Listener

	events   pushers.Channel
	director director.Director
	manager  *director.ContainerConnections
}

// NewProxyListener returns a new instance for a ProxyListener.
func NewProxyListener(l net.Listener, m *director.ContainerConnections, d director.Director, e pushers.Channel) *ProxyListener {
	return &ProxyListener{
		Listener: l,
		director: d,
		events:   e,
		manager:  m,
	}
}

// Accept returns a new net.Conn from the underline Proxy Listener.
func (lw *ProxyListener) Accept() (c net.Conn, err error) {
	c, err = lw.Listener.Accept()
	if err != nil {
		return nil, err
	}

	lw.events.Send(ConnectionOpenedEvent(c))

	// Attempt to GetContainer from director.
	container, err := lw.director.GetContainer(c)
	if err != nil {
		log.Error("proxy : %q : Failed to retrieve container : %q : %+q", c.RemoteAddr().String(), err)

		// Container does not exists on director, so ask for new one.
		container, err = lw.director.NewContainer(c.RemoteAddr().String())
		if err != nil {
			log.Error("proxy : %q : Failed to create container : %q : %+q", c.RemoteAddr().String(), err)

			lw.events.Send(ConnectionClosedEvent(c))

			c.Close()
			return nil, err
		}
	}

	lw.events.Send(UserSessionOpenedEvent(c, container.Detail(), nil))

	_, port, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		//TODO(alex): Decide what to do if this fails
	}

	// log.Debugf("Connecting to container port: %s", port)

	var c2 net.Conn

	// TODO(alex): Decide if changing the signature makes sense and if it does, shouldn't
	// there therefore be a time-stamp added to use the deadline capability of context?
	c2, err = container.Dial(context.Background(), port)
	if err != nil {
		log.Error("proxy : %q : Failed to dail container : %q : %+q", c.RemoteAddr().String(), err)

		lw.events.Send(UserSessionClosedEvent(c, container.Detail()))

		lw.events.Send(ConnectionClosedEvent(c))

		c.Close()
		return nil, err
	}

	proxyConn := &ProxyConn{
		Conn:      c,
		Server:    c2,
		Container: container,
		Event:     lw.events,
	}

	lw.manager.AddClient(proxyConn, container.Detail())

	return proxyConn, err
}

// Close closes the underline net.Listener.
func (lw *ProxyListener) Close() error {
	log.Info("Listener closed")

	lw.events.Send(ListenerClosedEvent(lw.Listener))

	return lw.Listener.Close()
}

// Addr returns the underline address of the internal net.Listener.
func (lw *ProxyListener) Addr() net.Addr {
	return lw.Listener.Addr()
}
