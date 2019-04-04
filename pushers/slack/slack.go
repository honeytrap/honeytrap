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
package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:slack")

var (
	_ = pushers.Register("slack", New)
)

// Config defines a struct which holds configuration field values used by the
// Backend for it's message delivery to the slack channel API.
type Config struct {
	WebhookURL string `toml:"webhook_url"`
	Username   string `toml:"username"`
	IconURL    string `toml:"icon_url"`
	IconEmoji  string `toml:"icon_emoji"`
}

// Backend provides a struct which holds the configured means by which
// slack notifications are sent into giving slack groups and channels.
type Backend struct {
	Config

	ch chan map[string]interface{}
}

// New returns a new instance of a Backend.
func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	c := Backend{
		ch: make(chan map[string]interface{}, 100),
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	if c.WebhookURL == "" {
		return nil, errors.New("Invalid Config: WebhookURL can not be empty")
	}

	go c.run()

	return &c, nil
}

func (b Backend) run() {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
		Timeout: time.Duration(20) * time.Second,
	}

	for {
		ev := <-b.ch

		//Attempt to encode message body first and if failed, log and continue.
		var messageBuffer bytes.Buffer

		category, ok := ev["category"].(string)
		if !ok {
			log.Errorf("Error event has no category value")
			return
		}

		sensor, ok := ev["sensor"].(string)
		if !ok {
			log.Errorf("Error event has no sensor value")
			return
		}

		etype, ok := ev["type"].(string)
		if !ok {
			log.Errorf("Error event has no type value")
			return
		}

		var newMessage Message
		newMessage.Text = fmt.Sprintf("Event with category %q of type %q for sensor %q occurred", category, etype, sensor)

		if m, ok := ev["message"].(string); ok {
			newMessage.Text = m
		}

		newMessage.IconURL = b.IconURL
		newMessage.IconEmoji = b.IconEmoji
		newMessage.Username = b.Username

		idAttachment := Attachment{
			Title:    "Event Identification",
			Author:   "HoneyTrap",
			Text:     "Event Sensor and Category",
			Fallback: "Event Sensor and Category",
		}

		idAttachment.AddField("Sensor", string(sensor)).
			AddField("Category", string(category)).
			AddField("Type", string(etype))

		fieldAttachment := Attachment{
			Title:    "Event Fields",
			Author:   "HoneyTrap",
			Text:     "Fields for events",
			Fallback: "Fields for events",
		}

		fieldAttachment.AddField("Sensor", string(sensor)).
			AddField("Category", string(category)).
			AddField("Type", string(etype))

		for name, value := range ev {
			switch vo := value.(type) {
			case string:
				fieldAttachment.AddField(name, vo)
			default:
				data, err := json.Marshal(value)
				if err != nil {
					continue
				}

				fieldAttachment.AddField(name, string(data))
			}
		}

		newMessage.AddAttachment(idAttachment)
		newMessage.AddAttachment(fieldAttachment)

		newMessage.AddAttachment(Attachment{
			Title:    "Event Data",
			Author:   "HoneyTrap",
			Fallback: messageBuffer.String(),
			Text:     messageBuffer.String(),
		})

		data := new(bytes.Buffer)
		if err := json.NewEncoder(data).Encode(newMessage); err != nil {
			log.Errorf("Error encoding new SlackMessage: %+q", err)
			return
		}

		req, err := http.NewRequest("POST", b.WebhookURL, data)
		if err != nil {
			log.Errorf("Error while creating new request object: %+q", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Errorf("Error while making request to endpoint(%q): %q", b.WebhookURL, err.Error())
			return
		}

		defer res.Body.Close()

		// Though we expect slack not to deliver any messages to us but to be safe
		// discard and close body.
		io.Copy(ioutil.Discard, res.Body)

		if res.StatusCode == http.StatusOK {
		} else if res.StatusCode == http.StatusCreated {
		} else {
			log.Errorf("API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, b.WebhookURL)
			return
		}

	}
}

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (b Backend) Send(e event.Event) {
	mp := make(map[string]interface{})

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	b.ch <- mp
}

// Message defines the base message to be included sent to a slack endpoint.
type Message struct {
	Text        string       `json:"text"`
	IconEmoji   string       `json:"icon_emoji"`
	IconURL     string       `json:"icon_url"`
	Username    string       `json:"username"`
	Attachments []Attachment `json:"attachments"`
}

// AddAttachment adds a field into the slice for the given attachment.
func (a *Message) AddAttachment(attachment Attachment) {
	a.Attachments = append(a.Attachments, attachment)
}

// Attachment defines a struct to define an attachment to be included with a event.
type Attachment struct {
	Title     string  `json:"title"`
	Author    string  `json:"author_name,omitempty"`
	Fallback  string  `json:"fallback,omitempty"`
	Fields    []Field `json:"fields"`
	Text      string  `json:"text"`
	Timestamp int64   `json:"ts"`
}

// AddField adds a field into the slice for the given attachment.
func (a *Attachment) AddField(title string, value string) *Attachment {
	a.Fields = append(a.Fields, Field{Title: title, Value: value, Short: true})
	return a
}

// Field defines a field item to be shown on a event.
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
