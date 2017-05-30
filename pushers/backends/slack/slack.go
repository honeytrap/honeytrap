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
	Host  string `toml:"host"`
	Token string `toml:"token"`
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

	if config.Host == "" {
		return nil, errors.New("slack.APIConfig Invalid: Host can not be empty")
	}

	if config.Token == "" {
		return nil, errors.New("slack.APIConfig Invalid: Token can not be empty")
	}

	return New(config), nil
}

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (mc SlackBackend) Send(message message.Event) {
	//Attempt to encode message body first and if failed, log and continue.
	messageBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(messageBuffer).Encode(message.Data); err != nil {
		log.Errorf("SlackMessageBackend: Error encoding data: %q", err.Error())
		return
	}

	// Create the appropriate fields for the giving slack message.
	var fields []newSlackField

	fields = append(fields, newSlackField{
		Title: "Sensor",
		Value: message.Sensor,
		Short: true,
	})

	fields = append(fields, newSlackField{
		Title: "Category",
		Value: string(message.Category),
		Short: true,
	})

	fields = append(fields, newSlackField{
		Title: "Session ID",
		Value: message.SessionID,
		Short: true,
	})

	fields = append(fields, newSlackField{
		Title: "Container ID",
		Value: message.ContainerID,
		Short: true,
	})

	var slackMessage newSlackMessage
	slackMessage.Text = fmt.Sprintf("New Sensor Message from %q with Category %q", message.Sensor, message.Category)
	slackMessage.Attachments = append(slackMessage.Attachments, newSlackAttachment{
		Title:    "Sensor Data",
		Author:   "HoneyTrap",
		Fields:   fields,
		Text:     string(messageBuffer.Bytes()),
		Fallback: fmt.Sprintf("New SensorMessage (Sensor: %q, Category: %q, Session: %q, Container: %q). Check Slack for more", message.Sensor, message.Category, message.SessionID, message.ContainerID),
	})

	slackMessageBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(slackMessageBuffer).Encode(slackMessage); err != nil {
		log.Errorf("Error encoding new SlackMessage: %+q", err)
		return
	}

	reqURL := fmt.Sprintf("%s/%s", mc.config.Host, mc.config.Token)
	req, err := http.NewRequest("POST", reqURL, slackMessageBuffer)
	if err != nil {
		log.Errorf("Error while creating new request object: %+q", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := mc.Do(req)
	if err != nil {
		log.Errorf("Error while making request to endpoint(%q): %q", reqURL, err.Error())
		return
	}

	defer res.Body.Close()

	// Though we expect slack not to deliver any messages to us but to be safe
	// discard and close body.
	io.Copy(ioutil.Discard, res.Body)

	if res.StatusCode == http.StatusOK {
	} else if res.StatusCode == http.StatusCreated {
	} else {
		log.Errorf("SlackMessageBackend: API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, reqURL)
	}
}

type newSlackMessage struct {
	Text        string               `json:"text"`
	Backend     string               `json:"channel"`
	Attachments []newSlackAttachment `json:"attachments"`
}

type newSlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type newSlackAttachment struct {
	Title     string          `json:"title"`
	Author    string          `json:"author_name,omitempty"`
	Fallback  string          `json:"fallback,omitempty"`
	Fields    []newSlackField `json:"fields"`
	Text      string          `json:"text"`
	Timestamp int64           `json:"ts"`
}
