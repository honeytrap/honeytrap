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
	//client   *ServerConn
	clt, srv net.Conn
)

/*
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}
*/

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
	log.Debug("Test: client started")

	if err := client.Login(user, password); err != nil {
		t.Errorf("Could not login user: %s password: %s", user, password)
	}

	/*
			if _, err := client.List("/"); err != nil {
				t.Errorf("Error in LIST command: %s", err.Error())
			}

				if _, err := client.NameList(""); err != nil {
					t.Errorf("Error in NLST command: %s", err.Error())
				}

		if err := client.ChangeDir("/not/valid"); err == nil {
			t.Error("CWD command: Error expected!")
		}

		if err := client.ChangeDir("/mydir"); err != nil {
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

			if err := client.Logout(); err != nil {
				t.Errorf("Error with Logout: %s", err.Error())
			}
	*/
	if err := client.Quit(); err != nil {
		t.Errorf("Error with Quit: %s", err.Error())
	}
}
