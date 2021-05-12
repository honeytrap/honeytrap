// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package smtp

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/services"
)

const (
	loopTreshold = 100
	cmdSupported = "HELO EHLO STARTTLS RCPT DATA RSET MAIL QUIT HELP AUTH BDAT NOOP QUIT"
)

type conn struct {
	rwc    net.Conn
	Text   *textproto.Conn
	domain string
	msg    *Message
	server *Server
	rcv    chan string
	i      int
	authed bool
	evnt   chan event.Event
}

func (c *conn) newMessage() *Message {

	return &Message{
		Body:   &bytes.Buffer{},
		Buffer: &bytes.Buffer{},
	}
}

func (c *conn) RemoteAddr() net.Addr {
	return c.rwc.RemoteAddr()
}

func (c *conn) LocalAddr() net.Addr {
	return c.rwc.LocalAddr()
}

type stateFn func(c *conn) stateFn

func (c *conn) PrintfLine(format string, args ...interface{}) error {
	fmt.Printf("< ")
	fmt.Printf(format, args...)
	fmt.Println("")
	return c.Text.PrintfLine(format, args...)
}

func (c *conn) ReadLine() (string, error) {
	s, err := c.Text.ReadLine()
	if err != nil {
		return s, err
	}

	fmt.Printf("> ")
	fmt.Println(s)

	//send line to log channel
	c.rcv <- s

	return s, nil
}

func startState(c *conn) stateFn {
	c.PrintfLine("220 %s", c.server.Banner)
	return helloState
}

func unrecognizedState(c *conn) stateFn {
	c.PrintfLine("500 unrecognized command")
	return loopState
}

func errorState(format string, args ...interface{}) stateFn {
	msg := fmt.Sprintf(format, args...)
	return func(c *conn) stateFn {
		c.PrintfLine("500 %s", msg)
		return nil
	}
}

func outOfSequenceState() stateFn {
	return func(c *conn) stateFn {
		c.PrintfLine("503 command out of sequence")
		return nil
	}
}

func authPlainState(msg string) stateFn {
	if msg == "" {
		return func(c *conn) stateFn {
			c.PrintfLine("334 ")
			msg, err := c.ReadLine()
			if err != nil {
				log.Error("[authPlainState] error: %s", err.Error())
				return authErrState
			}
			return authPlainState(msg)
		}
	}

	upb, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		log.Error("[authPlainState] error: %s", err.Error())
		return authErrState
	}

	var up []string
	for _, bs := range bytes.Split(upb, []byte{0}) {
		up = append(up, string(bs))
	}
	if len(up) < 3 {
		return authErrState
	}

	return func(c *conn) stateFn {
		c.evnt <- event.New(
			services.EventOptions,
			event.Category("smtp"),
			event.Type("input"),
			event.SourceAddr(c.RemoteAddr()),
			event.DestinationAddr(c.LocalAddr()),
			event.Custom("smtp.auth-protocol", "plain"),
			event.Custom("smtp.user", up[1]),
			event.Custom("smtp.password", up[2]),
		)
		return authSuccesState
	}
}

func authLoginState(msg string) stateFn {
	if msg == "" {
		return func(c *conn) stateFn {
			c.PrintfLine("334 VXNlcm5hbWU6")
			u64, err := c.ReadLine()
			if err != nil {
				return authErrState
			}
			return authLoginState(u64)
		}
	}
	return func(c *conn) stateFn {
		u, err := base64.StdEncoding.DecodeString(msg)
		if err != nil {
			return authErrState
		}

		c.PrintfLine("334 UGFzc3dvcmQ6")
		p64, err := c.ReadLine()
		if err != nil {
			return authErrState
		}

		p, err := base64.StdEncoding.DecodeString(p64)
		if err != nil {
			return authErrState
		}

		c.evnt <- event.New(
			services.EventOptions,
			event.Category("smtp"),
			event.Type("input"),
			event.SourceAddr(c.RemoteAddr()),
			event.DestinationAddr(c.LocalAddr()),
			event.Custom("smtp.auth-protocol", "login"),
			event.Custom("smtp.user", string(u)),
			event.Custom("smtp.password", string(p)),
		)
		return authSuccesState
	}
}

