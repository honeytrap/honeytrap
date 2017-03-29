package server

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/tidwall/resp"
)

const telnetIsJSON = false

// Type is resp type
type Type int

const (
	Null Type = iota
	RESP
	Telnet
	Native
	HTTP
	WebSocket
	JSON
)

// String return a string for type.
func (t Type) String() string {
	switch t {
	default:
		return "Unknown"
	case Null:
		return "Null"
	case RESP:
		return "RESP"
	case Telnet:
		return "Telnet"
	case Native:
		return "Native"
	case HTTP:
		return "HTTP"
	case WebSocket:
		return "WebSocket"
	case JSON:
		return "JSON"
	}
}

type errRESPProtocolError struct {
	msg string
}

func (err errRESPProtocolError) Error() string {
	return "Protocol error: " + err.msg
}

// Message is a resp message
type Message struct {
	Command    string
	Values     []resp.Value
	ConnType   Type
	OutputType Type
	Auth       string
}

// AnyReaderWriter is resp or native reader writer.
type AnyReaderWriter struct {
	rd *bufio.Reader
	wr io.Writer
	ws bool
}

// NewAnyReaderWriter returns an AnyReaderWriter object.
func NewAnyReaderWriter(rd io.Reader) *AnyReaderWriter {
	ar := &AnyReaderWriter{}
	if rd2, ok := rd.(*bufio.Reader); ok {
		ar.rd = rd2
	} else {
		ar.rd = bufio.NewReader(rd)
	}
	if wr, ok := rd.(io.Writer); ok {
		ar.wr = wr
	}
	return ar
}

func (ar *AnyReaderWriter) peekcrlfline() (string, error) {
	// this is slow operation.
	for i := 0; ; i++ {
		bb, err := ar.rd.Peek(i)
		if err != nil {
			return "", err
		}
		if len(bb) > 2 && bb[len(bb)-2] == '\r' && bb[len(bb)-1] == '\n' {
			return string(bb[:len(bb)-2]), nil
		}
	}
}

func (ar *AnyReaderWriter) readcrlfline() (string, error) {
	var line []byte
	for {
		bb, err := ar.rd.ReadBytes('\r')
		if err != nil {
			return "", err
		}
		if line == nil {
			line = bb
		} else {
			line = append(line, bb...)
		}
		b, err := ar.rd.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '\n' {
			return string(line[:len(line)-1]), nil
		}
		line = append(line, b)
	}
}

// ReadMessage reads the next resp message.
func (ar *AnyReaderWriter) ReadMessage() (*Message, error) {
	b, err := ar.rd.ReadByte()
	if err != nil {
		return nil, err
	}
	if err := ar.rd.UnreadByte(); err != nil {
		return nil, err
	}
	switch b {
	case 'G', 'P':
		line, err := ar.peekcrlfline()
		if err != nil {
			return nil, err
		}
		if len(line) > 9 && line[len(line)-9:len(line)-3] == " HTTP/" {
			return ar.readHTTPMessage()
		}
	case '$':
		return ar.readNativeMessage()
	}
	// MultiBulk also reads telnet
	return ar.readMultiBulkMessage()
}

func readNativeMessageLine(line []byte) (*Message, error) {
	values := make([]resp.Value, 0, 16)
reading:
	for len(line) != 0 {
		if line[0] == '{' {
			// The native protocol cannot understand json boundaries so it assumes that
			// a json element must be at the end of the line.
			values = append(values, resp.StringValue(string(line)))
			break
		}
		if line[0] == '"' && line[len(line)-1] == '"' {
			if len(values) > 0 &&
				strings.ToLower(values[0].String()) == "set" &&
				strings.ToLower(values[len(values)-1].String()) == "string" {
				// Setting a string value that is contained inside double quotes.
				// This is only because of the boundary issues of the native protocol.
				values = append(values, resp.StringValue(string(line[1:len(line)-1])))
				break
			}
		}
		i := 0
		for ; i < len(line); i++ {
			if line[i] == ' ' {
				value := string(line[:i])
				if value != "" {
					values = append(values, resp.StringValue(value))
				}
				line = line[i+1:]
				continue reading
			}
		}
		values = append(values, resp.StringValue(string(line)))
		break
	}
	return &Message{Command: commandValues(values), Values: values, ConnType: Native, OutputType: JSON}, nil
}

