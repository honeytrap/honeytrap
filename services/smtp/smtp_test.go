package smtp

import (
	"fmt"
	"net"
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
	storage.SetDataDir(os.TempDir())
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

		for {
			if err := s.Handle(nil, conn); err != nil {
				t.Errorf("Handler error: %s", err.Error())
				break
			}
		}
	}(server)

	n, err := fmt.Fprintf(client, "EHLO test")
	if err != nil {
		t.Error(err)
	} else if n != 9 {
		t.Errorf("wrong number %d: %s", n, err.Error())
	}

	/*
		fmt.Println("Creating smtp client...")

		//Create smtp client
		smtpClient, err := smtp.NewClient(client, hostname)
		if err != nil {
			t.Error(err)
		}

		fmt.Println("Client created")

		// check connection
		err = smtpClient.Noop()
		if err != nil {
			t.Fatalf("Can not create client: %s", err.Error())
		}

		// Set the sender and recipient first
		err = smtpClient.Mail(sender)
		if err != nil {
			t.Error(err)
		}
		err = smtpClient.Rcpt(recipient)
		if err != nil {
			t.Error(err)
		}

		// Send the email body.
		wc, err := smtpClient.Data()
		if err != nil {
			t.Error(err)
		}
		_, err = fmt.Fprintf(wc, body)
		if err != nil {
			t.Error(err)
		}
		err = wc.Close()
		if err != nil {
			t.Error(err)
		}

		//BUG: reading/writing from closed pipe, only in testing
		// Send the QUIT command and close the connection.
		err = smtpClient.Quit()
		if err != nil {
			t.Error(err)
		}
		// Check if data is received.
		// with file channel?
	*/
}
