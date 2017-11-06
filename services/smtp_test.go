package services

import (
	"fmt"
	"net"
	"net/smtp"
	"testing"
)

const smtpAddress = ":8025"

func TestSMTP(t *testing.T) {
	//Create a server
	ln, err := net.Listen("tcp", smtpAddress)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
	}

	c := SMTP()
	c.Handle(conn)

	// Connect to the remote SMTP server.
	c, err := smtp.Dial("tcp", "localhost"+smtpAddress)
	if err != nil {
		t.Fatalf("Error in Dial(localhost:8025): %s", err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("Error in Mail(\"sender@example.org\"): %s", err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("Error in Rcpt(\"recipient@example.net\"): %s", err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Error in Data(): %s", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Error writing the email body: %s", err)
	}
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		t.Fatal(err)
	}
}
