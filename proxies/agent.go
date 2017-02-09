// +build ignore

package proxies

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/satori/go.uuid"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	protocol "github.com/honeytrap/honeytrap/protocol"
	pushers "github.com/honeytrap/honeytrap/pushers"
	utils "github.com/honeytrap/honeytrap/utils"

	"github.com/golang/protobuf/proto"
)

const (
	MessageTypeForward = 0x1
	MessageTypePing    = 0x2
)

var ErrAgentUnsupportedProtocol = fmt.Errorf("Unsupported agent protocol")

func NewAgentServer(director *director.Director, pusher *pushers.Pusher, cfg *config.Config) *AgentServer {
	return &AgentServer{director, pusher, cfg}
}

type AgentServer struct {
	director *director.Director
	pusher   *pushers.Pusher
	config   *config.Config
}

type AgentConn struct {
	net.Conn
	remoteAddr string
	localAddr  string
	token      string
	as         *AgentServer
}

type AgentPing struct {
	Date      time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	LocalAddr string    `json:"local_addr"`
	Token     string    `json:"token"`
}

type AgentRequest struct {
	Date       time.Time `json:"timestamp"`
	Host       string    `json:"host"`
	RemoteAddr string    `json:"remote_addr"`
	LocalAddr  string    `json:"local_addr"`
	Protocol   string    `json:"protocol"`
	Token      string    `json:"token"`
}

func (c *AgentConn) RemoteAddr() net.Addr {
	// ac.remoteAddr
	addr, port, _ := net.SplitHostPort(c.remoteAddr)

	if value, err := strconv.Atoi(port); err != nil {
		return &net.TCPAddr{net.ParseIP(addr), value, ""}
	}

	return &net.TCPAddr{net.ParseIP(addr), 0, ""}
}

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

	ac.as.pusher.Push("agent", "ping", "", "", AgentPing{
		Date:      time.Now(),
		Host:      ac.Conn.RemoteAddr().String(),
		LocalAddr: *msg.LocalAddress,
		Token:     *msg.Token,
	})

	return nil
}

func (ac *AgentConn) Forward() error {
	length := int32(0)
	binary.Read(ac, binary.LittleEndian, &length)

	// add type, for health and forwarder

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

	container, err := ac.as.director.GetContainer(ac)
	if err != nil {
		return err
	}

	sessionID := uuid.NewV4()

	ac.as.pusher.Push("agent", "request", container.Name(), sessionID.String(), AgentRequest{
		Date:       time.Now(),
		LocalAddr:  ac.localAddr,
		RemoteAddr: ac.remoteAddr,
		Host:       ac.Conn.RemoteAddr().String(),
		Protocol:   *payload.Protocol,
		Token:      ac.token,
	})

	log.Debug("Received Agent connection from: %s with token: %s", ac.remoteAddr, ac.token)

	// TODO: make configurable
	var dport string
	switch {
	case *payload.Protocol == "ssh":
		dport = "22"
	case *payload.Protocol == "http":
		dport = "80"
	case *payload.Protocol == "smtp":
		dport = "25"
	default:
		log.Error("Unsupported agent protocol: %s", *payload.Protocol)
		return ErrAgentUnsupportedProtocol
	}

	log.Debug("Agent forwarding protocol: %s(%s) %s", *payload.RemoteAddress, dport, *payload.Protocol)

	var c2 net.Conn
	c2, err = container.Dial(dport)
	if err != nil {
		return err
	}

	defer c2.Close()

	pc := ProxyConn{ac, c2, container, ac.as.pusher}

	var sp Proxyer
	switch {
	case *payload.Protocol == "ssh":
		sp = &SSHProxyConn{&pc, &ac.as.config.Proxies.SSH}
	case *payload.Protocol == "http":
		sp = &HTTPProxyConn{&pc}
	case *payload.Protocol == "smtp":
		forwarder, err := NewSMTPForwarder(ac.as.config)
		if err != nil {
			return err
		}

		sp = forwarder.Forward(&pc)
	default:
		log.Error("Unsupported agent protocol: %s", *payload.Protocol)
		return ErrAgentUnsupportedProtocol
	}

	return sp.Proxy()
}

func (ac *AgentConn) Serve() error {
	defer ac.Close()

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
		return fmt.Errorf("Unknown message type.")
	}
}

func (as *AgentServer) newConn(conn net.Conn) *AgentConn {
	return &AgentConn{Conn: conn, as: as}
}

func (as AgentServer) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error(err.Error())
			continue
		}

		ac := as.newConn(conn)
		go func() {
			defer utils.RecoverHandler()

			ac.Serve()
		}()
	}
}

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
