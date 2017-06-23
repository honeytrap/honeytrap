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

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/event"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:slack")

var (
	_ = pushers.RegisterBackend("slack", NewWith)
)

// Config defines a struct which holds configuration field values used by the
// SlackBackend for it's message delivery to the slack channel API.
type Config struct {
	WebhookURL string `toml:"webhook_url"`
	Username   string `toml:"username"`
	IconURL    string `toml:"icon_url"`
	IconEmoji  string `toml:"icon_emoji"`
}

// SlackBackend provides a struct which holds the configured means by which
// slack notifications are sent into giving slack groups and channels.
type SlackBackend struct {
	config Config

	ch chan map[interface{}]interface{}
}

// New returns a new instance of a SlackBackend.
func New(config Config) *SlackBackend {
	backend := SlackBackend{
		config: config,
		ch:     make(chan map[interface{}]interface{}, 100),
	}

	go backend.run()

	return &backend
}

// NewWith defines a function to return a pushers.Backend which delivers
// new messages to a giving underline slack channel defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var config Config

	if err := meta.PrimitiveDecode(data, &config); err != nil {
		return nil, err
	}

	if config.WebhookURL == "" {
		return nil, errors.New("Invalid Config: WebhookURL can not be empty")
	}

	return New(config), nil
}

func (b SlackBackend) run() {
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
		newMessage.Text = fmt.Sprintf("Event with Category %q of Type %q for Sensor %q occured", category, etype, sensor)

		if m, ok := ev["message"].(string); ok {
			newMessage.Text = m
		}

		newMessage.IconURL = b.config.IconURL
		newMessage.IconEmoji = b.config.IconEmoji
		newMessage.Username = b.config.Username

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
				fieldAttachment.AddField(fmt.Sprintf("%+s", name), vo)
				break

			default:
				data, err := json.Marshal(value)
				if err != nil {
					continue
				}

				fieldAttachment.AddField(fmt.Sprintf("%+s", name), string(data))
			}
		}

		newMessage.AddAttachment(idAttachment)
		newMessage.AddAttachment(fieldAttachment)

		newMessage.AddAttachment(Attachment{
			Title:    "Event Data",
			Author:   "HoneyTrap",
			Fallback: string(messageBuffer.Bytes()),
			Text:     string(messageBuffer.Bytes()),
		})

		data := new(bytes.Buffer)
		if err := json.NewEncoder(data).Encode(newMessage); err != nil {
			log.Errorf("Error encoding new SlackMessage: %+q", err)
			return
		}

		req, err := http.NewRequest("POST", b.config.WebhookURL, data)
		if err != nil {
			log.Errorf("Error while creating new request object: %+q", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Errorf("Error while making request to endpoint(%q): %q", b.config.WebhookURL, err.Error())
			return
		}

		defer res.Body.Close()

		// Though we expect slack not to deliver any messages to us but to be safe
		// discard and close body.
		io.Copy(ioutil.Discard, res.Body)

		if res.StatusCode == http.StatusOK {
		} else if res.StatusCode == http.StatusCreated {
		} else {
			log.Errorf("API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, b.config.WebhookURL)
			return
		}

	}
}

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (b SlackBackend) Send(e event.Event) {
	mp := make(map[interface{}]interface{})

	e.Range(func(key, value interface{}) bool {
		mp[key] = value
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
