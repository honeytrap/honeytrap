package elasticsearch

import (
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

// SearchConfig defines a struct which holds configuration values for a SearchBackend.
type SearchConfig struct {
	Host string
}

// SearchBackend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type SearchBackend struct {
	client *http.Client
	config SearchConfig
}

// New returns a new instance of a SearchBackend.
func New(conf SearchConfig) SearchBackend {
	return SearchBackend{
		config: conf,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		},
	}
}

// NewWith defines a function to return a pushers.Backend which delivers
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
func (hc SearchBackend) Send(message message.Event) {
	log.Infof("ElasticSearchBackend: Sending %d actions.", message)

	category, _, sensor := message.Identity()

	if string(sensor) == "honeytrap" && string(category) == "ping" {
		// ignore
		return
	}

	messageID := uuid.NewV4()
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%+s/%+s/%s", hc.config.Host, sensor, category, messageID.String()), message.DataReader())
	if err != nil {
		log.Errorf("ElasticSearchBackend: Error while preparing request: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	if resp, err = hc.client.Do(req); err != nil {
		log.Errorf("ElasticSearchBackend: Error while sending messages: %s", err.Error())
		return
	}

	// for keep-alive
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Errorf("ElasticSearchBackend: Unexpected status code: %d", resp.StatusCode)
		return
	}

	log.Infof("ElasticSearchBackend: Sent %d actions.", message)
}
