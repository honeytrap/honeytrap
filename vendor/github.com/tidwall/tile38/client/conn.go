package client

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LiveJSON is the value returned when a connection goes "live".
const LiveJSON = `{"ok":true,"live":true}`

// MaxMessageSize is maximum accepted message size
const MaxMessageSize = 0x1FFFFFFF // 536,870,911 bytes

// Proto is the protocol value.
type Proto int

const (
	Native    Proto = 0 // native protocol
	Telnet    Proto = 1 // telnet protocol
	HTTP      Proto = 2 // http protocol
	WebSocket Proto = 3 // websocket protocol
)

// Conn represents a connection to a tile38 server.
type Conn struct {
	c        net.Conn
	rd       *bufio.Reader
	pool     *Pool
	detached bool
}

// Dial connects to a tile38 server.
func Dial(addr string) (*Conn, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Conn{c: c, rd: bufio.NewReader(c)}, nil
}

// DialTimeout connects to a tile38 server with a timeout.
func DialTimeout(addr string, timeout time.Duration) (*Conn, error) {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	return &Conn{c: c, rd: bufio.NewReader(c)}, nil
}

// Close will close a connection.
func (conn *Conn) Close() error {
	if conn.pool == nil {
		if !conn.detached {
			conn.Do("QUIT")
		}
		return conn.c.Close()
	}
	return conn.pool.put(conn)
}

// SetDeadline sets the connection deadline for reads and writes.
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.c.SetDeadline(t)
}

// SetDeadline sets the connection deadline for reads.
func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.c.SetReadDeadline(t)
}

// SetDeadline sets the connection deadline for writes.
func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.c.SetWriteDeadline(t)
}

// Do sends a command to the server and returns the received reply.
func (conn *Conn) Do(command string) ([]byte, error) {
	if err := WriteMessage(conn.c, []byte(command)); err != nil {
		conn.pool = nil
		return nil, err
	}
	message, _, _, err := ReadMessage(conn.rd, nil)
	if err != nil {
		conn.pool = nil
		return nil, err
	}
	if string(message) == LiveJSON {
		conn.pool = nil // detach from pool
	}
	return message, nil
}

// ReadMessage returns the next message. Used when reading live connections
func (conn *Conn) ReadMessage() (message []byte, err error) {
	message, _, _, err = readMessage(conn.c, conn.rd)
	if err != nil {
		conn.pool = nil
		return message, err
	}
	return message, nil
}

// Reader returns the underlying reader.
func (conn *Conn) Reader() io.Reader {
	conn.pool = nil      // Remove from the pool because once the reader is called
	conn.detached = true // we will assume that this connection is detached.
	return conn.rd
}

// WriteMessage write a message to an io.Writer
func WriteMessage(w io.Writer, message []byte) error {
	h := []byte("$" + strconv.FormatUint(uint64(len(message)), 10) + " ")
	b := make([]byte, len(h)+len(message)+2)
	copy(b, h)
	copy(b[len(h):], message)
	b[len(b)-2] = '\r'
	b[len(b)-1] = '\n'
	_, err := w.Write(b)
	return err
}

// WriteHTTP writes an http message to the connection and closes the connection.
func WriteHTTP(conn net.Conn, data []byte) error {
	var buf bytes.Buffer
	buf.WriteString("HTTP/1.1 200 OK\r\n")
	buf.WriteString("Content-Length: " + strconv.FormatInt(int64(len(data))+1, 10) + "\r\n")
	buf.WriteString("Content-Type: application/json\r\n")
	buf.WriteString("Connection: close\r\n")
	buf.WriteString("\r\n")
	buf.Write(data)
	buf.WriteByte('\n')
	_, err := conn.Write(buf.Bytes())
	return err
}

// WriteWebSocket writes a websocket message.
func WriteWebSocket(conn net.Conn, data []byte) error {
	var msg []byte
	buf := make([]byte, 10+len(data))
	buf[0] = 129 // FIN + TEXT
	if len(data) <= 125 {
		buf[1] = byte(len(data))
		copy(buf[2:], data)
		msg = buf[:2+len(data)]
	} else if len(data) <= 0xFFFF {
		buf[1] = 126
		binary.BigEndian.PutUint16(buf[2:], uint16(len(data)))
		copy(buf[4:], data)
		msg = buf[:4+len(data)]
	} else {
		buf[1] = 127
		binary.BigEndian.PutUint64(buf[2:], uint64(len(data)))
		copy(buf[10:], data)
		msg = buf[:10+len(data)]
	}
	_, err := conn.Write(msg)
	return err
}

