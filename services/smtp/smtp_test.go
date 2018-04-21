package smtp

import (
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
	os.Exit(m.Run())
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