func authCramMD5State() stateFn {
	return func(c *conn) stateFn {
		challenge := md5.New()
		challenge.Write([]byte(time.Now().Format(time.RFC3339)))

		cMsg := base64.StdEncoding.EncodeToString(challenge.Sum(nil))
		c.PrintfLine("334 %s", cMsg)

		line, err := c.ReadLine()
		if err != nil {
			return authErrState
		}

		msg, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			return authErrState
		}

		parts := bytes.Split(msg, []byte(" "))
		if len(parts) < 2 {
			return authErrState
		}

		c.evnt <- event.New(
			services.EventOptions,
			event.Category("smtp"),
			event.Type("input"),
			event.SourceAddr(c.RemoteAddr()),
			event.DestinationAddr(c.LocalAddr()),
			event.Custom("smtp.auth-protocol", "cram-md5"),
			event.Custom("smtp.user", string(parts[0])),
			event.Custom("smtp.secret", string(parts[1])),
		)
		return authSuccesState
	}
}

func authErrState(c *conn) stateFn {
	c.PrintfLine("454 4.7.0 Temporary authentication failure")
	return loopState
}

func authSuccesState(c *conn) stateFn {
	c.PrintfLine("235 2.7.0 Authentication Succeeded")
	c.authed = true
	return loopState
}

func alreadyAuthedState(c *conn) stateFn {
	c.PrintfLine("503 You are already authenticated")
	return loopState
}

func authState(msg string) stateFn {
	return func(c *conn) stateFn {
		if c.authed {
			return alreadyAuthedState
		}
		ps := strings.Split(msg, " ")
		if len(ps) < 2 {
			return unrecognizedState
		}
		switch strings.ToUpper(ps[1]) {
		case "PLAIN":
			if len(ps) < 3 {
				return authPlainState("")
			}
			return authPlainState(ps[2])
		case "LOGIN":
			if len(ps) < 3 {
				return authLoginState("")
			}
			return authLoginState(ps[2])
		case "CRAM-MD5":
			return authCramMD5State()
		default:
			return unrecognizedState
		}
	}
}

func isCommand(line string, cmd string) bool {
	return strings.HasPrefix(strings.ToUpper(line), cmd)
}

func mailFromState(c *conn) stateFn {
	line, err := c.ReadLine()
	if err != nil {
		return errorState("[mailFromState] %s", err.Error())
	}

	if line == "" {
		return loopState
	} else if isCommand(line, "RSET") {
		c.PrintfLine("250 Ok")

		c.msg = c.newMessage()
		return loopState
	} else if isCommand(line, "RCPT TO") {

		c.PrintfLine("250 Ok")
		return mailFromState
	} else if isCommand(line, "BDAT") {
		parts := strings.Split(line, " ")

		var count int64
		if count, err = strconv.ParseInt(parts[1], 10, 32); err != nil {
			return errorState("[bdat]: error %s", err)
		}

		if _, err = io.CopyN(c.msg.Buffer, c.Text.R, count); err != nil {
			return errorState("[bdat]: error %s", err)
		}

		last := (len(parts) == 3 && parts[2] == "LAST")
		if !last {
			c.PrintfLine("250 Ok")
			return mailFromState
		}

		hasher := sha1.New()
		if err := c.msg.Read(io.TeeReader(c.msg.Buffer, hasher)); err != nil {
			return errorState("[bdat]: error %s", err)
		}

		c.PrintfLine("250 Ok : queued as +%x", hasher.Sum(nil))

		serverHandler{c.server}.Serve(*c.msg)

		c.msg = c.newMessage()
		return loopState
	} else if isCommand(line, "DATA") {
		c.PrintfLine("354 Enter message, ending with \".\" on a line by itself")

		hasher := sha1.New()
		err := c.msg.Read(io.TeeReader(c.Text.DotReader(), hasher))
		if err != nil {
			return errorState("[data]: error %s", err)
		}

		c.PrintfLine("250 Ok : queued as +%x", hasher.Sum(nil))

		serverHandler{c.server}.Serve(*c.msg)

		c.msg = c.newMessage()
		return loopState
	} else if isCommand(line, "HELP") {
		c.PrintfLine("214 Following SMTP commands are supported:")
		c.PrintfLine("214 %s", cmdSupported)
		return mailFromState
	}

	return unrecognizedState
}

