package services

import (
	"net"
	"net/smtp"
	"testing"
)

const smtpAddress = ":8025"

func TestSMTP(t *testing.T) {
	//Create a server routine
	go func() {
		ln, err := net.Listen("tcp", smtpAddress)
		if err != nil {
			t.Fatal(err)
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			server := SMTP()
			err = server.Handle(conn)
			if err != nil {
				t.Fatalf("Error handling server!: %s", err)
			}
		}
	}()

	// Connect to the remote SMTP server.
	c, err := smtp.Dial("localhost" + smtpAddress)
	if err != nil {
		t.Fatalf("Error in Dial(localhost%s): %s", smtpAddress, err)
	}
	// Check STARTTLS
	//auth := smtp.PlainAuth("", "user@example.com", "password", "localhost"+smtpAddress)
	//t.Errorf("Auth = %v", auth)
	//if err = c.StartTLS(s.srv.TLSConfig); err != nil {
	//	t.Error(err)
	//}
	// Set the sender and recipient
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Error in Mail(\"sender@example.org\"): %s", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Error in Rcpt(\"recipient@example.net\"): %s", err)
	}
}
