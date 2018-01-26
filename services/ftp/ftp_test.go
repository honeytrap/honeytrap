package ftp

import (
	"net"
	"os"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
)

var client *ServerConn
var clt, srv net.Conn

func TestMain(m *testing.M) {

	//Setup client and server
	clt, srv = net.Pipe()
	defer clt.Close()
	defer srv.Close()

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
	}(srv)

	client, err := Connect(clt)
	if err != nil {
		t.Fatal(err)
	}

	user := "admin"
	password := "god"

	if err := client.Login(user, password); err != nil {
		t.Errorf("Could not login user: %s password: %s", user, password)
	}

	if _, err := client.List(""); err != nil {
		t.Errorf("Error in LIST command: %s", err.Error())
	}

	if _, err := client.NameList(""); err != nil {
		t.Errorf("Error in NLST command: %s", err.Error())
	}

	if err := client.ChangeDir("/private"); err != nil {
		t.Errorf("Error with CWD command: %s", err.Error())
	}

	if err := client.ChangeDirToParent(); err != nil {
		t.Errorf("Error with CDUP command: %s", err.Error())
	}

	if _, err := client.CurrentDir(); err != nil {
		t.Errorf("Error with PWD command: %s", err.Error())
	}

	if r, err := client.Retr("myfile"); err != nil {
		r.Close()
		t.Errorf("Error with RETR command: %s", err.Error())
	} else {
		r.Close()
	}

	if r, err := client.RetrFrom("myfile", 64); err != nil {
		r.Close()
		t.Errorf("Error with RETR command with offset: %s", err.Error())
	} else {
		r.Close()
	}
}
