package ftp

import (
	"context"
	"net"

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

	opts := &ServerOpts{}
	srv := NewServer(opts)

	driver := &DummyFS{}
	ftpConn := srv.newConn(conn, driver)

	ftpConn.Serve()

	log.Debug("Sending event data")

	s.c.Send(event.New(
		services.EventOptions,
		event.Category("ftp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("ftp.user", ftpConn.user),
		event.Custom("ftp.password", ftpConn.password),
	))

	return nil
}
