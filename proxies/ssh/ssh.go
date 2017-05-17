package proxies

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	logging "github.com/op/go-logging"
	"github.com/satori/go.uuid"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/proxies"
	"github.com/honeytrap/honeytrap/pushers"

	"golang.org/x/crypto/ssh"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

var _ = proxies.Register("ssh", Listen)

var log = logging.MustGetLogger("honeytrap:proxy:ssh")

// contains different message header strings for ssh sensors.
var (
	SSHSensorTypeOutgoing = "Session-Outgoing-packet"
	SSHSensorTypeIncoming = "Session-Incoming-packet"
)

// Config defines the configuration passed in to create a ssh connection.
type Config struct {
	Key    *PrivateKey `toml:"key"`
	Port   string      `toml:"port"`
	Banner string      `toml:"banner"`
}

func generateKey() (*PrivateKey, error) {
	// TODO: cache generated key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Errorf("Could not create ssh key: %s", err.Error())
		return nil, err
	}

	if cerr := priv.Validate(); cerr != nil {
		log.Errorf("Validation failed: %s", cerr.Error())
		return nil, cerr
	}

	privder := x509.MarshalPKCS1PrivateKey(priv)

	privblk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privder,
	}

	privateBytes := pem.EncodeToMemory(&privblk)

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{private}, nil
}

// PrivateKey holds the ssh.Signer instance to unsign received data.
type PrivateKey struct {
	ssh.Signer
}

// UnmarshalText unmarshalls the giving text as the Signers data.
func (t *PrivateKey) UnmarshalText(data []byte) (err error) {
	keyFile := string(data)

	log.Debug("Loading ssh private key: %s ", keyFile)
	b, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Errorf("Could not open file:%s : %s", keyFile, err.Error())
		return err
	}

	private, err := ssh.ParsePrivateKey(b)
	if err != nil {
		log.Errorf("Validation failed: %s", err.Error())
		return err
	}

	(*t) = PrivateKey{private}
	return err
}

// Listen initializes the ssh connection processes.
// can we do something with:
// https://github.com/golang/crypto/blob/master/ssh/agent/forward.go
//
// we have toml with config
// so we can use this func with toml configurattion
// toml.PrimitiveDecode(primitive, &SSHConfig{})
// we don't need address anymore
//
// how to pass the pusher and director
// Maybe just have a Listener() function that will return the listener?
//
// and how will this fit in the newer solution?
//
// also we don't want the proxy listener to listen, but have a custom listener interface
// that we can swap. So we can use for emxapl cowrie as well, but also our raw stack
//
// TODO: Change amount of params.
func Listen(address string, d director.Director, p *pushers.Pusher, events pushers.Events, primitive toml.Primitive) (net.Listener, error) {
	c := Config{
		Key:    nil,
		Port:   ":8022",
		Banner: "SSH-2.0-OpenSSH_6.6.1p1 2020Ubuntu-2ubuntu2",
	}

	if err := toml.PrimitiveDecode(primitive, &c); err != nil {
		return nil, err
	}

	if c.Key != nil {
	} else if key, err := generateKey(); err != nil {
		return nil, fmt.Errorf("Could not generate key: %s", err.Error())
	} else {
		c.Key = key
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		proxies.NewProxyListener(l, d, p, events),
		&c,
	}, nil
}

// Listener defines a custom Listener for handling ssh connections.
type Listener struct {
	*proxies.ProxyListener

	c *Config
}

// Accept calls the internal accept method for the underline ProxyListener.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.ProxyListener.Accept()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &Conn{conn.(*proxies.ProxyConn), l.c}, err
}

// Conn defines a custom connection for handling ssh proxying.
type Conn struct {
	*proxies.ProxyConn

	c *Config
}

