package nscanary

import (
	"reflect"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

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
