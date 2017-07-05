// +build ignore

package proxies

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	protocol "github.com/honeytrap/honeytrap/protocol"
	pushers "github.com/honeytrap/honeytrap/pushers"
	utils "github.com/honeytrap/honeytrap/utils"
	uuid "github.com/satori/go.uuid"

	"github.com/golang/protobuf/proto"
)

// contains message type constants for different messages.
const (
	MessageTypeForward = 0x1
	MessageTypePing    = 0x2
)

// ErrAgentUnsupportedProtocol is returned when an Unsupported agent is seen.
var ErrAgentUnsupportedProtocol = fmt.Errorf("Unsupported agent protocol")

// NewAgentServer returns a new AgentServer instance.
func NewAgentServer(manager *director.ContainerConnections, director director.Director, events pushers.Channel, cfg *config.Config) *AgentServer {
	return &AgentServer{director, events, cfg, manager}
}

// AgentServer defines an a struct which implements a server to handle agent based
// connections.
type AgentServer struct {
	director director.Director

	events  pushers.Channel
	config  *config.Config
	manager *director.ContainerConnections
}

// AgentConn defines the a struct which holds the underline net.Conn.
type AgentConn struct {
	net.Conn
	remoteAddr string
	localAddr  string
	token      string
	as         *AgentServer
}

// AgentSession defines a struct which holds the giving session value of the
// provided current Agent.
type AgentSession struct {
	Session string `json:"session"`
}

// AgentPing defines a data struct to hold ping request/response data.
type AgentPing struct {
	Date      time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	LocalAddr string    `json:"local_addr"`
	Token     string    `json:"token"`
}

// AgentRequest defines a data struct to hold a giving request.
type AgentRequest struct {
	Date       time.Time `json:"timestamp"`
	Host       string    `json:"host"`
	RemoteAddr string    `json:"remote_addr"`
	LocalAddr  string    `json:"local_addr"`
	Protocol   string    `json:"protocol"`
	Token      string    `json:"token"`
}

// RemoteAddr returns the RemoteAddr of the underline net.Conn.
func (ac *AgentConn) RemoteAddr() net.Addr {
	// ac.remoteAddr
	addr, port, _ := net.SplitHostPort(ac.remoteAddr)

	if value, err := strconv.Atoi(port); err != nil {
		return &net.TCPAddr{
			IP:   net.ParseIP(addr),
			Port: value,
			Zone: "",
		}
	}

	return &net.TCPAddr{
		IP:   net.ParseIP(addr),
		Port: 0,
		Zone: "",
	}
}

// Ping delivers a Ping message to the underline agent conn.
func (ac *AgentConn) Ping() error {
	length := int32(0)
	binary.Read(ac, binary.LittleEndian, &length)

	data := make([]byte, length)
	ac.Read(data)

	msg := &protocol.PingMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		log.Error("Error unmarshalling: ", err.Error())
		return err
	}

	log.Debug("Received ping from agent: %s with token: %s", *msg.LocalAddress, *msg.Token)

	ac.as.events.Send(PingEvent(ac.Conn, AgentPing{
		Date:      time.Now(),
		Token:     *msg.Token,
		Host:      ac.Conn.RemoteAddr().String(),
		LocalAddr: *msg.LocalAddress,
	}))

	return nil
}

