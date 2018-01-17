package ftp

import (
	"net"
	"os"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
)

var client, server net.Conn

func TestMain(m *testing.M) {

	//Setup client and server
	client, server = net.Pipe()
	defer client.Close()
	defer server.Close()

	os.Exit(m.Run())
}

func TestFTP(t *testing.T) {
	s := FTP().(*ftpService)

	c, _ := pushers.Dummy()
	s.SetChannel(c)

	//Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Fatal(err)
		}
	}(server)

	//Create ftp client
}
