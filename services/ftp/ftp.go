package ftp

import (
	"context"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

var (
	_ = services.Register("ftp", FTP)
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

	s.c.Send(event.New(
		EventOptions,
		event.Category("ftp"),
		event.SourceAddr(conn.remoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
	))

	return nil
}
