package nscanary

import "testing"

func TestEcludeLogProtos(t *testing.T) {
	testcases := []struct {
		exclude []string
		want    uint32
	}{
		{exclude: []string{}, want: ProtoAll},
		{exclude: []string{"ip4"}, want: ProtoAll &^ ProtoIPv4},
		{exclude: []string{"unknown"}, want: ProtoAll},
		{exclude: []string{"unknown", "ip4"}, want: ProtoAll &^ ProtoIPv4},
		{exclude: []string{"ip4", "ip6", "arp", "udp", "tcp", "icmp"}, want: 0},
	}

	for _, tc := range testcases {
		ExcludeLogProtos(tc.exclude)
		if LogProtos != tc.want {
			t.Errorf("Logprotos = %d, want: %d", LogProtos, tc.want)
		}
	}
}
