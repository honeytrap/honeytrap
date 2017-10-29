/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package services

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("smtp", SMTP)
)

type Message struct {
	From *mail.Address
	To   []*mail.Address

	Domain     string
	RemoteAddr string

	Header mail.Header
	Buffer *bytes.Buffer
	Body   *bytes.Buffer
}

func (c *smtpconn) newMessage() *Message {

	return &Message{
		To:         []*mail.Address{},
		Body:       &bytes.Buffer{},
		Buffer:     &bytes.Buffer{},
		Domain:     c.domain,
		RemoteAddr: c.RemoteAddr().String(),
	}
}

func (m *Message) Read(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	msg, err := mail.ReadMessage(bytes.NewReader(buff))
	if err != nil {
		m.Body = bytes.NewBuffer(buff)
		return err
	}

	m.Header = msg.Header

	buff, err = ioutil.ReadAll(msg.Body)
	if err != nil {
		return err
	}

	m.Body = bytes.NewBuffer(buff)
	return err
}

func (c *smtpconn) RemoteAddr() net.Addr {
	return c.rwc.RemoteAddr()
}

type stateFn func(c *smtpconn) stateFn

func (c *smtpconn) PrintfLine(format string, args ...interface{}) error {
	fmt.Printf("< ")
	fmt.Printf(format, args...)
	fmt.Println("")
	return c.Text.PrintfLine(format, args...)
}

func (c *smtpconn) ReadLine() (string, error) {
	s, err := c.Text.ReadLine()
	fmt.Printf("> ")
	fmt.Println(s)
	return s, err
}

func startState(c *smtpconn) stateFn {
	c.PrintfLine("220 %s", c.banner)
	return helloState
}

func unrecognizedState(c *smtpconn) stateFn {
	c.PrintfLine("500 unrecognized command")
	return loopState
}

func errorState(format string, args ...interface{}) stateFn {
	msg := fmt.Sprintf(format, args...)
	return func(c *smtpconn) stateFn {
		c.PrintfLine("500 %s", msg)
		return nil
	}
}

func outOfSequenceState() stateFn {
	return func(c *smtpconn) stateFn {
		c.PrintfLine("503 command out of sequence")
		return nil
	}
}

func isCommand(line string, cmd string) bool {
	return strings.HasPrefix(strings.ToUpper(line), cmd)
}

func mailFromState(c *smtpconn) stateFn {
	line, err := c.ReadLine()
	if err != nil {
		log.Error("[mailFromState] error: %s", err.Error())
		return nil
	}

	if line == "" {
		return loopState
	}

	if isCommand(line, "RSET") {
		c.PrintfLine("250 Ok")

		c.msg = c.newMessage()
		return loopState
	} else if isCommand(line, "RCPT TO") {
		addr, _ := mail.ParseAddress(line[8:])

		c.msg.To = append(c.msg.To, addr)

		c.PrintfLine("250 Ok")
		return mailFromState
	} else if isCommand(line, "BDAT") {
		parts := strings.Split(line, " ")

		count := int64(0)
		var err error
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

		c.msg = c.newMessage()
		return loopState
	} else {
		return unrecognizedState
	}
}

func loopState(c *smtpconn) stateFn {
	line, err := c.ReadLine()
	if err != nil {
		log.Error("[loopState] error: %s", err.Error())
		return nil
	}

	if line == "" {
		return loopState
	}

	c.i++

	if c.i > 100 {
		return errorState("Error: invalid.")
	}

	if isCommand(line, "MAIL FROM") {
		c.msg.From, _ = mail.ParseAddress(line[10:])
		c.PrintfLine("250 Ok")
		return mailFromState
	} else if isCommand(line, "STARTTLS") {
		c.PrintfLine("220 Ready to start TLS")

		tlsconn := tls.Server(c.rwc, c.TLSConfig)

		if err := tlsconn.Handshake(); err != nil {
			log.Error("Error during tls handshake: %s", err.Error())
			return nil
		}

		c.Text = textproto.NewConn(tlsconn)
		return helloState
	} else if isCommand(line, "RSET") {
		c.msg = c.newMessage()
		c.PrintfLine("250 Ok")
		return loopState
	} else if isCommand(line, "QUIT") {
		c.PrintfLine("221 Bye")
		return nil
	} else if isCommand(line, "NOOP") {
		c.PrintfLine("250 Ok")
		return loopState
	} else if strings.Trim(line, " \r\n") == "" {
		return loopState
	} else {
		return unrecognizedState
	}
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

func helloState(c *smtpconn) stateFn {
	line, _ := c.ReadLine()

	if isCommand(line, "HELO") {
		domain, err := parseHelloArgument(line)
		if err != nil {
			return errorState(err.Error())
		}

		c.domain = domain
		c.msg.Domain = domain

		c.PrintfLine("250 Hello %s, I am glad to meet you", domain)
		return loopState
	} else if isCommand(line, "EHLO") {
		domain, err := parseHelloArgument(line)
		if err != nil {
			return errorState(err.Error())
		}

		c.domain = domain
		c.msg.Domain = domain

		c.PrintfLine("250-Hello %s", domain)
		c.PrintfLine("250-SIZE 35882577")
		c.PrintfLine("250-8BITMIME")
		c.PrintfLine("250-STARTTLS")
		c.PrintfLine("250-ENHANCEDSTATUSCODES")
		c.PrintfLine("250-PIPELINING")
		c.PrintfLine("250-CHUNKING")
		c.PrintfLine("250 SMTPUTF8")
		return loopState
	} else {
		return errorState("Before we shake hands it will be appropriate to tell me who you are.")
	}
}

func (c *smtpconn) serve() {
	c.Text = textproto.NewConn(c.rwc)
	defer c.Text.Close()

	// todo add idle timeout here

	for state := startState; state != nil; {
		state = state(c)
	}
}

type smtpconn struct {
	rwc       net.Conn
	TLSConfig *tls.Config
	Text      *textproto.Conn
	domain    string
	msg       *Message
	banner    string `toml:"banner"`
	i         int
}

// SMTP
func SMTP(options ...ServicerFunc) Servicer {

	s := &SMTPService{
		sconn: &smtpconn{
			TLSConfig: nil,
		},
	}

	for _, o := range options {
		o(s)
	}
	return s
}

type SMTPService struct {
	c     pushers.Channel
	sconn *smtpconn
}

func (s *SMTPService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *SMTPService) Handle(conn net.Conn) error {

	s.sconn.rwc = conn
	s.sconn.msg = s.sconn.newMessage()
	// Use a go routine here???
	s.sconn.serve()

	s.c.Send(event.New(
		EventOptions,
		event.Category("smtp"),
		event.Type("mail"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("smtp.From", s.sconn.msg.From),
		event.Custom("smtp.To", s.sconn.msg.To),
		event.Custom("smtp.Body", s.sconn.msg.Body.String()),
	))
	return nil
}
