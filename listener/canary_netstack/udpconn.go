package nscanary

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
)

//UDPConn extends gonet.UDPConn.
type UDPConn struct {
	*gonet.UDPConn
}

//WriteToUDP satifies listener.DummyUDPConn.Fn
func (c *UDPConn) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return c.WriteTo(b, net.Addr(addr))
}

func (c *UDPConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, addr, err := c.ReadFrom(b)
	return n, addr.(*net.UDPAddr), err
}
