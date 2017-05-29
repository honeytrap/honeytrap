// +build ignore

package proxies

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/op/go-logging"

	config "github.com/honeytrap/honeytrap/config"
	director "github.com/honeytrap/honeytrap/director"
	pushers "github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/proxies"
)

var log = logging.MustGetLogger("honeytrap:proxy:smtp")

// ListenSMTP returns a new proxy handler for the smtp provider.
// TODO: Change amount of params.
// combine listensmtp, smtpforwarder
func ListenSMTP(address string,m *director.ContainerConnections, d director.Director, p *pushers.Pusher, e pushers.Channel, c *config.Config) (net.Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := tls.LoadX509KeyPair(
		c.Proxies.SMTP.TLS.Certificate,
		c.Proxies.SMTP.TLS.CertificateKey,
	)
	if err != nil {
		return nil, err
	}

	tlsconfig := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionSSL30,
		MaxVersion:   tls.VersionTLS12,
	}

	return &SMTPProxyListener{
		&ProxyListener{
			l,
			m,
			d,
			p,
			e,
		},
		&tlsconfig,
	}, nil
}

// SMTPProxyListener defines a listener for smtp connections.
type SMTPProxyListener struct {
	*ProxyListener
	tlsconfig *tls.Config
}

// NewSMTPForwarder returns a new instance of the SMTPForwarder.
func NewSMTPForwarder(c *config.Config) (Forwarder, error) {
	cert, err := tls.LoadX509KeyPair(
		c.Proxies.SMTP.TLS.Certificate,
		c.Proxies.SMTP.TLS.CertificateKey,
	)
	if err != nil {
		return nil, err
	}

	tlsconfig := tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionSSL30,
		MaxVersion:   tls.VersionTLS12,
	}

	return &SMTPForwarder{
		tlsconfig: &tlsconfig,
	}, nil
}

// SMTPForwarder defines a forwader for smtp connection.
type SMTPForwarder struct {
	tlsconfig *tls.Config
}

// Forwarder defines a function to fowarded to the provided ProxyConn.
func (sf *SMTPForwarder) Forward(pc *ProxyConn) Proxyer {
	return &SMTPProxyConn{
		ProxyConn: pc,
		tlsconfig: sf.tlsconfig,
		meta: map[string]interface {
		}{},
	}
}

// Fowarder defines a struct which contains the underline forwarding logic for
// a proxy connection.
type Forwarder interface {
	Forward(pc *ProxyConn) Proxyer
}

// Accept handles the accept calls for provided SMTPProxyListener.
func (l *SMTPProxyListener) Accept() (net.Conn, error) {
	// _ = smconf := l.config.Proxies.SMTP

	conn, err := l.ProxyListener.Accept()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &SMTPProxyConn{
		ProxyConn: conn.(*ProxyConn),
		tlsconfig: l.tlsconfig,
		meta: map[string]interface {
		}{},
	}, err

}

type (
	//SMTPProxyConn provides a low-level smtp processor of a net.Conn
	SMTPProxyConn struct {
		*ProxyConn
		text1     *textproto.Conn
		text2     *textproto.Conn
		msg       *Message
		tlsconfig *tls.Config
		sessionID uuid.UUID
		meta      map[string]interface{}
	}

	SMTPConnect struct {
		Date       time.Time              `json:"timestamp"`
		MessageID  string                 `json:"message_id"`
		Host       string                 `json:"host"`
		RemoteAddr string                 `json:"remote_addr"`
		Meta       map[string]interface{} `json:"meta"`
	}

	SMTPMessage struct {
		Date       time.Time           `json:"timestamp"`
		MessageID  string              `json:"message_id"`
		Host       string              `json:"host"`
		RemoteAddr string              `json:"remote_addr"`
		From       *mail.Address       `json:"from"`
		To         []*mail.Address     `json:"to"`
		Headers    map[string][]string `json:"headers"`
	}

	SMTPMessageBody struct {
		Date      time.Time `json:"timestamp"`
		MessageID string    `json:"message_id"`
		Body      string    `json:"body"`
	}

	SMTPMessagePart struct {
		Date      time.Time           `json:"timestamp"`
		MessageID string              `json:"message_id"`
		Headers   map[string][]string `json:"headers"`
		Body      string              `json:"body"`
	}

	Message struct {
		From *mail.Address
		To   []*mail.Address

		Header mail.Header
		Body   []byte
	}
)

func newMessage() *Message {
	return &Message{To: []*mail.Address{}}
}

// Read reads the giving data from the provided Reader.
func (m *Message) Read(r io.Reader) error {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}

	m.Header = msg.Header
	m.Body, err = ioutil.ReadAll(msg.Body)
	return err
}

var receiveChan chan mail.Message

type stateFn func() (stateFn, error)

func (s *SMTPProxyConn) startState() (stateFn, error) {
	if _, err := s.proxy(s.text1, s.text2); err != nil {
		return nil, err
	}

	return s.loopState, nil
}

func isCommand(line string, cmd string) bool {
	return strings.HasPrefix(strings.ToUpper(line), cmd)
}

