// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sendgrid sends email using SendGrid.
package sendgrid // import "upspin.io/cloud/mail/sendgrid"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"upspin.io/cloud/mail"
	"upspin.io/errors"
)

// sendgrid implements cloud.mail.Mail using SendGrid as the underlying
// substratum.
type sendgrid struct {
	apiKey string
	domain string
}

var _ mail.Mail = (*sendgrid)(nil)

// New allocates a Mail type for sending email with SendGrid.
func New(apiKey, domain string) mail.Mail {
	return &sendgrid{
		apiKey: apiKey,
		domain: domain,
	}
}

// apiSend is the endpoint for requests. It's a var so tests can change it.
var apiSend = "https://api.sendgrid.com/v3/mail/send"

// Types below are Go JSON representations of SendGrid's API. The types are not
// exported, but their fields are because json.marshal must be able to see them.

// personalizations is a SendGrid's internal representation of the To and
// Subject fields.
type personalizations struct {
	To      []addr
	Subject string
}

// addr represents an email address with an optional recipient's name.
type addr struct {
	Email string
	Name  string
}

// content represents the body of the email with a type of "text/plain" or
// "text/html".
type content struct {
	Type  string
	Value string
}

// message represents a message to send.
// When more than one type of content is present plain must come before html.
type message struct {
	Personalizations []personalizations
	From             addr
	Content          []content
}

// Send implements cloud.mail.Mail.
func (s *sendgrid) Send(to, from, subject, text, html string) error {
	const op = "cloud/mail/sendgrid.Send"
	if text == "" && html == "" {
		return errors.E(op, errors.Invalid, errors.Str("text or html body must be provided"))
	}
	msg := message{
		Personalizations: []personalizations{
			{
				To: []addr{
					{
						Email: to,
					},
				},
				Subject: subject,
			},
		},
		From: addr{
			Email: fmt.Sprintf("%s@%s", from, s.domain),
		},
	}
	// The order of Content must be: plain, html.
	if text != "" {
		msg.Content = append(msg.Content, content{
			Type:  "text/plain",
			Value: text,
		})
	}
	if html != "" {
		msg.Content = append(msg.Content, content{
			Type:  "text/html",
			Value: html,
		})
	}

	data, err := json.Marshal(&msg)
	if err != nil {
		return errors.E(op, errors.IO, err)
	}
	req, err := http.NewRequest("POST", apiSend, bytes.NewBuffer(data))
	if err != nil {
		return errors.E(op, errors.IO, err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.E(op, errors.IO, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		errStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.E(op, errors.IO, err)
		}
		return errors.E(op, errors.IO, errors.Str(string(errStr)))
	}

	return nil
}

// Domain implements cloud.mail.Mail.
func (s *sendgrid) Domain() string {
	return s.domain
}
