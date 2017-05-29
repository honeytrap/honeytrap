package proxies

import (
	"io"
	"net"
	"time"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	proxies "github.com/honeytrap/honeytrap/proxies"
	pushers "github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
	"github.com/satori/go.uuid"
)

var log = logging.MustGetLogger("honeytrap:proxy:sip")

// ListenSIP returns a new net.Listener for listening for sip connections.
// TODO: Change amount of params.
func ListenSIP(address string, m *director.ContainerConnections, d director.Director, p *pushers.Pusher, e pushers.Channel, c *config.Config) (net.Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	return &SIPProxyListener{
		proxies.NewProxyListener(l, m, d, p, e),
	}, nil
}

// SIPProxyListener defines a struct which embedds and handle the underline connection
// details for the provided ProxyListener.
type SIPProxyListener struct {
	*proxies.ProxyListener
}

// Accept returns a new connection from the underlined listener.
func (l *SIPProxyListener) Accept() (net.Conn, error) {
	conn, err := l.ProxyListener.Accept()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return SIPProxyConn{conn.(*proxies.ProxyConn)}, err
}

// SIPProxyConn defines a SIPProxyConn to handle proxying details to a underline
// proxy connection.
type SIPProxyConn struct {
	*proxies.ProxyConn
}

// SIPAction defines a data struct for holding a SIP action request/response.
type SIPAction struct {
	Date          time.Time `json:"timestamp"`
	Host          string    `json:"host"`
	URL           string    `json:"url"`
	RemoteAddr    string    `json:"remote_addr"`
	Method        string    `json:"method"`
	Referer       string    `json:"referer"`
	ContentLength int64     `json:"content_length"`
}

// Proxy initiates and proxies the underline proxy connections.
func (p SIPProxyConn) Proxy() error {
	sessionID := uuid.NewV4()

	_ = sessionID

	go io.Copy(p.Conn, p.Server)
	io.Copy(p.Server, p.Conn)
	return nil
}
