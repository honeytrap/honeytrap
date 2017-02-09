package proxies

import (
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"
	"net"
)

type ProxyListener struct {
	net.Listener
	director *director.Director
	pusher   *pushers.Pusher
}

func NewProxyListener(l net.Listener, d *director.Director, p *pushers.Pusher) *ProxyListener {
	return &ProxyListener{
		l,
		d,
		p,
	}
}

func (lw *ProxyListener) Accept() (c net.Conn, err error) {
	c, err = lw.Listener.Accept()
	if err != nil {
		return nil, err
	}

	container, err := lw.director.GetContainer(c)
	if err != nil {
		c.Close()
		return nil, err
	}

	_, port, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		c.Close()
		return nil, err
	}

	log.Debugf("Connecting to container port: %s", port)

	var c2 net.Conn
	c2, err = container.Dial(port)
	if err != nil {
		c.Close()
		return nil, err
	}

	return &ProxyConn{c, c2, container, lw.pusher}, err
}

func (lw *ProxyListener) Close() error {
	log.Info("Listener closed")
	return lw.Listener.Close()
}

func (lw *ProxyListener) Addr() net.Addr {
	return lw.Listener.Addr()
}
