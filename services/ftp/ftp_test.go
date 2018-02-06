package ftp

import (
	"net"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
)

const (
	user     = "admin"
	password = "god"
)

var (
	clt, srv net.Conn
)

func TestFTP(t *testing.T) {

	s := FTP().(*ftpService)

	c, _ := pushers.Dummy()
	s.SetChannel(c)

	//Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Fatal(err)
		}
	}(srv)

	client, err := Connect(clt)
	if err != nil {
		t.Fatal(err)
	}
	//log.Debug("Test: client started")

	if err := client.Login(user, password); err != nil {
		t.Errorf("Could not login user: %s password: %s", user, password)
	}

	if err := client.Quit(); err != nil {
		t.Errorf("Error with Quit: %s", err.Error())
	}
}
