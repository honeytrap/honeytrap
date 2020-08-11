package nscanary

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pkg/peek"
	"github.com/honeytrap/honeytrap/pushers"
)

var tlsVersion = map[uint16]string{
	tls.VersionTLS10: "1.0",
	tls.VersionTLS11: "1.1",
	tls.VersionTLS12: "1.2",
	tls.VersionTLS13: "1.3",
	tls.VersionSSL30: "SSL 3.0",
}

type TLS map[uint16]*tls.Config

func (t TLS) AddConfig(port uint16, c *tls.Config) {
	t[port] = c
}

// MaybeTLS checks for a tls signature and does a tls handshake if it is tls.
// return a tls.Conn or the given connection.
func (t TLS) MaybeTLS(conn net.Conn, port uint16, events pushers.Channel) (net.Conn, error) {
	config := t[port]
	if config == nil {
		config = t[0]
	}
	if config == nil {
		// tls not available.
		return conn, nil
	}

	var signature [3]byte

	pconn := peek.NewConn(conn)
	if _, err := pconn.Peek(signature[:]); err != nil {
		pconn.Close()
		return nil, err
	}

	fmt.Printf("signature = %x\n", signature)

	if signature[0] == 0x16 && signature[1] == 0x03 && signature[2] <= 0x03 {
		// tls signature found,
		return NewTLSConn(pconn, config, events)
	}

	return pconn, nil
}

type TLSConn struct {
	net.Conn

	events pushers.Channel
}

// NewTLSConn sets up a tls connection which captures the payloads.
// if the passed tls config is nil it will panic.
func NewTLSConn(conn net.Conn, conf *tls.Config, events pushers.Channel) (*TLSConn, error) {
	tlsConn := tls.Server(conn, conf)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	//TODO (jerry): Send tls data (JA3??)
	state := tlsConn.ConnectionState()

	events.Send(event.New(
		CanaryOptions,
		event.Category("tcp"),
		event.Type("tls"),
		event.SourceAddr(tlsConn.RemoteAddr()),
		event.DestinationAddr(tlsConn.LocalAddr()),
		event.Custom("tls-version", tlsVersion[state.Version]),
		event.Custom("tls-ciphersuite", tls.CipherSuiteName(state.CipherSuite)),
	))

	c := &TLSConn{
		Conn:   tlsConn,
		events: events,
	}

	return c, nil
}

func (t *TLSConn) Read(p []byte) (int, error) {
	buf := make([]byte, len(p))

	n, err := t.Conn.Read(buf)

	t.events.Send(event.New(
		CanaryOptions,
		event.Category("tls"),
		event.Type("tls"),
		event.SourceAddr(t.Conn.RemoteAddr()),
		event.DestinationAddr(t.Conn.LocalAddr()),
		event.Payload(buf[:n]),
	))

	copy(p, buf)
	return n, err
}
