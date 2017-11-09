package services

import (
	"fmt"
	"net"
	"net/smtp"
	"testing"
)

const (
	hostname = "testing.com"
	sender   = "sender@testing.com"
)

func TestSMTP(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	//Create Servicer
	//TODO: Use SMTP() to create server so TLS is also configured.
	s := &SMTPService{srv: &Server{TLSConfig: nil}}

	//Set Handle connection
	go func() {
		err := s.Handle(server)
		if err != nil {
			t.Fatal(err)
		}
	}()

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	//Send data client->server
	/*
			//Is TLS available? It should
			if err := smtpClient.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
				t.Error(err)
			}
		auth := smtp.PlainAuth("", sender, "password", hostname)
		if err = smtpClient.Auth(auth); err != nil {
			t.Error(err)
		}
	*/
	// Set the sender and recipient first
	if err := smtpClient.Mail("sender@example.org"); err != nil {
		t.Fatal(err)
	}
	if err := smtpClient.Rcpt("recipient@example.net"); err != nil {
		t.Fatal(err)
	}

	// Send the email body.
	wc, err := smtpClient.Data()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, "From: John Smith <jsmith@port25.com>\r\nTo: Jane Doe <jdoe@port25.com>\r\nSubject: test message\r\nDate: Wed, 11 May 2011 16:19:57 -0400\r\n\r\n.\r\n")
	if err != nil {
		t.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = smtpClient.Quit()
	if err != nil {
		t.Fatal(err)
	}
	//Check if data is received.

}
