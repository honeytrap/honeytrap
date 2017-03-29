package proxies

import "io"

// TCPProxyConn defines a struct which embeds the ProxyConn and provides
// the needed tcp operation call to proxy tcp connections.
type TCPProxyConn struct {
	ProxyConn
}

// Proxy defines a function to copy connection details from the server to a
// underline connection.
func (p TCPProxyConn) Proxy() error {
	defer func() {
		p.Close()
	}()

	go io.Copy(p.Conn, p.Server)
	_, err := io.Copy(p.Server, p.Conn)
	return err
}
