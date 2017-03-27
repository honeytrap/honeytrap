package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:elasticsearch")

// SearchChannel defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type SearchChannel struct {
	client *http.Client
	host   string
}

// UnmarshalConfig attempts to unmarshal the provided value into the giving
// ElasticSearchChannel.
func (hc *SearchChannel) UnmarshalConfig(m interface{}) error {
	conf, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("Expected to receive a map")
	}

	if hc.client == nil {
		hc.client = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		}
	}

	if hc.host, ok = conf["host"].(string); !ok {
		return fmt.Errorf("Host not set for channel elasticsearch")
	}

	return nil
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc SearchChannel) Send(messages []message.PushMessage) {
	for _, message := range messages {
		buf := new(bytes.Buffer)

		if message.Sensor == "honeytrap" && message.Category == "ping" {
			// ignore
			continue
		}

		if err := json.NewEncoder(buf).Encode(message.Data); err != nil {
			log.Errorf("ElasticSearchChannel: Error encoding data: %s", err.Error())
			continue
		}

		messageID := uuid.NewV4()
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s/%s", hc.host, message.Sensor, message.Category, messageID.String()), buf)
		if err != nil {
			log.Errorf("ElasticSearchChannel: Error while preparing request: %s", err.Error())
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		var resp *http.Response
		if resp, err = hc.client.Do(req); err != nil {
			log.Errorf("ElasticSearchChannel: Error while sending messages: %s", err.Error())
			continue
		}

		// for keep-alive
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			log.Errorf("ElasticSearchChannel: Unexpected status code: %d", resp.StatusCode)
			continue
		}

	}

	log.Infof("ElasticSearchChannel: Sent %d actions.", len(messages))
}
