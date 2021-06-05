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
package kafka

import (
	"encoding/json"

	sarama "github.com/Shopify/sarama"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("kafka", New)
)

var log = logging.MustGetLogger("channels/kafka")

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	producer sarama.AsyncProducer

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

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(c.Brokers, config)
	if err != nil {
		return nil, err
	}
	c.producer = producer

	go c.run()

	return &c, nil
}

func (hc Backend) run() {
	defer hc.producer.AsyncClose()

	for {
		data := <-hc.ch
		marshalledData, err := json.Marshal(data)
		if err != nil {
			log.Errorf("Error marshaling event: %s", err.Error())
			continue
		}

		hc.producer.Input() <- &sarama.ProducerMessage{
			Topic: hc.Topic,
			Key:   nil,
			Value: sarama.ByteEncoder(marshalledData),
		}

		select {
		case <-hc.producer.Successes():
		case msg := <-hc.producer.Errors():
			log.Errorf("Error producing event to kafka: %s", msg)
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
