package ftp

import (
	"context"
	"net"
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var (
	_   = services.Register("ftp", FTP)
	log = logging.MustGetLogger("services/ftp")
)

func FTP(options ...services.ServicerFunc) services.Servicer {

	s := &ftpService{
		Config: Config{
			recv:   make(chan string),
			driver: NewDummyfs(),
		},
	}

	for _, o := range options {
		o(s)
	}

	opts := &ServerOpts{
		Auth:           &FtpUser{},
		Name:           s.ServerName,
		WelcomeMessage: s.Banner,
		PassivePorts:   "20200-20250",
	}
	s.server = NewServer(opts)

	fs := NewDummyfs()

	log.Debugf("dummyfiles: %v", s.Dummyfiles)
	for _, df := range strings.Split(s.Dummyfiles, " ") {
		fs.makefile(df)
	}
	/*
		fs.makefile("passwords.txt")
		fs.makefile("users.db.bak")
		fs.MakeDir("tmp")
		fs.makefile("/tmp/index.html")
	*/
	s.driver = fs

	if store, err := Storage(); err != nil {
		log.Errorf("FTP: Could not initialize storage. %s", err.Error())
	} else {
		cert, err := store.Certificate()
		if err != nil {
			log.Errorf("TLS error: %s", err.Error())
		} else {
			s.server.tlsConfig = simpleTLSConfig(cert)
			//s.server.TLS = true
			s.server.ExplicitFTPS = true
			log.Debug("FTP: set TLS OK")
		}
	}

	return s
}

type Config struct {
	recv chan string

	server *Server

	driver Driver

	Banner string `toml:"banner"`

	Dummyfiles string `toml:"files"`

	ServerName string `toml:"name"`
}

type ftpService struct {
	Config

	c pushers.Channel
}

func (s *ftpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *ftpService) Handle(ctx context.Context, conn net.Conn) error {

	ftpConn := s.server.newConn(conn, s.driver, s.recv)

	go func() {
		for {
			select {
			case msg := <-s.recv:
				/*
					if msg == "q" {
						break
					}
				*/
				s.c.Send(event.New(
					services.EventOptions,
					event.Category("ftp"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Custom("ftp.sessionid", ftpConn.sessionid),
					event.Custom("ftp.command", strings.Trim(msg, "\r\n")),
				))
			}
		}
	}()

	ftpConn.Serve()

	return nil
}
