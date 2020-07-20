package nscanary

import (
	"net"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
)

func TestParseAddr(t *testing.T) {
	ip4 := ipv4.ProtocolNumber
	ip6 := ipv6.ProtocolNumber

	addr6 := make(net.IP, 16)
	addr6[15] = 1

	testcases := []struct {
		ip    net.IP
		port  int
		proto tcpip.NetworkProtocolNumber
	}{
		{ip: net.IP{192, 0, 2, 23}, port: 80, proto: ip4},
		{ip: net.IP{0, 0, 0, 0}, port: 80, proto: ip4},
		{ip: net.IP{}, port: 80, proto: 0},
		{ip: addr6, port: 80, proto: ip6},
	}

	//TCPAddr
	for _, tc := range testcases {
		netAddr := &net.TCPAddr{IP: tc.ip, Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		if full.Addr != tcpip.Address(tc.ip) {
			t.Errorf("tcp: address not correct, want %s, got %s", tc.ip, full.Addr.String())
		}
		if int(full.Port) != tc.port {
			t.Errorf("tcp: port not correct, want: %d, got %d", tc.port, full.Port)
		}
		if prot != tc.proto {
			t.Errorf("tcp: %s network protocol not correct, want: %d, got %d", tc.ip, tc.proto, prot)
		}
	}

	//UDPAddr
	for _, tc := range testcases {
		netAddr := &net.UDPAddr{IP: tc.ip, Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		if full.Addr != tcpip.Address(tc.ip) {
			t.Errorf("tcp: address not correct, want %s, got %s", tc.ip, full.Addr.String())
		}
		if int(full.Port) != tc.port {
			t.Errorf("tcp: port not correct, want: %d, got %d", tc.port, full.Port)
		}
		if prot != tc.proto {
			t.Errorf("tcp: network protocol not correct, want: %d, got %d", tc.proto, prot)
		}
	}
}
