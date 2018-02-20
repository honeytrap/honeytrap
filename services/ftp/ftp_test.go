package ftp

import (
	"net"
	"os"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/storage"
)

const (
	user     = "anonymous"
	password = "anonymous"
)

var (
	clt, srv net.Conn
)

func TestMain(m *testing.M) {
	storage.SetDataDir("/tmp")
	os.Exit(m.Run())
}

func TestFTP(t *testing.T) {

	//Setup client and server
	clt, srv = net.Pipe()
	defer clt.Close()
	defer srv.Close()

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
