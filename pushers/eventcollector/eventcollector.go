// Copyright 2019 Ubiwhere (https://www.ubiwhere.com/)
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

package eventcollector

import (
	"encoding/hex"
	"fmt"
	sarama "github.com/Shopify/sarama"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/events"
	"unicode"
	"unicode/utf8"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("eventcollector", New)
)

var log = logging.MustGetLogger("channels/eventcollector")

// Backend defines a struct which provides a channel for delivery
// push messages to the EventCollector brokers (kafka).
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
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.NoResponse

	producer, err := sarama.NewAsyncProducer(c.Brokers, config)
	if err != nil {
		return nil, err
	}

	c.producer = producer

	go c.run()
	return &c, nil
}

func printify(s string) string {
	o := ""
	for _, rune := range s {
		if !unicode.IsPrint(rune) {
			buf := make([]byte, 4)

			n := utf8.EncodeRune(buf, rune)
			o += fmt.Sprintf("\\x%s", hex.EncodeToString(buf[:n]))
			continue
		}

		o += string(rune)
	}
	return o
}

func (hc Backend) run() {
	defer hc.producer.AsyncClose()

	for e := range hc.ch {
		//var params []string

		procEvent := events.ProcessEvent(e)

/*		for k, v := range e {

			switch x := v.(type) {
			case net.IP:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case uint32, uint16, uint8, uint,
				int32, int16, int8, int:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case time.Time:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case string:
				params = append(params, fmt.Sprintf("%s=%s", k, printify(x)))
			default:
				params = append(params, fmt.Sprintf("%s=%#v", k, v))
			}
		}
		sort.Strings(params)*/
		message := &sarama.ProducerMessage{
			Topic: hc.Topic,
			Key:   nil,
			Value: sarama.StringEncoder(procEvent),
		}
		select {
		case hc.producer.Input() <- message:
			log.Debug("Produced message:\n", message)
		case err := <- hc.producer.Errors():
			log.Error("Failed to commit message: ", err)
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

