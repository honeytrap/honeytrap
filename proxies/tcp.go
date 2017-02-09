package proxies

import "io"

type TCPProxyConn struct {
	ProxyConn
}

func (p TCPProxyConn) Proxy() error {
	defer func() {
		p.Close()
	}()

	go io.Copy(p.Conn, p.Server)
	_, err := io.Copy(p.Server, p.Conn)
	return err
}
