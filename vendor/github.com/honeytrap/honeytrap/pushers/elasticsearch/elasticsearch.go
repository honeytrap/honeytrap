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
package elasticsearch

import (
	"context"

	"time"

	uuid "github.com/satori/go.uuid"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("elasticsearch", New)
)

var log = logging.MustGetLogger("channels/elasticsearch")

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	es *elastic.Client
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

	es, err := elastic.NewClient(
		c.options...,
	)
	if err != nil {
		return nil, err
	}

	c.es = es

	go c.run()

	return &c, nil
}

func (hc Backend) run() {
	log.Debug("Indexer started...")
	defer log.Debug("Indexer stopped...")

	bulk := hc.es.Bulk()

	count := 0
	for {
		select {
		case doc := <-hc.ch:
			messageID := uuid.NewV4()

			bulk = bulk.Add(elastic.NewBulkIndexRequest().
				Index(hc.index).
				Type("event").
				Id(messageID.String()).
				Doc(doc),
			)

			if bulk.NumberOfActions() < 10 {
				continue
			}
		case <-time.After(time.Second * 10):
		}

		if bulk.NumberOfActions() == 0 {
		} else if response, err := bulk.Do(context.Background()); err != nil {
			log.Errorf("Error indexing: %s", err.Error())
		} else {
			indexed := response.Indexed()
			count += len(indexed)

			for _, item := range response.Failed() {
				log.Errorf("Error indexing item: %s with error: %+v", item.Id, *item.Error)
			}

			log.Debugf("Bulk indexing: %d total %d", len(indexed), count)
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
