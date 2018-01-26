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
	s := &ftpService{}

	for _, o := range options {
		o(s)
	}
	return s
}

type ftpService struct {
	c pushers.Channel
}

func (s *ftpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *ftpService) Handle(ctx context.Context, conn net.Conn) error {

	opts := &ServerOpts{
		Auth: &FtpUser{},
	}
	srv := NewServer(opts)

	rcv := make(chan string)
	driver := &DummyFS{}
	ftpConn := srv.newConn(conn, driver, rcv)

	go func() {
		for {
			select {
			case msg := <-rcv:
				if msg == "q" {
					break
				}
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
