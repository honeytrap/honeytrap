package ftp

import (
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"sync"
)

// A data socket is used to send non-control data between the client and
// server.
type DataSocket interface {
	Host() string

	Port() int

	// the standard io.Reader interface
	Read(p []byte) (n int, err error)

	// the standard io.Writer interface
	Write(p []byte) (n int, err error)

	// the standard io.Closer interface
	Close() error
}

type ftpActiveSocket struct {
	conn *net.TCPConn
	host string
	port int
}

func newActiveSocket(remote string, port int, sessionid string) (DataSocket, error) {
	connectTo := net.JoinHostPort(remote, strconv.Itoa(port))

	log.Debugf("%s - Opening active data connection to %s ", sessionid, connectTo)

	raddr, err := net.ResolveTCPAddr("tcp", connectTo)

	if err != nil {
		log.Debugf("%s: %s", sessionid, err.Error())
		return nil, err
	}

	tcpConn, err := net.DialTCP("tcp", nil, raddr)

	if err != nil {
		log.Debug(sessionid, err.Error())
		return nil, err
	}

	socket := new(ftpActiveSocket)
	socket.conn = tcpConn
	socket.host = remote
	socket.port = port

	return socket, nil
}

func (socket *ftpActiveSocket) Host() string {
	return socket.host
}

func (socket *ftpActiveSocket) Port() int {
	return socket.port
}

func (socket *ftpActiveSocket) Read(p []byte) (n int, err error) {
	return socket.conn.Read(p)
}

func (socket *ftpActiveSocket) Write(p []byte) (n int, err error) {
	return socket.conn.Write(p)
}

func (socket *ftpActiveSocket) Close() error {
	return socket.conn.Close()
}

type ftpPassiveSocket struct {
	conn      net.Conn
	port      int
	host      string
	ingress   chan []byte
	egress    chan []byte
	wg        sync.WaitGroup
	err       error
	tlsConfig *tls.Config
}

func newPassiveSocket(host string, port int, sessionid string, tlsConfig *tls.Config) (DataSocket, error) {
	socket := &ftpPassiveSocket{
		host:      host,
		port:      port,
		tlsConfig: tlsConfig,
		ingress:   make(chan []byte),
		egress:    make(chan []byte),
	}

	if err := socket.GoListenAndServe(sessionid); err != nil {
		return nil, err
	}
	return socket, nil
}

func (socket *ftpPassiveSocket) Host() string {
	return socket.host
}

func (socket *ftpPassiveSocket) Port() int {
	return socket.port
}

func (socket *ftpPassiveSocket) Read(p []byte) (n int, err error) {
	if err := socket.waitForOpenSocket(); err != nil {
		return 0, err
	}
	return socket.conn.Read(p)
}

func (socket *ftpPassiveSocket) Write(p []byte) (n int, err error) {
	if err := socket.waitForOpenSocket(); err != nil {
		return 0, err
	}
	return socket.conn.Write(p)
}

func (socket *ftpPassiveSocket) Close() error {
	if socket.conn != nil {
		return socket.conn.Close()
	}
	return nil
}

func (socket *ftpPassiveSocket) GoListenAndServe(sessionid string) (err error) {
	laddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort("", strconv.Itoa(socket.port)))
	if err != nil {
		log.Debug(sessionid, err.Error())
		return
	}

	var listener net.Listener
	listener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Debug(sessionid, err.Error())
		return
	}

	add := listener.Addr()
	parts := strings.Split(add.String(), ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Debug(sessionid, err.Error())
		return
	}

	socket.port = port
	socket.wg.Add(1)

	if socket.tlsConfig != nil {
		listener = tls.NewListener(listener, socket.tlsConfig)
	}

	go func() {
		conn, err := listener.Accept()
		socket.wg.Done()
		if err != nil {
			socket.err = err
			return
		}
		socket.err = nil
		socket.conn = conn
	}()
	return nil
}

func (socket *ftpPassiveSocket) waitForOpenSocket() error {
	if socket.conn != nil {
		return nil
	}
	socket.wg.Wait()
	return socket.err
}
