package smtp

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/storage"
)

const (
	hostname  = "testing.com"
	sender    = "sender@testing.com"
	recipient = "recipient@example.net"
	body      = "Subject: test message\r\nDate: Wed, 11 May 2011 16:19:57 -0400\r\n\r\nTESTING...\r\n.\r\n"
)

func TestMain(m *testing.M) {
	storage.SetDataDir("")
	c := m.Run()
	if err := storage.Close(); err != nil {
		log.Fatal(err)
	}
	if err := os.RemoveAll("badger.db"); err != nil {
		log.Fatal(err)
	}
	os.Exit(c)
}

func TestSMTP(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	// Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Fatal(err)
		}
	}(server)

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	//Send data client->server

	//Is TLS available?
	conf := s.srv.tlsConfig
	if conf == nil {
		t.Error("TLS config is not set")
	}

	err = smtpClient.StartTLS(conf)
	if err != nil {
		t.Error(err)
	}

	auth := smtp.PlainAuth("", "john", "bye", hostname)
	if err = smtpClient.Auth(auth); err != nil {
		t.Fatal(err)
	}

	// Set the sender and recipient first
	err = smtpClient.Mail(sender)
	if err != nil {
		t.Fatal(err)
	}
	err = smtpClient.Rcpt(recipient)
	if err != nil {
		t.Fatal(err)
	}

	// Send the email body.
	wc, err := smtpClient.Data()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, body)
	if err != nil {
		t.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}

	/*
		//BUG: reading/writing from closed pipe, only in testing
			// Send the QUIT command and close the connection.
			err = smtpClient.Quit()
			if err != nil {
				t.Error(err)
			}
	*/
	// Check if data is received.
	// with file channel?
}

func TestAuthLogin(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	// Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Fatal(err)
		}
	}(server)

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	auth := &loginAuth{"john", "bye"}

	if err := smtpClient.Auth(auth); err != nil {
		t.Fatal(err)
	}
}

func TestCramMD5(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	// Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err != nil {
			t.Fatal(err)
		}
	}(server)

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	auth := smtp.CRAMMD5Auth("john", "secret key")

	if err := smtpClient.Auth(auth); err != nil {
		t.Fatal(err)
	}
}

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown question")
		}
	}
	return nil, nil
}
