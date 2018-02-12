package ftp

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	mrand "math/rand"
)

const (
	defaultWelcomeMessage = "Welcome to the Go FTP Server"
)

type Conn struct {
	conn          net.Conn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer
	dataConn      DataSocket
	driver        Driver
	auth          Auth
	server        *Server
	tlsConfig     *tls.Config
	sessionid     string
	namePrefix    string
	reqUser       string
	user          string
	password      string
	renameFrom    string
	lastFilePos   int64
	appendData    bool
	closed        bool
	tls           bool
	rcv           chan string
}

func (conn *Conn) LoginUser() string {
	return conn.user
}

func (conn *Conn) IsLogin() bool {
	return len(conn.user) > 0
}

func (conn *Conn) PublicIp() string {
	return conn.server.PublicIp
}

func (conn *Conn) passiveListenIP() string {
	if len(conn.PublicIp()) > 0 {
		return conn.PublicIp()
	}
	return conn.conn.LocalAddr().(*net.TCPAddr).IP.String()
}

func (conn *Conn) PassivePort() int {
	if len(conn.server.PassivePorts) > 0 {
		portRange := strings.Split(conn.server.PassivePorts, "-")

		if len(portRange) != 2 {
			log.Debug("empty port")
			return 0
		}

		minPort, _ := strconv.Atoi(strings.TrimSpace(portRange[0]))
		maxPort, _ := strconv.Atoi(strings.TrimSpace(portRange[1]))

		port := minPort + mrand.Intn(maxPort-minPort)
		return port
	}
	// let system automatically chose one port
	return 0
}

// returns a random 20 char string that can be used as a unique session ID
func newSessionID() string {
	hash := sha256.New()
	_, err := io.CopyN(hash, rand.Reader, 50)
	if err != nil {
		return "????????????????????"
	}
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[0:20]
}

// Serve starts an endless loop that reads FTP commands from the client and
// responds appropriately. terminated is a channel that will receive a true
// message when the connection closes. This loop will be running inside a
// goroutine, so use this channel to be notified when the connection can be
// cleaned up.
func (conn *Conn) Serve() {
	log.Debugf("%s: Connection Established", conn.sessionid)
	// send welcome
	conn.writeMessage(220, conn.server.WelcomeMessage)
	// read commands
	for {
		line, err := conn.controlReader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Errorf("%v Error: %s", conn.sessionid, err.Error())
			}

			break
		}
		conn.receiveLine(line)
		// QUIT command closes connection, break to avoid error on reading from
		// closed socket
		if conn.closed == true {
			break
		}
	}
	conn.Close()
	log.Debugf("%s: Connection Terminated", conn.sessionid)
}

// Close will manually close this connection, even if the client isn't ready.
func (conn *Conn) Close() {
	//send quit message
	//conn.rcv <- "q"

	conn.conn.Close()
	conn.closed = true
	if conn.dataConn != nil {
		conn.dataConn.Close()
		conn.dataConn = nil
	}
}

func (conn *Conn) upgradeToTLS() error {
	log.Debugf("%s: Upgrading connection to TLS", conn.sessionid)
	tlsConn := tls.Server(conn.conn, conn.tlsConfig)
	err := tlsConn.Handshake()
	if err == nil {
		conn.conn = tlsConn
		conn.controlReader = bufio.NewReader(tlsConn)
		conn.controlWriter = bufio.NewWriter(tlsConn)
		conn.tls = true
	}
	return err
}

// receiveLine accepts a single line FTP command and co-ordinates an
// appropriate response.
func (conn *Conn) receiveLine(line string) {

	//Log received commands in honeytrap log
	conn.rcv <- line

	command, param := conn.parseLine(line)
	cmdObj := commands[strings.ToUpper(command)]
	if cmdObj == nil {
		conn.writeMessage(500, "")
		return
	}
	if cmdObj.RequireParam() && param == "" {
		conn.writeMessage(553, "action aborted, required param missing")
	} else if cmdObj.RequireAuth() && conn.user == "" {
		conn.writeMessage(530, "")
	} else {
		cmdObj.Execute(conn, param)
	}
}

func (conn *Conn) parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}

// writeMessage will send a standard FTP response back to the client.
func (conn *Conn) writeMessage(code int, message string) (wrote int, err error) {
	if message == "" {
		message = statusText[code]
	}
	line := fmt.Sprintf("%d %s\r\n", code, message)
	wrote, err = conn.controlWriter.WriteString(line)
	conn.controlWriter.Flush()
	return
}

// writeMessage will send a standard FTP response back to the client.
func (conn *Conn) writeMessageMultiline(code int, message string) (wrote int, err error) {
	line := fmt.Sprintf("%d-%s\r\n%d END\r\n", code, message, code)
	wrote, err = conn.controlWriter.WriteString(line)
	conn.controlWriter.Flush()
	return
}

// sendOutofbandData will send a string to the client via the currently open
// data socket. Assumes the socket is open and ready to be used.
func (conn *Conn) sendOutofbandData(data []byte) {
	bytes := len(data)
	if conn.dataConn != nil {
		conn.dataConn.Write(data)
		conn.dataConn.Close()
		conn.dataConn = nil
	}
	message := "Closing data connection, sent " + strconv.Itoa(bytes) + " bytes"
	conn.writeMessage(226, message)
}

func (conn *Conn) sendOutofBandDataWriter(data io.ReadCloser) error {
	conn.lastFilePos = 0
	bytes, err := io.Copy(conn.dataConn, data)
	if err != nil {
		conn.dataConn.Close()
		conn.dataConn = nil
		return err
	}
	message := "Closing data connection, sent " + strconv.Itoa(int(bytes)) + " bytes"
	conn.writeMessage(226, message)
	conn.dataConn.Close()
	conn.dataConn = nil

	return nil
}