func (s *SMTPProxyConn) dataState() (stateFn, error) {
	log.Info(">DATA")

	defer func() {
		log.Info("<DATA")
	}()

	// data
	buff := new(bytes.Buffer)
	for {
		line, err := s.proxy(s.text2, s.text1)
		if err != nil {
			return nil, err
		}

		if line == "." {
			break
		}

		buff.WriteString(line)
		buff.WriteString("\n")
	}

	line2, err := s.proxy(s.text1, s.text2)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(line2, "250 ") {
		return s.loopState, nil
	}

	// queued successfully
	// send
	log.Info("%#v", s.msg)

	messageID := uuid.NewV4()
	headers := map[string][]string{}

	defer func() {

		s.ProxyConn.Event.Send(proxies.DataReadEvent(s.ProxyConn, "SMTP:Message-Part",SMTPMessagePart{
			Date:       time.Now(),
			MessageID:  messageID.String(),
			RemoteAddr: s.ProxyConn.RemoteHost(),
			From:       s.msg.From,
			To:         s.msg.To,
			Headers:    headers,
		}))

	}()

	msg, err := mail.ReadMessage(buff)
	if err != nil {
		log.Error(err.Error())

		// if no mime email.
		s.ProxyConn.Event.Send(proxies.ConnectionReadErrorEvent(s.ProxyConn, err))

		return s.loopState, nil
	}

	headers = msg.Header

	contentType := msg.Header.Get("Content-Type")

	mediatype, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Error(err.Error())
	}

	log.Info(mediatype)

	pr := multipart.NewReader(msg.Body, params["boundary"])
	for {
		part, err := pr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			s.ProxyConn.Event.Send(proxies.ConnectionReadErrorEvent(s.ProxyConn, err))
			log.Error(err.Error())
		}

		log.Info("%#v", part.Header)

		body, _ := ioutil.ReadAll(part)

		// TODO: should actually fix this, not in memory
		defer func() {
			s.ProxyConn.Event.Send(proxies.DataReadEvent(s.ProxyConn, "SMTP:Message-Part",SMTPMessagePart{
				Date:      time.Now(),
				MessageID: messageID.String(),
				Headers:   part.Header,
				Body:      string(body),
			}))
		}()

	}

	return s.loopState, nil
}

func (s *SMTPProxyConn) loopState() (stateFn, error) {
	line, err := s.proxy(s.text2, s.text1)
	if err != nil {
		return nil, err
	}

	line2, err := s.proxy(s.text1, s.text2)
	if err != nil {
		return nil, err
	}

	// BYE
	if strings.HasPrefix(line2, "221 ") {
		return nil, nil
	}

	if isCommand(line, "MAIL FROM") {
		s.msg = newMessage()

		if !strings.HasPrefix(line2, "250 ") {
			return s.loopState, nil
		}

		from, err := mail.ParseAddress(line[10:])
		if err != nil {
			// could not parse, assume it is only email
			from = &mail.Address{Address: line[10:]}
		}

		s.msg.From = from
		return s.loopState, nil
	} else if isCommand(line, "RCPT TO") {
		if !strings.HasPrefix(line2, "250 ") {
			return s.loopState, nil
		}

		if strings.HasPrefix(line2, "250 ") {

			to, err := mail.ParseAddress(line[8:])
			if err != nil {
				to = &mail.Address{Address: line[8:]}
			}

			log.Info("RCPT: ", to.String())

			s.msg.To = append(s.msg.To, to)
		}

		return s.loopState, nil
	} else if isCommand(line, "DATA") {
		if !strings.HasPrefix(line2, "354 ") {
			return s.loopState, nil
		}
		return s.dataState, nil
	} else if isCommand(line, "STARTTLS") {
		if !strings.HasPrefix(line2, "220") {
			return s.loopState, nil
		}

		tlsServerConn := tls.Client(s.server, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err := tlsServerConn.Handshake(); err != nil {
			log.Errorf("Error during server handshake: %s", err.Error())
			return s.loopState, nil
		}

		s.server = tlsServerConn
		s.text2 = textproto.NewConn(s.server)

		log.Info("TLS server handshake successfull.")

		tlsConn := tls.Server(s.Conn, s.tlsconfig)
		if err := tlsConn.Handshake(); err != nil {
			log.Errorf("Error during client handshake: %s", err.Error())
			return s.loopState, nil
		}

		s.Conn = tlsConn
		s.text1 = textproto.NewConn(s.Conn)

		log.Info("TLS client handshake successfull.")

	} else if isCommand(line, "HELLO") {
		if !strings.HasPrefix(line2, "250") {
			return s.loopState, nil
		}

		domain, err := parseHelloArgument(line)
		if err == nil {
			s.meta["domain"] = domain
		}

		return s.helloState, nil
	} else if isCommand(line, "EHLO") {
		if !strings.HasPrefix(line2, "250") {
			return s.loopState, nil
		}

		domain, err := parseHelloArgument(line)
		if err == nil {
			s.meta["domain"] = domain
		}

		return s.helloState, nil
	}

	return s.loopState, nil
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

func (s *SMTPProxyConn) helloState() (stateFn, error) {

	for {
		line, err := s.proxy(s.text1, s.text2)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(line, "250 ") {
			break
		}
	}

	return s.loopState, nil
}

func (s *SMTPProxyConn) proxy(text1, text2 *textproto.Conn) (string, error) {
	line, _, err := text2.R.ReadLine()
	if err != nil {
		return string(line), err
	}

	text1.W.Write(line)
	text1.W.Write([]byte{0xa})
	return string(line), text1.W.Flush()
}

// Proxy initializes and proxy the internal connection details for each connection.
func (s *SMTPProxyConn) Proxy( /*c *ProxyConfig*/ ) error {
	s.sessionID = uuid.NewV4()
	s.text1 = textproto.NewConn(s.Conn)
	s.text2 = textproto.NewConn(s.server)

	var state stateFn = s.startState
	var err error

	for {
		if state, err = state(); err != nil {
			log.Info("Error: %s", err.Error())
			break
		} else if state == nil {
			break
		}
	}

	return nil
}
