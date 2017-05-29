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

	"github.com/BurntSushi/toml"
	"github.com/satori/go.uuid"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:elasticsearch")

// SearchConfig defines a struct which holds configuration values for a SearchChannel.
type SearchConfig struct {
	Host string
}

// SearchChannel defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type SearchChannel struct {
	client *http.Client
	config SearchConfig
}

// New returns a new instance of a SearchChannel.
func New(conf SearchConfig) SearchChannel {
	return SearchChannel{
		config: conf,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		},
	}
}

// NewWith defines a function to return a pushers.Channel which delivers
// new messages to a giving underline ElasticSearch API defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig SearchConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.Host == "" {
		return nil, errors.New("elasticsearch.SearchConfig Invalid: Host can not be empty")
	}

	return New(apiconfig), nil
}

func init() {
	pushers.RegisterBackend("elasticsearch", NewWith)
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc SearchChannel) Send(messages ...message.Event) {
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
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s/%s", hc.config.Host, message.Sensor, message.Category, messageID.String()), buf)
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
