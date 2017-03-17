package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

var log = logging.MustGetLogger("honeytrap:channels:slack")

// MessageChannel provides a struct which holds the configured means by which
// slack notifications are sent into giving slack groups and channels.
type MessageChannel struct {
	client   *http.Client
	host     string
	token    string
	fields   map[string]*regexp.Regexp
	channels []channelSelector
}

// Unmarshal attempts to unmarshal the provided value into the giving
// MessageChannel.
func (mc *MessageChannel) UnmarshalConfig(m interface{}) error {
	conf, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("Expected to receive a map")
	}

	var host string
	var token string

	if host, ok = conf["host"].(string); !ok {
		return errors.New("Host not provided for Slack Channel")
	}

	if token, ok = conf["token"].(string); !ok {
		return errors.New("Token not provided for slack Channel")
	}

	fieldMatchers := make(map[string]*regexp.Regexp)
	if fields, ok := conf["fields"].(map[string]interface{}); ok {
		for key, value := range fields {
			switch realValue := value.(type) {
			case *regexp.Regexp:
				fieldMatchers[key] = realValue
			case string:
				fieldMatchers[key] = regexp.MustCompile(realValue)
			default:
				// TODO: Do we want to continue or return error here?
				continue
			}
		}
	}

	channelSelectors := make([]channelSelector, 0)
	if selections, ok := conf["channels"].([]map[string]interface{}); ok {
		for _, selection := range selections {
			var matcher *regexp.Regexp

			switch rx := selection["value"].(type) {
			case *regexp.Regexp:
				matcher = rx
			case string:
				matcher = regexp.MustCompile(rx)
			default:
				// TODO: Do we want to continue or return error here?
				continue
			}

			field := selection["field"].(string)
			if !ok {
				continue
			}

			channel, ok := selection["channel"].(string)
			if !ok {
				continue
			}

			channelToken, ok := selection["token"].(string)
			if !ok {
				continue
			}

			channelSelectors = append(channelSelectors, channelSelector{
				Channel: channel,
				Field:   field,
				Matcher: matcher,
				Token:   channelToken,
			})
		}
	}

	mc.host = host
	mc.token = token
	mc.fields = fieldMatchers
	mc.channels = channelSelectors
	mc.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
		Timeout: time.Duration(20) * time.Second,
	}

	return nil
}

type channelSelector struct {
	Field   string
	Token   string
	Channel string
	Matcher *regexp.Regexp
}

// Send delivers the giving push messages to the required slack channel.
// TODO: Ask if Send shouldnt return an error to allow proper delivery validation.
func (mc MessageChannel) Send(messages []*message.PushMessage) {
	for _, message := range messages {

		// Run through all the available fields and their regexp,
		// if the field regexp fails to match, then we skip the message.
		if matcher, ok := mc.fields["sensor"]; ok && !matcher.MatchString(message.Sensor) {
			log.Errorf("SlackMessageChannel: Failed to match sensor names match requirement")
			continue
		}

		if matcher, ok := mc.fields["category"]; ok && !matcher.MatchString(message.Category) {
			log.Errorf("SlackMessageChannel: Failed to match category with match requirement")
			continue
		}

		if matcher, ok := mc.fields["container_id"]; ok && !matcher.MatchString(message.ContainerID) {
			log.Errorf("SlackMessageChannel: Failed to match container_id with match requirement")
			continue
		}

		if matcher, ok := mc.fields["session_id"]; ok && !matcher.MatchString(message.SessionID) {
			log.Errorf("SlackMessageChannel: Failed to match session_id with match requirement")
			continue
		}

		// Attempt to match the first possible filter which provides true and select that
		// items Channel as the target channel. Its a first match first serve type of logic.
		var selectedChannel string
		var selectedToken string

		{
		chanSelect:
			for _, channel := range mc.channels {
				switch strings.ToLower(channel.Field) {
				case "sensor":
					if channel.Matcher.MatchString(message.Sensor) {
						selectedToken = channel.Token
						selectedChannel = channel.Channel
						break chanSelect
					}
				case "category":
					if channel.Matcher.MatchString(message.Category) {
						selectedToken = channel.Token
						selectedChannel = channel.Channel
						break chanSelect
					}
				case "session_id":
					if channel.Matcher.MatchString(message.SessionID) {
						selectedToken = channel.Token
						selectedChannel = channel.Channel
						break chanSelect
					}
				case "container_id":
					if channel.Matcher.MatchString(message.ContainerID) {
						selectedToken = channel.Token
						selectedChannel = channel.Channel
						break chanSelect
					}
				}
			}
		}

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
		slackMessage.Channel = selectedChannel
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

		var channelToken string

		if selectedToken == "" {
			channelToken = mc.token
		} else {
			channelToken = selectedToken
		}

		reqURL := fmt.Sprintf("%s/%s", mc.host, channelToken)
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

type newSlackMessage struct {
	Text        string               `json:"text"`
	Channel     string               `json:"channel"`
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
