package ldap

import (
	"reflect"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
)

var (
	tlsRequest = []byte{
		0x30, 0x1d, // Begin the LDAPMessage sequence
		0x02, 0x01, 0x01, // The message ID (integer value 1)
		0x77, 0x18, // Begin the extended request protocol op
		0x80, 0x16, 0x31, 0x2e, 0x33, 0x2e, 0x36, 0x2e, 0x31, 0x2e, // The extended request OID
		0x34, 0x2e, 0x31, 0x2e, 0x31, 0x34, 0x36, 0x36, // (octet string "1.3.6.1.4.1.1466.20037"
		0x2e, 0x32, 0x30, 0x30, 0x33, 0x37, // with type context-specific primitive zero)
	}
	tlsResponse_success = []byte{
		0x30, 0x0c, // Begin the LDAPMessage sequence
		0x02, 0x01, 0x01, // The message ID (integer value 1)
		0x78, 0x07, // Begin the extended response protocol op
		0x0a, 0x01, 0x00, // success result code (enumerated value 0)
		0x04, 0x00, // No matched DN (0-byte octet string)
		0x04, 0x00, // No diagnostic message (0-byte octet string)
	}
)

func TestHandleTLS(t *testing.T) {
	cases := []struct {
		arg, want []byte
	}{
		{tlsRequest, tlsResponse_success},
	}

	h := &extFuncHandler{
		tlsFunc: func() error {
			// TLS always succeeds
			return nil
		},
	}

	for _, c := range cases {
		p := ber.DecodePacket(c.arg)
		el := make(eventLog)
		got := h.handle(p, el)

		if len(got) > 0 {
			if !reflect.DeepEqual(got[0].Bytes(), c.want) {
				t.Errorf("StartTLS:\nwant %v\n got %v", c.want, got[0].Bytes())
			}
		} else {
			t.Errorf("StartTLS: want %v got nothing", c.want)
		}

		if rtype, ok := el["ldap.request-type"]; !ok || rtype != "extended.tls" {
			t.Errorf("StartTLS: Wrong request type, want 'tls', got %s", rtype)
		}
	}
}
