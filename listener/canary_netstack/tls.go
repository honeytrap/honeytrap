package nscanary

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pkg/peek"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/open-ch/ja3"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/waiter"
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

func (t TLS) MaybeTLS(ep tcpip.Endpoint, wq *waiter.Queue, port uint16, events pushers.Channel) (net.Conn, error) {
	log.Debugf("maybe tls, port: %d", port)

	// find the tls.Config for this port.
	config := t[port]
	if config == nil {
		config = t[0]
	}

	var hello [16 * 1024]byte // if tls this is the  client hello record.
	var helloLen int

	waitEntry, notifyCh := waiter.NewChannelEntry(nil)

	wq.EventRegister(&waitEntry, waiter.EventIn)

	for {
		n, _, err := ep.Peek([][]byte{hello[:]})
		if err != nil {
			if err == tcpip.ErrWouldBlock {
				<-notifyCh
				continue
			}
			return nil, fmt.Errorf("peek error: %s", err)
		}
		helloLen = int(n)
		log.Debugf("peeked %d bytes", n)
		break
	}
	wq.EventUnregister(&waitEntry)

	var isTLS bool

	j, err := ja3.ComputeJA3FromSegment(hello[:helloLen])
	if err != nil {
		// If the packet is no Client Hello an error is thrown as soon as the parsing fails
		// isTLS = false
		log.Debugf("compute ja3: %v", err)
	} else {
		isTLS = true
	}

	conn := gonet.NewTCPConn(wq, ep)

	if isTLS {
		log.Debugf("found tls signature on port: %d", port)

		if config == nil {
			// tls not available.
			log.Debugf("no tls config found for port: %d", port)
			return conn, nil
		}

		c, eopt, err := NewTLSConn(conn, config, events)
		if err != nil {
			return nil, err
		}

		// get the plaintext payload.
		tconn := peek.NewConn(c)
		buf := make([]byte, 65535)
		n, err := tconn.Peek(buf)
		if err != nil {
			log.Debugf("Peek: %v", err)
		}

		events.Send(event.New(
			event.Category("tcp"),
			event.Type("tls"),
			event.SourceAddr(tconn.RemoteAddr()),
			event.DestinationAddr(tconn.LocalAddr()),
			event.Payload(buf[:n]),
			event.Custom("ja3.digest", j.GetJA3Hash()),
			event.Custom("ja3.string", j.GetJA3String()),
			event.Custom("tls.sni", j.GetSNI()),
			event.Custom("tls.client.hello.payload", string(hello[:helloLen])),
			eopt,
		))

		return tconn, nil
	}

	return conn, nil
}

type TLSConn struct {
	net.Conn

	events pushers.Channel
}

func NewTLSConn(conn net.Conn, conf *tls.Config, events pushers.Channel) (*TLSConn, event.Option, error) {

	tlsConn := tls.Server(conn, conf)
	if err := tlsConn.Handshake(); err != nil {
		return nil, nil, fmt.Errorf("tls handshake error: %v", err)
	}

	state := tlsConn.ConnectionState()

	eopts := event.NewWith(
		event.Custom("tls.version", tlsVersion[state.Version]),
		event.Custom("tls.ciphersuite", tls.CipherSuiteName(state.CipherSuite)),
	)

	c := &TLSConn{
		Conn:   tlsConn,
		events: events,
	}

	return c, eopts, nil
}
