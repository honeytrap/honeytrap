package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:slack")

var (
	_ = pushers.RegisterBackend("slack", NewWith)
)

// APIConfig defines a struct which holds configuration field values used by the
// SlackBackend for it's message delivery to the slack channel API.
type APIConfig struct {
	WebhookURL string `toml:"webhook_url"`
	Username   string `toml:"username"`
	IconURL    string `toml:"icon_url"`
	IconEmoji  string `toml:"icon_emoji"`
}

// SlackBackend provides a struct which holds the configured means by which
// slack notifications are sent into giving slack groups and channels.
type SlackBackend struct {
	*http.Client
	config APIConfig
}

// New returns a new instance of a SlackBackend.
func New(config APIConfig) SlackBackend {
	return SlackBackend{
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		},
		config: config,
	}
}

// NewWith defines a function to return a pushers.Backend which delivers
// new messages to a giving underline slack channel defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var config APIConfig

	if err := meta.PrimitiveDecode(data, &config); err != nil {
		return nil, err
	}

	if config.WebhookURL == "" {
		return nil, errors.New("Invalid Config: WebhookURL can not be empty")
	}

	return New(config), nil
}

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (mc SlackBackend) Send(message message.Event) {
	//Attempt to encode message body first and if failed, log and continue.
	messageBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(messageBuffer).Encode(message.Data); err != nil {
		log.Errorf("SlackBackend: Error encoding data: %q", err.Error())
		return
	}

	// Create the appropriate fields for the giving slack message.
	var fields []Field
	var sensors []Field

	sensors = append(sensors, Field{
		Title: "Sensor",
		Value: message.Sensor,
		Short: true,
	})

	sensors = append(sensors, Field{
		Title: "Category",
		Value: string(message.Category),
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Sensor",
		Value: message.Sensor,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Date",
		Value: message.Date.UTC().String(),
		Short: true,
	})

	fields = append(fields, Field{
		Title: "HostAddr",
		Value: message.HostAddr,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "LocalAddr",
		Value: message.LocalAddr,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Token",
		Value: message.Token,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "End Time",
		Value: message.Ended.UTC().String(),
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Start Time",
		Value: message.Started.UTC().String(),
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Location",
		Value: message.Location,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Category",
		Value: string(message.Category),
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Session ID",
		Value: message.SessionID,
		Short: true,
	})

	fields = append(fields, Field{
		Title: "Container ID",
		Value: message.ContainerID,
		Short: true,
	})

	var newMessage Message

	newMessage.IconURL = mc.config.IconURL
	newMessage.IconEmoji = mc.config.IconEmoji
	newMessage.Username = mc.config.Username
	newMessage.Text = message.EventMessage()

	newMessage.Attachments = append(newMessage.Attachments, Attachment{
		Title:    "Event Identification",
		Author:   "HoneyTrap",
		Fields:   sensors,
		Text:     "Event Sensor and Category",
		Fallback: "Event Sensor and Category",
	})

	newMessage.Attachments = append(newMessage.Attachments, Attachment{
		Title:    "Event Fields",
		Author:   "HoneyTrap",
		Fields:   fields,
		Text:     "Fields for events",
		Fallback: "Fields for events",
	})

	newMessage.Attachments = append(newMessage.Attachments, Attachment{
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

	req, err := http.NewRequest("POST", mc.config.WebhookURL, data)
	if err != nil {
		log.Errorf("Error while creating new request object: %+q", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := mc.Do(req)
	if err != nil {
		log.Errorf("Error while making request to endpoint(%q): %q", mc.config.WebhookURL, err.Error())
		return
	}

	defer res.Body.Close()

	// Though we expect slack not to deliver any messages to us but to be safe
	// discard and close body.
	io.Copy(ioutil.Discard, res.Body)

	if res.StatusCode == http.StatusOK {
	} else if res.StatusCode == http.StatusCreated {
	} else {
		log.Errorf("SlackMessageBackend: API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, mc.config.WebhookURL)
	}
}

// Message defines the base message to be included sent to a slack endpoint.
type Message struct {
	Text        string       `json:"text"`
	IconEmoji   string       `json:"icon_emoji"`
	IconURL     string       `json:"icon_url"`
	Username    string       `json:"username"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment defines a struct to define an attachment to be included with a message.
type Attachment struct {
	Title     string  `json:"title"`
	Author    string  `json:"author_name,omitempty"`
	Fallback  string  `json:"fallback,omitempty"`
	Fields    []Field `json:"fields"`
	Text      string  `json:"text"`
	Timestamp int64   `json:"ts"`
}

// Field defines a field item to be shown on a Message.
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
