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

	testcases := []struct {
		ip    string
		port  int
		proto tcpip.NetworkProtocolNumber
	}{
		{ip: "192.0.2.33", port: 80, proto: ip4},
		{ip: "0.0.0.0", port: 80, proto: ip4},
		{ip: "", port: 80, proto: ip4},
		{ip: "2001:db8::123:12:1", port: 80, proto: ip6},
		{ip: "::1", port: 80, proto: ip6},
	}

	//TCPAddr
	for _, tc := range testcases {
		netAddr := &net.TCPAddr{IP: net.ParseIP(tc.ip), Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		if full.Addr.String() != tc.ip {
			t.Errorf("tcp: address not correct, want %s, got %s", tc.ip, full.Addr.String())
		}
		if int(full.Port) != tc.port {
			t.Errorf("tcp: port not correct, want: %d, got %d", tc.port, full.Port)
		}
		if prot != tc.proto {
			t.Errorf("tcp: network protocol not correct, want: %d, got %d", tc.proto, prot)
		}
	}

	//UDPAddr
	for _, tc := range testcases {
		netAddr := &net.UDPAddr{IP: net.ParseIP(tc.ip), Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		if full.Addr.String() != tc.ip {
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
