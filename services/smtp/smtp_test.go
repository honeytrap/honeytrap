// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package smtp

import (
	"context"
	"crypto/tls"
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
	storage.SetDataDir(os.TempDir())
	os.Exit(m.Run())
}

func TestSMTP(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	done := make(chan bool)

	go func() {
		ctx := context.Background()
		err := s.handle(ctx, server)
		if err != nil {
			t.Errorf("Handling error: %s", err.Error())
		}
		done <- true
	}()

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	// check connection
	err = smtpClient.Noop()
	if err != nil {
		t.Errorf("Can not create client: %s", err.Error())
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

	// Send the QUIT command and close the connection.
	err = smtpClient.Quit()
	if err != nil {
		t.Errorf("QUIT Error: %s", err.Error())
	}

	<-done
}

func TestStartTLS(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	done := make(chan bool)

	go func() {
		ctx := context.Background()
		err := s.handle(ctx, server)
		if err != nil {
			t.Errorf("Handling error: %s", err.Error())
		}
		done <- true
	}()

	//Create smtp client
	smtpClient, err := smtp.NewClient(client, hostname)
	if err != nil {
		t.Error(err)
	}

	err = smtpClient.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Errorf("No StartTLS: %s", err.Error())
	}

	if state, ok := smtpClient.TLSConnectionState(); !ok {
		t.Errorf("TLS Connection state %v", state)
	}

	err = smtpClient.Close()

	<-done
}

func TestSmtpConn(t *testing.T) {
	//Create a pipe
	client, server := net.Pipe()
	defer server.Close()

	//Create Servicer
	s := SMTP().(*Service)

	// Create channel
	dc, _ := pushers.Dummy()
	s.SetChannel(dc)

	done := make(chan bool)

	go func() {
		ctx := context.Background()
		err := s.Handle(ctx, server)
		if err != nil {
			t.Errorf("Handling error: %s", err.Error())
		}
		done <- true
	}()

	tconn := tls.Client(client, &tls.Config{InsecureSkipVerify: true})
	defer tconn.Close()

	//Create smtp client
	smtpClient, err := smtp.NewClient(tconn, hostname)
	if err != nil {
		t.Error(err)
	}

	if err := smtpClient.Close(); err != nil {
		t.Errorf("Can not close tls client: %s", err.Error())
	}

	<-done
}
