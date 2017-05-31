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
	*http.Client
	config Config
}

// New returns a new instance of a SlackBackend.
func New(config Config) SlackBackend {
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
	var config Config

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
func (mc SlackBackend) Send(mesg message.Event) {
	log.Infof("Sending Message: %#v", mesg)

	//Attempt to encode message body first and if failed, log and continue.
	messageBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(messageBuffer).Encode(mesg.Data); err != nil {
		log.Errorf("Error encoding data: %q", err.Error())
		return
	}

	var newmesg Message

	newmesg.IconURL = mc.config.IconURL
	newmesg.IconEmoji = mc.config.IconEmoji
	newmesg.Username = mc.config.Username

	mo, ok := interface{}(mesg).(message.Messager)

	if ok {
		newmesg.Text = mo.Message()
	} else if mesg.Details != nil {

		if detailMessage, ok := mesg.Details["message"].(string); ok {
			newmesg.Text = detailMessage
		} else {
			newmesg.Text = mesg.String()
		}

	} else {
		newmesg.Text = mesg.String()
	}

	idAttachment := Attachment{
		Title:    "Event Identification",
		Author:   "HoneyTrap",
		Text:     "Event Sensor and Category",
		Fallback: "Event Sensor and Category",
	}

	idAttachment.AddField("Sensor", mesg.Sensor).
		AddField("Category", string(mesg.Category)).
		AddField("Type", string(mesg.Type))

	fieldAttachment := Attachment{
		Title:    "Event Fields",
		Author:   "HoneyTrap",
		Text:     "Fields for events",
		Fallback: "Fields for events",
	}

	fieldAttachment.AddField("Sensor", mesg.Sensor).
		AddField("Category", string(mesg.Category)).
		AddField("HostAddr", mesg.HostAddr).
		AddField("LocalAddr", mesg.LocalAddr).
		AddField("Token", mesg.Token).
		AddField("Type", string(mesg.Type)).
		AddField("Location", mesg.Location).
		AddField("Session ID", mesg.SessionID).
		AddField("Container ID", mesg.ContainerID).
		AddField("Date", mesg.Date.UTC().String()).
		AddField("Start Time", mesg.Started.UTC().String()).
		AddField("End Time", mesg.Ended.UTC().String())

	newmesg.AddAttachment(idAttachment)
	newmesg.AddAttachment(fieldAttachment)

	newmesg.AddAttachment(Attachment{
		Title:    "Event Data",
		Author:   "HoneyTrap",
		Fallback: string(messageBuffer.Bytes()),
		Text:     string(messageBuffer.Bytes()),
	})

	data := new(bytes.Buffer)
	if err := json.NewEncoder(data).Encode(newmesg); err != nil {
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
		log.Errorf("API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, mc.config.WebhookURL)
		return
	}

	log.Infof("Delivered Message: %#v", mesg)
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

// Attachment defines a struct to define an attachment to be included with a mesg.
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

// Field defines a field item to be shown on a mesg.
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