func loopState(c *conn) stateFn {
	line, err := c.ReadLine()
	if err != nil {
		return errorState("[loopState] %s", err.Error())
	}

	if line == "" {
		return loopState
	}

	c.i++

	if c.i > loopTreshold {
		return errorState("[loopState] error: exceeded server loop treshold > %d", loopTreshold)
	}

	if isCommand(line, "MAIL FROM") {
		//c.msg.From, _ = mail.ParseAddress(line[10:])
		c.PrintfLine("250 Ok")
		return mailFromState
	} else if isCommand(line, "STARTTLS") {
		c.PrintfLine("220 Ready to start TLS")

		if c.server.tlsConfig == nil {
			c.PrintfLine("500 5.3.3. Unrecognized Command.")
			return helloState
		}

		tlsConn := tls.Server(c.rwc, c.server.tlsConfig)

		if err := tlsConn.Handshake(); err != nil {
			log.Error("Error during tls handshake: %s", err.Error())
			return nil
		}

		c.Text = textproto.NewConn(tlsConn)
		return helloState
	} else if isCommand(line, "AUTH") {
		if c.authed {
			return outOfSequenceState()
		}
		return authState(line)
	} else if isCommand(line, "RSET") {
		c.msg = c.newMessage()
		c.PrintfLine("250 Ok")
		return loopState
	} else if isCommand(line, "HELP") {
		c.PrintfLine("214 Following SMTP commands are supported:")
		c.PrintfLine("214 %s", cmdSupported)
		return loopState
	} else if isCommand(line, "QUIT") {
		c.PrintfLine("221 Bye")
		return nil
	} else if isCommand(line, "NOOP") {
		c.PrintfLine("250 Ok")
		return loopState
	} else if strings.Trim(line, " \r\n") == "" {
		return loopState
	}

	return unrecognizedState
}

func parseHelloArgument(arg string) (string, error) {
	domain := arg
	if idx := strings.IndexRune(arg, ' '); idx >= 0 {
		domain = arg[idx+1:]
	}
	if domain == "" {
		return "", fmt.Errorf("Invalid domain")
	}
	return domain, nil
}

func helloState(c *conn) stateFn {
	line, err := c.ReadLine()
	if err != nil {
		return errorState("[helloState] ReadLine error. %s", err.Error())
	}

	if isCommand(line, "HELO") {
		domain, err := parseHelloArgument(line)
		if err != nil {
			return errorState(err.Error())
		}

		c.domain = domain

		c.PrintfLine("250 Hello %s, I am glad to meet you", domain)
		return loopState
	} else if isCommand(line, "EHLO") {
		domain, err := parseHelloArgument(line)
		if err != nil {
			return errorState(err.Error())
		}

		c.domain = domain

		c.PrintfLine("250-Hello %s", domain)
		c.PrintfLine("250-SIZE 35882577")
		c.PrintfLine("250-8BITMIME")

		if c.server.tlsConfig != nil {
			c.PrintfLine("250-STARTTLS")
		}

		c.PrintfLine("250-HELP")
		c.PrintfLine("250-ENHANCEDSTATUSCODES")
		c.PrintfLine("250-PIPELINING")
		c.PrintfLine("250-CHUNKING")
		c.PrintfLine("250-AUTH PLAIN LOGIN CRAM-MD5")
		c.PrintfLine("250 SMTPUTF8")
		return loopState
	} else if isCommand(line, "HELP") {
		c.PrintfLine("214 This server supports the following commands")
		c.PrintfLine("214 HELO EHLO STARTTLS RCPT DATA RSET MAIL QUIT HELP AUTH DATA BDAT")
		return helloState
	}

	return errorState("Before we shake hands it will be appropriate to tell me who you are.")
}

func (c *conn) serve() {
	c.Text = textproto.NewConn(c.rwc)
	defer c.Text.Close()

	for state := startState; state != nil; {
		state = state(c)
	}
}