func (ar *AnyReaderWriter) readNativeMessage() (*Message, error) {
	b, err := ar.rd.ReadBytes(' ')
	if err != nil {
		return nil, err
	}
	if len(b) > 0 && b[0] != '$' {
		return nil, errors.New("invalid message")
	}
	n, err := strconv.ParseUint(string(b[1:len(b)-1]), 10, 32)
	if err != nil {
		return nil, errors.New("invalid size")
	}
	if n > 0x1FFFFFFF { // 536,870,911 bytes
		return nil, errors.New("message too big")
	}
	b = make([]byte, int(n)+2)
	if _, err := io.ReadFull(ar.rd, b); err != nil {
		return nil, err
	}
	if b[len(b)-2] != '\r' || b[len(b)-1] != '\n' {
		return nil, errors.New("expecting crlf")
	}

	return readNativeMessageLine(b[:len(b)-2])
}

func commandValues(values []resp.Value) string {
	if len(values) == 0 {
		return ""
	}
	return strings.ToLower(values[0].String())
}

func (ar *AnyReaderWriter) readMultiBulkMessage() (*Message, error) {
	rd := resp.NewReader(ar.rd)
	v, telnet, _, err := rd.ReadMultiBulk()
	if err != nil {
		return nil, err
	}
	values := v.Array()
	if len(values) == 0 {
		return nil, nil
	}
	if telnet && telnetIsJSON {
		return &Message{Command: commandValues(values), Values: values, ConnType: Telnet, OutputType: JSON}, nil
	}
	return &Message{Command: commandValues(values), Values: values, ConnType: RESP, OutputType: RESP}, nil

}

func (ar *AnyReaderWriter) readHTTPMessage() (*Message, error) {
	msg := &Message{ConnType: HTTP, OutputType: JSON}
	line, err := ar.readcrlfline()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, errors.New("invalid HTTP request")
	}
	method := parts[0]
	path := parts[1]
	if len(path) == 0 || path[0] != '/' {
		return nil, errors.New("invalid HTTP request")
	}
	path, err = url.QueryUnescape(path[1:])
	if err != nil {
		return nil, errors.New("invalid HTTP request")
	}
	if method != "GET" && method != "POST" {
		return nil, errors.New("invalid HTTP method")
	}
	contentLength := 0
	websocket := false
	websocketVersion := 0
	websocketKey := ""
	for {
		header, err := ar.readcrlfline()
		if err != nil {
			return nil, err
		}
		if header == "" {
			break // end of headers
		}
		if header[0] == 'a' || header[0] == 'A' {
			if strings.HasPrefix(strings.ToLower(header), "authorization:") {
				msg.Auth = strings.TrimSpace(header[len("authorization:"):])
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
					return nil, err
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
					return nil, err
				}
				contentLength = int(n)
			}
		}
	}
	if websocket && websocketVersion >= 13 && websocketKey != "" {
		msg.ConnType = WebSocket
		if ar.wr == nil {
			return nil, errors.New("connection is nil")
		}
		sum := sha1.Sum([]byte(websocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		accept := base64.StdEncoding.EncodeToString(sum[:])
		wshead := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n"
		if _, err = ar.wr.Write([]byte(wshead)); err != nil {
			return nil, err
		}
		ar.ws = true
	} else if contentLength > 0 {
		msg.ConnType = HTTP
		buf := make([]byte, contentLength)
		if _, err = io.ReadFull(ar.rd, buf); err != nil {
			return nil, err
		}
		path += string(buf)
	}
	if path == "" {
		return msg, nil
	}
	nmsg, err := readNativeMessageLine([]byte(path))
	if err != nil {
		return nil, err
	}
	msg.OutputType = JSON
	msg.Values = nmsg.Values
	msg.Command = commandValues(nmsg.Values)
	return msg, nil
}