// ReadMessage reads the next message from a bufio.Reader.
func readMessage(wr io.Writer, rd *bufio.Reader) (message []byte, proto Proto, auth string, err error) {
	h, err := rd.Peek(1)
	if err != nil {
		return nil, proto, auth, err
	}
	switch h[0] {
	case '$':
		return readProtoMessage(rd)
	}
	message, proto, err = readTelnetMessage(rd)
	if err != nil {
		return nil, proto, auth, err
	}
	if len(message) > 6 && string(message[len(message)-9:len(message)-2]) == " HTTP/1" {
		return readHTTPMessage(string(message), wr, rd)
	}
	return message, proto, auth, nil

}

// ReadMessage read the next message from a bufio Reader.
func ReadMessage(rd *bufio.Reader, wr io.Writer) (message []byte, proto Proto, auth string, err error) {
	return readMessage(wr, rd)
}

func readProtoMessage(rd *bufio.Reader) (message []byte, proto Proto, auth string, err error) {
	b, err := rd.ReadBytes(' ')
	if err != nil {
		return nil, Native, auth, err
	}
	if len(b) > 0 && b[0] != '$' {
		return nil, Native, auth, errors.New("not a proto message")
	}
	n, err := strconv.ParseUint(string(b[1:len(b)-1]), 10, 32)
	if err != nil {
		return nil, Native, auth, errors.New("invalid size")
	}
	if n > MaxMessageSize {
		return nil, Native, auth, errors.New("message too big")
	}
	b = make([]byte, int(n)+2)
	if _, err := io.ReadFull(rd, b); err != nil {
		return nil, Native, auth, err
	}
	if b[len(b)-2] != '\r' || b[len(b)-1] != '\n' {
		return nil, Native, auth, errors.New("expecting crlf suffix")
	}
	return b[:len(b)-2], Native, auth, nil
}

func readTelnetMessage(rd *bufio.Reader) (command []byte, proto Proto, err error) {
	line, err := rd.ReadBytes('\n')
	if err != nil {
		return nil, Telnet, err
	}
	if len(line) > 1 && line[len(line)-2] == '\r' {
		line = line[:len(line)-2]
	} else {
		line = line[:len(line)-1]
	}
	return line, Telnet, nil
}

func readHTTPMessage(line string, wr io.Writer, rd *bufio.Reader) (command []byte, proto Proto, auth string, err error) {
	proto = HTTP
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		err = errors.New("invalid HTTP request")
		return
	}
	method := parts[0]
	path := parts[1]
	if len(path) == 0 || path[0] != '/' {
		err = errors.New("invalid HTTP request")
		return
	}
	path, err = url.QueryUnescape(path[1:])
	if err != nil {
		err = errors.New("invalid HTTP request")
		return
	}
	if method != "GET" && method != "POST" {
		err = errors.New("invalid HTTP method")
		return
	}
	contentLength := 0
	websocket := false
	websocketVersion := 0
	websocketKey := ""
	for {
		var b []byte
		b, _, err = readTelnetMessage(rd) // read a header line
		if err != nil {
			return
		}
		header := string(b)
		if header == "" {
			break // end of headers
		}
		if header[0] == 'a' || header[0] == 'A' {
			if strings.HasPrefix(strings.ToLower(header), "authorization:") {
				auth = strings.TrimSpace(header[len("authorization:"):])
			}
		} else if header[0] == 'u' || header[0] == 'U' {
			if strings.HasPrefix(strings.ToLower(header), "upgrade:") && strings.ToLower(strings.TrimSpace(header[len("upgrade:"):])) == "websocket" {
				websocket = true
			}
		} else if header[0] == 's' || header[0] == 'S' {
			if strings.HasPrefix(strings.ToLower(header), "sec-websocket-version:") {
				var n uint64
				n, err = strconv.ParseUint(strings.TrimSpace(header[len("sec-websocket-version:"):]), 10, 64)
				if err != nil {
					return
				}
				websocketVersion = int(n)
			} else if strings.HasPrefix(strings.ToLower(header), "sec-websocket-key:") {
				websocketKey = strings.TrimSpace(header[len("sec-websocket-key:"):])
			}
		} else if header[0] == 'c' || header[0] == 'C' {
			if strings.HasPrefix(strings.ToLower(header), "content-length:") {
				var n uint64
				n, err = strconv.ParseUint(strings.TrimSpace(header[len("content-length:"):]), 10, 64)
				if err != nil {
					return
				}
				contentLength = int(n)
			}
		}
	}
	if websocket && websocketVersion >= 13 && websocketKey != "" {
		proto = WebSocket
		if wr == nil {
			err = errors.New("connection is nil")
			return
		}
		sum := sha1.Sum([]byte(websocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		accept := base64.StdEncoding.EncodeToString(sum[:])
		wshead := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n"
		if _, err = wr.Write([]byte(wshead)); err != nil {
			return
		}
	} else if contentLength > 0 {
		proto = HTTP
		buf := make([]byte, contentLength)
		if _, err = io.ReadFull(rd, buf); err != nil {
			return
		}
		path += string(buf)
	}
	command = []byte(path)
	return
}
