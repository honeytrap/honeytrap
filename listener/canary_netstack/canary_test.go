package nscanary

import (
	"net"
	"reflect"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
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

func TestGetTransportProtos(t *testing.T) {
	testcases := []struct {
		protos []string
		want   []stack.TransportProtocol
	}{
		{protos: []string{}, want: []stack.TransportProtocol{tcp.NewProtocol(), udp.NewProtocol(), icmp.NewProtocol4(), icmp.NewProtocol6()}},
		{protos: []string{"tcp"}, want: []stack.TransportProtocol{tcp.NewProtocol()}},
		{protos: []string{"udp"}, want: []stack.TransportProtocol{udp.NewProtocol()}},
		{protos: []string{"icmp4"}, want: []stack.TransportProtocol{icmp.NewProtocol4()}},
		{protos: []string{"icmp6"}, want: []stack.TransportProtocol{icmp.NewProtocol6()}},
	}

	for _, tc := range testcases {
		got, _ := getTransportProtos(tc.protos)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("want: %v, got: %v", tc.want, got)
		}
	}
}

func TestGetTransportProtosErr(t *testing.T) {
	bad := "bad"

	_, err := getTransportProtos([]string{bad})
	t.Log(err)

	if err == nil {
		t.Errorf("proto: '%s' expected an error, got none", bad)
	}
}
