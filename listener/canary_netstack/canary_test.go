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
		{ip: "::1", port: 80, proto: ip6},
	}

	//TCPAddr
	for _, tc := range testcases {
		ip := net.ParseIP(tc.ip).To4()
		if ip == nil {
			ip = net.ParseIP(tc.ip).To16()
		}
		netAddr := &net.TCPAddr{IP: ip, Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		t.Logf("IP: %s", tc.ip)
		t.Logf("len full.Addr: %d", len(full.Addr))
		if full.Addr != tcpip.Address(netAddr.IP) {
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
		ip := net.ParseIP(tc.ip).To4()
		if ip == nil {
			ip = net.ParseIP(tc.ip).To16()
		}
		netAddr := &net.TCPAddr{IP: ip, Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		if full.Addr != tcpip.Address(netAddr.IP) {
			t.Errorf("udp: address not correct, want %s, got %s", tc.ip, full.Addr.String())
		}
		if int(full.Port) != tc.port {
			t.Errorf("udp: port not correct, want: %d, got %d", tc.port, full.Port)
		}
		if prot != tc.proto {
			t.Errorf("udp: network protocol not correct, want: %d, got %d", tc.proto, prot)
		}
	}
}

func TestParseAddrBad(t *testing.T) {
	ip4 := ipv4.ProtocolNumber
	ip6 := ipv6.ProtocolNumber

	testcases := []struct {
		ip    string
		port  int
		proto tcpip.NetworkProtocolNumber
	}{
		{ip: "192.0.2", port: 80, proto: ip4},
		{ip: "192", port: 80, proto: ip6},
		{ip: "", port: 80, proto: ip4},
	}

	//TCPAddr
	for _, tc := range testcases {
		ip := net.ParseIP(tc.ip).To4()
		if ip == nil {
			ip = net.ParseIP(tc.ip).To16()
		}
		netAddr := &net.TCPAddr{IP: ip, Port: tc.port}
		full, prot := parseAddr(netAddr, 1)
		t.Logf("IP: %s", tc.ip)
		t.Logf("len full.Addr: %d", len(full.Addr))
		if full.Addr != tcpip.Address("") {
			t.Errorf("tcp: address not correct, want '%s', got %s", "", full.Addr.String())
		}
		if int(full.Port) != 0 {
			t.Errorf("tcp: port not correct, want: %d, got %d", 0, full.Port)
		}
		if prot != 0 {
			t.Errorf("tcp: %s network protocol not correct, want: %d, got %d", tc.ip, 0, prot)
		}
	}
}
