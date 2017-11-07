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

	/*
		// Set the sender and recipient
		if err = c.Mail("sender@example.org"); err != nil {
			t.Fatalf("Error in Mail(\"sender@example.org\"): %s", err)
		}
		if err = c.Rcpt("recipient@example.net"); err != nil {
			t.Fatalf("Error in Rcpt(\"recipient@example.net\"): %s", err)
		}
			if err = c.Mail(""); err != nil {
				t.Fatalf("Error in Mail(\"\"): %s", err)
			}
			if err = c.Rcpt(""); err != nil {
				t.Fatalf("Error in Rcpt(\"\"): %s", err)
			}
		// Send the email body.
		wc, err := c.Data()
		if err != nil {
			t.Fatalf("Error in Data(): %s", err)
		}
		_, err = fmt.Fprintf(wc, "MIME-version: 1.0\r\nThis is the email body\r\n.\r\n")
		if err != nil {
			t.Fatalf("Error writing the email body: %s", err)
		}
		err = wc.Close()
		if err != nil {
			t.Fatal(err)
		}
	*/

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		t.Fatal(err)
	}

	// Using PLAIN AUTH
	hostname := "localhost"
	auth := smtp.PlainAuth("", "user@example.com", "password", hostname)

	err = smtp.SendMail(hostname+smtpAddress, auth, "user@example.com", []string{"somebody@example.com"}, []byte("hello somebody"))
	if err != nil {
		t.Fatal(err)
	}
}
