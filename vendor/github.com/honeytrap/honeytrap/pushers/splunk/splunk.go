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
package splunk

import (
	"net/http"

	"time"

	hec "github.com/fuyufjh/splunk-hec-go"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("splunk", New)
)

var log = logging.MustGetLogger("channels:splunk")

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

func (hc Backend) run() {
	log.Debug("Splunk indexer started...")
	defer log.Debug("Splunk indexer stopped...")

	client := hec.NewCluster(
		hc.Config.Endpoints,
		hc.Config.Token,
	)

	client.SetHTTPClient(&http.Client{Transport: &http.Transport{
		TLSClientConfig: hc.tlsConfig,
	}})

	batch := []*hec.Event{}

	count := 0
	for {
		select {
		case doc := <-hc.ch:
			event := hec.NewEvent(doc)
			event.SetTime(time.Now())

			batch = append(batch, event)
			if len(batch) < 10 {
				continue
			}
		case <-time.After(time.Second * 10):
		}

		if len(batch) == 0 {
			continue
		}

		if err := client.WriteBatch(batch); err != nil {
			log.Errorf("Error indexing: %s", err.Error())
		} else {
			count += len(batch)

			log.Infof("Bulk indexing: %d total %d", len(batch), count)

			batch = []*hec.Event{}
		}
	}
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc Backend) Send(message event.Event) {
	mp := make(map[string]interface{})

	message.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	hc.ch <- mp
}