// Proxy setsup the needed recorders to record ssh connection details.
func (p *Conn) Proxy() error {
	recorder := NewSSHRecorder(p.Pusher, p.Event)

	rs := recorder.NewSession(p.ProxyConn)

	rs.Connect()

	c2Config := ssh.ClientConfig{}
	c2Config.SetDefaults()

	c2Conn := ssh.NewClientConn2(p.Server, &c2Config)
	defer c2Conn.Close()

	err := c2Conn.Handshake2(p.Server.RemoteAddr().String(), &c2Config)
	if err != nil {
		log.Errorf("Client handshake failed: %s.", err.Error())
		return err
	}

	username := ""
	password := ""

	// TODO: do we have an other option for this?
	config := ssh.ServerConfig{
		ServerVersion: p.c.Banner,
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			log.Infof("Publickey auth: %s %x", key.Type(), key.Marshal())
			rs.AuthorizationPublicKey(conn.User(), key.Type(), key.Marshal())
			return nil, errors.New("Unknown key")
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password2 []byte) (*ssh.Permissions, error) {
			c2Config.User = conn.User()
			c2Config.Auth = []ssh.AuthMethod{ssh.Password(string(password2))}
			c2Config.ClientVersion = string(conn.ClientVersion())

			username = conn.User()
			password = string(password2)

			cerr := c2Conn.Authorize2(&c2Config)
			if cerr == nil {
				return nil, cerr
			} else if cerr == io.EOF {
				// TODO: How to handle client c2conn close? Server conn should be closed as well
				// close server connection
				// defer p.Close()
			} else if _, ok := cerr.(syscall.Errno); ok {
				// TODO: How to handle client c2conn close? Server conn should be closed as well
				// close server connection
				// defer p.Close()
			}

			log.Errorf("Authorization failed (ip: %s, client: %s, username: %s, password: %s): %s", p.Server.RemoteAddr().String(), string(conn.ClientVersion()), username, password, err.Error())
			rs.AuthorizationFailed(conn.User(), string(password), string(conn.ClientVersion()))
			return nil, err
		},
	}

	config.AddHostKey(p.c.Key)

	serverConn, chans, reqs, err := ssh.NewServerConn(p.Conn, &config)
	if err == io.EOF {
		// server closed connection
		log.Error("Client closed connection.")
		return nil
	} else if err != nil {
		return (err)
	}

	go ssh.DiscardRequests(reqs)

	rs.AuthorizationSuccess(username, password, string(serverConn.ClientVersion()))
	log.Infof("Authorization succeeded (ip: %s, client: %s, username: %s, password: %s)", p.Server.RemoteAddr().String(), string(serverConn.ClientVersion()), username, password)

	_, _, err = c2Conn.Mux2()
	if err != nil {
		log.Error("failed to mux")
		return (err)
	}

	rs.Start() // SSH Connection succesfully instantiated
	defer rs.Stop()

	// todo:
	// register client software
	// select {
	// <- chans
	// <- c2Chans
	for newChannel := range chans {
		channel2, requests2, err2 := c2Conn.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
		if err2 != nil {
			log.Error("Could not accept client channel: ", err2)
			return err2
		}

		channel, requests, cerr := newChannel.Accept()
		if cerr != nil {
			log.Error("Could not accept server channel: ", cerr)
			return cerr
		}

		channelID := uuid.NewV4()

		// connect requests
		go func() {
			log.Debug("Waiting for request")

		r:
			for {
				var req *ssh.Request
				var dst ssh.Channel

				select {
				case req = <-requests:
					dst = channel2
				case req = <-requests2:
					dst = channel
				}

				if req == nil {
					log.Debug("Request is nil??? %s %s", username, password)
					break
				}

				b, cerr := dst.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					log.Error("Reply Error", cerr)
				}

				if req.WantReply {
					req.Reply(b, nil)
				}

				data := map[string]interface{}{"name": req.Type, "payload": req.Payload}

				pack, cerr := json.Marshal(data)
				if cerr != nil {
					log.Error("Unable to Marshal Payload Channel Request Type", cerr)
				} else {
					rs.Data("Session-RequestType-packet", channelID, pack)
				}

				switch req.Type {
				case "exit-status":
					break r
				case "pty-req":
					log.Info(string(req.Payload))
				case "shell":
				case "exec":
					// not supported (yet)
				default:
					log.Error(req.Type)
				}
			}

			channel.Close()
			channel2.Close()
		}()

		// TODO: use a conf or something for this.
		var wrappedChannel io.ReadCloser = NewSSHRecorderStream(SSHSensorTypeIncoming, rs, channelID, username, password, channel)
		var wrappedChannel2 io.ReadCloser = NewSSHRecorderStream(SSHSensorTypeOutgoing, rs, channelID, username, password, channel2)

		go io.Copy(channel2, wrappedChannel)
		go io.Copy(channel, wrappedChannel2)

		defer wrappedChannel.Close()
		defer wrappedChannel2.Close()
	}

	return err
}

// NewSSHRecorderStream returns a new ssh recorder stream.
func NewSSHRecorderStream(sensor string, rs *SSHRecorderSession /*meta, */, channelID uuid.UUID, username, password string, r io.ReadCloser) io.ReadCloser {
	return &SSHPacketRecorder{sensor: sensor, rs: rs, channelID: channelID, ReadCloser: r, username: username, password: password, time: time.Now()}
}

// SSHPacketRecorder defines a custom packet recorder for ssh connections
type SSHPacketRecorder struct {
	rs *SSHRecorderSession
	io.ReadCloser

	channelID uuid.UUID
	sensor    string
	username  string
	password  string
	time      time.Time
	buffer    bytes.Buffer
}

// Read reads the internal data from the underline reader.
func (lr *SSHPacketRecorder) Read(p []byte) (n int, err error) {
	n, err = lr.ReadCloser.Read(p)
	lr.rs.Data(lr.sensor, lr.channelID, p[:n])
	return n, err
}

// String returns the string version of the internal recorder buffer.
func (lr *SSHPacketRecorder) String() string {
	return lr.buffer.String()
}

// Close closes the underline reader.
func (lr *SSHPacketRecorder) Close() error {
	return lr.ReadCloser.Close()
}
