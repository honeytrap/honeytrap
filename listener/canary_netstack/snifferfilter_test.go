package nscanary

import (
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip"
)

func TestBlockPortFn(t *testing.T) {
	testcases := []struct {
		filter []string
		proto  string
		port   uint16
		want   bool
	}{
		{filter: []string{}, proto: "tcp", port: 0, want: false},
		{filter: []string{}, proto: "udp", port: 0, want: false},
		{filter: []string{"udp/0"}, proto: "tcp", port: 0, want: false},
		{filter: []string{"udp/22"}, proto: "tcp", port: 22, want: false},
		{filter: []string{"tcp/22"}, proto: "tcp", port: 1, want: false},
		{filter: []string{"tcp/1", "tcp/2"}, proto: "tcp", port: 2, want: true},
		{filter: []string{"tcp/1", "tcp/2"}, proto: "tcp", port: 3, want: false},
		{filter: []string{"udp/1", "udp/2"}, proto: "udp", port: 3, want: false},
		{filter: []string{"udp/1", "udp/2"}, proto: "udp", port: 1, want: true},
		{filter: []string{"tcp/22", "udp/1"}, proto: "tcp", port: 1, want: false},
		{filter: []string{"tcp/22", "udp/1"}, proto: "udp", port: 1, want: true},
		{filter: []string{"tcp/22", "udp/1"}, proto: "udp", port: 22, want: false},
		{filter: []string{"tcp/0"}, proto: "udp", port: 0, want: false},
		{filter: []string{"tcp/22"}, proto: "udp", port: 22, want: false},
		{filter: []string{"/22"}, proto: "udp", port: 22, want: false},
		{filter: []string{"/22"}, proto: "tcp", port: 22, want: false},
		{filter: []string{"22"}, proto: "tcp", port: 22, want: false},
	}

	for _, tc := range testcases {
		blocked := BlockPortFn(tc.filter, tc.proto)(tc.port)
		if blocked != tc.want {
			t.Errorf("BlockPortFn(%v, %s)(%d) = %v: want: %v", tc.filter, tc.proto, tc.port, blocked, tc.want)
		}
	}
}

func TestBlockIPFn(t *testing.T) {
	testcases := []struct {
		filter []string
		addr   []byte
		want   bool
	}{
		{filter: []string{}, addr: nil, want: false},
		{filter: []string{}, addr: []byte{1, 2, 3, 4}, want: false},
		{filter: []string{}, addr: []byte{0, 0, 0, 0}, want: false},
		{filter: []string{"::1"}, addr: []byte{0, 0, 0, 1}, want: false},
		{filter: []string{"1.2.3.4"}, addr: []byte{5, 6, 7, 8}, want: false},
		{filter: []string{"1.2.3.4"}, addr: []byte{1, 2, 3, 4}, want: true},
		{filter: []string{"1.2.3.4", "5.6.7.8"}, addr: []byte{5, 6, 7, 8}, want: true},
		{filter: []string{"::ffff:c0a8:5909"}, addr: []byte{192, 168, 89, 9}, want: true},
	}

	for _, tc := range testcases {
		blocked := BlockIPFn(tc.filter)(tcpip.Address(tc.addr))
		if blocked != tc.want {
			t.Errorf("BlockIPFn(%v) = %v: want %v", tc.filter, blocked, tc.want)
		}
	}
}