// Forward delivers a forward data request to the underline net.Conn.
func (ac *AgentConn) Forward() error {
	length := int32(0)
	binary.Read(ac, binary.LittleEndian, &length)

	// add type, for health and forwarder

	sessionID := uuid.NewV4()

	data := make([]byte, length)
	ac.Read(data)

	payload := &protocol.PayloadMessage{}
	if err := proto.Unmarshal(data, payload); err != nil {
		log.Error("Error unmarshalling: ", err.Error())
		return err
	}

	ac.localAddr = *payload.LocalAddress
	ac.remoteAddr = *payload.RemoteAddress
	ac.token = *payload.Token

	ac.as.events.Send(ServiceStartedEvent(ac.Conn, AgentRequest{
		Date:       time.Now(),
		LocalAddr:  ac.localAddr,
		RemoteAddr: ac.remoteAddr,
		Host:       ac.Conn.RemoteAddr().String(),
		Protocol:   *payload.Protocol,
		Token:      ac.token,
	}, nil))

	container, err := ac.as.director.GetContainer(ac)
	if err != nil {
		ac.as.events.Send(ServiceEndedEvent(ac.Conn, AgentRequest{
			Date:       time.Now(),
			LocalAddr:  ac.localAddr,
			RemoteAddr: ac.remoteAddr,
			Host:       ac.Conn.RemoteAddr().String(),
			Protocol:   *payload.Protocol,
			Token:      ac.token,
		}, nil))
		return err
	}

	ac.as.events.Send(AgentRequestEvent(ac.Conn, sessionID, AgentRequest{
		Date:       time.Now(),
		LocalAddr:  ac.localAddr,
		RemoteAddr: ac.remoteAddr,
		Host:       ac.Conn.RemoteAddr().String(),
		Protocol:   *payload.Protocol,
		Token:      ac.token,
	}, map[string]interface{}{
		"container": container.Detail(),
	}))

	log.Debugf("Received Agent connection from: %s with token: %s", ac.remoteAddr, ac.token)

	// TODO: make configurable
	// var dport string
	// switch {
	// case *payload.Protocol == "ssh":
	// 	dport = "22"
	// case *payload.Protocol == "http":
	// 	dport = "80"
	// case *payload.Protocol == "smtp":
	// 	dport = "25"
	// default:
	// 	log.Errorf("Unsupported agent protocol: %s", *payload.Protocol)

	// 	ac.as.events.Send(ServiceEndedEvent(ac.Conn, ErrAgentUnsupportedProtocol, nil))

	// 	return ErrAgentUnsupportedProtocol
	// }

	_, port, err := net.SplitHostPort(ac.Conn.LocalAddr().String())
	if err != nil {
		//TODO(alex): Decide what to do if this fails
	}

	log.Debugf("Agent forwarding protocol: %s(%s) %s", *payload.RemoteAddress, dport, *payload.Protocol)

	var c2 net.Conn
	c2, err = container.Dial(context.Background())
	if err != nil {
		ac.as.events.Send(ServiceEndedEvent(ac.Conn, err, nil))

		return err
	}

	defer c2.Close()

	pc := ProxyConn{ac, c2, container, ac.as.pusher, ac.as.events}

	ac.as.manager.AddClient(&pc, container.Detail())

	var sp Proxyer

	switch *payload.Protocol {
	case "ssh":
		sp = &SSHProxyConn{&pc, &ac.as.config.Proxies.SSH}
	case "http":
		sp = &HTTPProxyConn{&pc}
	case "smtp":
		forwarder, err := NewSMTPForwarder(ac.as.config)
		if err != nil {
			ac.as.events.Send(ServiceEndedEvent(ac.Conn, err, nil))
			return err
		}

		sp = forwarder.Forward(&pc)
	default:
		log.Errorf("Unsupported agent protocol: %s", *payload.Protocol)

		ac.as.events.Send(ServiceEndedEvent(ac.Conn, ErrAgentUnsupportedProtocol, nil))

		return ErrAgentUnsupportedProtocol
	}

	defer ac.as.events.Send(ServiceEndedEvent(ac.Conn, nil, nil))

	return sp.Proxy()
}

// Serve initializes the AgentConn and it's operations.
func (ac *AgentConn) Serve() error {
	defer ac.as.events.Send(ConnectionClosedEvent(ac.Conn))
	defer ac.Close()

	ac.as.events.Send(ConnectionOpenedEvent(ac.Conn))

	if ac.as.config.Agent.TLS.Enabled {
		ac.Conn = tls.Server(ac.Conn, &tls.Config{})
	}

	// TODO: add gzip support

	msgtype := int32(0)
	binary.Read(ac, binary.LittleEndian, &msgtype)

	switch msgtype {
	case MessageTypePing:
		return ac.Ping()
	case MessageTypeForward:
		return ac.Forward()
	default:
		return fmt.Errorf("unknown message type")
	}
}

// newConn initializes the AgentServer with a new net.Conn returning a new
// AgentConn.
func (as *AgentServer) newConn(conn net.Conn) *AgentConn {
	return &AgentConn{Conn: conn, as: as}
}

// Serve initializes the AgentServer and it's operations.
func (as AgentServer) Serve(l net.Listener) error {
	as.events.Send(ListenerOpenedEvent(l, nil, nil))

	defer as.events.Send(ListenerClosedEvent(l))

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err.Error())

			// TODO: Shouldnt we check when this gets closed? We dont want
			// endless loop running
			continue
		}

		ac := as.newConn(conn)
		go func() {
			defer utils.RecoverHandler()

			ac.Serve()
		}()
	}
}

// ListenAndServe initializes the AgentServer to begin serving requests and response.
func (as *AgentServer) ListenAndServe() error {
	log.Infof("Agent server Listening on port: %s", as.config.Agent.Port)

	l, err := net.Listen("tcp", as.config.Agent.Port)
	if err != nil {
		log.Fatal(err)
		return err
	}

	defer l.Close()

	return as.Serve(l)
}

//================================================================================
