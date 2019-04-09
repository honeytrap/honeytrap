// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package marija

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	"io"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("marija", New)
)

var log = logging.MustGetLogger("channels:marija")

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	ch chan map[string]interface{}
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	go c.run()

	return &c, nil
}

func Insecure(config *tls.Config) *tls.Config {
	config.InsecureSkipVerify = true
	return config
}

func (hc Backend) run() {
	log.Debug("Marija channel started...")
	defer log.Debug("Marija channel stopped...")

	tlsClientConfig := &tls.Config{}

	if hc.Insecure {
		tlsClientConfig = Insecure(tlsClientConfig)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsClientConfig,
		},
	}

	docs := make([]map[string]interface{}, 0)

	send := func(docs []map[string]interface{}) {
		if len(docs) == 0 {
			return
		}

		pr, pw := io.Pipe()
		go func() {
			var err error

			defer pw.CloseWithError(err)

			for _, doc := range docs {
				err = json.NewEncoder(pw).Encode(doc)
				if err != nil {
					return
				}
			}
		}()

		req, err := http.NewRequest(http.MethodPost, hc.URL, pr)
		if err != nil {
			log.Errorf("Could create new request: %s", err.Error())
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("Could not submit event to Marija: %s", err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Errorf("Could not submit event to Marija: %d", resp.StatusCode)
			return
		}
	}

	for {
		select {
		case doc := <-hc.ch:
			docs = append(docs, doc)

			if len(docs) < 10 {
				continue
			}

			send(docs)

			docs = make([]map[string]interface{}, 0)
		case <-time.After(time.Second * 2):
			send(docs)

			docs = make([]map[string]interface{}, 0)
		}
	}
}

func filter(key string) bool {
	validKeys := []string{
		"source-ip",
		"destination-ip",
		"destination-port",
	}

	for _, vk := range validKeys {
		if vk == key {
			return false
		}
	}

	return true
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc Backend) Send(message event.Event) {
	mp := make(map[string]interface{})

	message.Range(func(key, value interface{}) bool {
		if filter(key.(string)) {
			return true
		}

		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}

		return true
	})

	select {
	case hc.ch <- mp:
	default:
		log.Errorf("Could not send more messages, channel full")
	}
}
