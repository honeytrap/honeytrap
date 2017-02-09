package pushers

import (
	"bytes"
	"encoding/json"
	"fmt"

	pushers "github.com/honeytrap/honeytrap/pushers"

	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:elasticsearch")

var (
	_ = pushers.RegisterChannel("elasticsearch", NewElasticSearchChannel())
)

type ElasticSearchChannel struct {
	client *http.Client
	host   string
}

func NewElasticSearchChannel() pushers.ChannelFunc {
	return func(conf map[string]interface{}) (pushers.Channel, error) {
		hc := &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
			},
			Timeout: time.Duration(20) * time.Second,
		}

		esc := ElasticSearchChannel{
			client: hc,
		}

		var ok bool
		if esc.host, ok = conf["host"].(string); !ok {
			return nil, fmt.Errorf("Host not set for channel elasticsearch")
		}

		return &esc, nil
	}
}

func (hc ElasticSearchChannel) Send(messages []*pushers.PushMessage) {
	for _, message := range messages {
		buf := new(bytes.Buffer)

		if message.Sensor == "honeytrap" && message.Category == "ping" {
			// ignore
			continue
		}

		if err := json.NewEncoder(buf).Encode(message.Data); err != nil {
			log.Error("ElasticSearchChannel: Error encoding data: %s", err.Error())
			continue
		}

		messageID := uuid.NewV4()
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s/%s", hc.host, message.Sensor, message.Category, messageID.String()), buf)
		if err != nil {
			log.Error("ElasticSearchChannel: Error while preparing request: %s", err.Error())
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		var resp *http.Response
		if resp, err = hc.client.Do(req); err != nil {
			log.Error("ElasticSearchChannel: Error while sending messages: %s", err.Error())
			continue
		}

		// for keep-alive
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			log.Error("ElasticSearchChannel: Unexpected status code: %d", resp.StatusCode)
			continue
		}

	}

	log.Infof("ElasticSearchChannel: Sent %d actions.", len(messages))
}
