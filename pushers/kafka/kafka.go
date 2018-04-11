/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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

	for doc := range <-hc.ch {
		data, err := json.Marshal(doc)
		if err != nil {
			log.Errorf("Error marshaling event: %s", err.Error())
			continue
		}

		hc.producer.Input() <- &sarama.ProducerMessage{
			Topic: hc.Topic,
			Key:   nil,
			Value: sarama.ByteEncoder(data),
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
