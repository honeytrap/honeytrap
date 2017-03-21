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

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:slack")
var _ = pushers.Register("slack", NewMessageChannel())

// MessageChannel provides a struct which holds the configured means by which
// slack notifications are sent into giving slack groups and channels.
type MessageChannel struct {
	client *http.Client
	host   string
	token  string
}

// NewMessageChannel returns a new instance of a slack MessageChannel.
func NewMessageChannel() pushers.ChannelFunc {
	return func(conf map[string]interface{}) (pushers.Channel, error) {
		var client http.Client
		client.Transport = &http.Transport{MaxIdleConnsPerHost: 5}
		client.Timeout = 20 * time.Second

		var host string
		var token string

		var ok bool
		if host, ok = conf["host"].(string); !ok {
			return nil, errors.New("Host not provided for Slack Channel")
		}

		if token, ok = conf["token"].(string); !ok {
			return nil, errors.New("Token not provided for slack Channel")
		}

		return &MessageChannel{
			client: &client,
			host:   host,
			token:  token,
		}, nil
	}
}

type newSlackMessage struct {
	Text        string               `json:"text"`
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

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (mc MessageChannel) Send(messages []*pushers.PushMessage) {
	for _, message := range messages {

		// TODO: Implement message filtering through channel regexp.

		//Attempt to encode message body first and if failed, log and continue.
		messageBuffer := new(bytes.Buffer)
		if err := json.NewEncoder(messageBuffer).Encode(message.Data); err != nil {
			log.Errorf("SlackMessageChannel: Error encoding data: %q", err.Error())
			continue
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
			Value: message.Category,
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
			log.Errorf("SlackMessageChannel: Error encoding new SlackMessage: %+q", err)
			continue
		}

		reqURL := fmt.Sprintf("%s/%s", mc.host, mc.token)
		req, err := http.NewRequest("POST", reqURL, slackMessageBuffer)
		if err != nil {
			log.Errorf("SlackMessageChannel: Error while creating new request object: %+q", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := mc.client.Do(req)
		if err != nil {
			log.Errorf("SlackMessageChannel: Error while making request to endpoint(%q): %q", reqURL, err.Error())
			continue
		}

		// Though we expect slack not to deliver any messages to us but to be safe
		// discard and close body.
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			log.Errorf("SlackMessageChannel: API Response with unexpected Status Code[%d] to endpoint: %q", res.StatusCode, reqURL)
			continue
		}
	}
}
