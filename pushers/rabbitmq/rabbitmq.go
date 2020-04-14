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
package rabbitmq

import (
	"fmt"
	"io"

	"encoding/json"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"github.com/streadway/amqp"
)

var (
	_ = pushers.Register("rabbitmq", New)
)

var (
	log = logging.MustGetLogger("rabbitmq")
)

type AMQPConfig struct {
	Address string `toml:"address"` // Like "amqp://guest:guest@localhost:5672/"
	Queue   string `toml:"queue"`
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	c := AMQPObject{
		AMQPConfig: AMQPConfig{},
		ch:         ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	if c.AMQPConfig.Address == "" {
		return nil, fmt.Errorf("AMQP address not set")
	}
	if c.AMQPConfig.Queue == "" {
		return nil, fmt.Errorf("AMQP queue not set")
	}

	conn, err := amqp.Dial(c.AMQPConfig.Address)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	c.amqpChannel = channel
	q, err := channel.QueueDeclare(
		c.AMQPConfig.Queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.queueName = q.Name

	return &c, nil
}

type AMQPObject struct {
	AMQPConfig
	io.Writer
	ch          chan map[string]interface{}
	amqpChannel *amqp.Channel
	queueName   string
}

func (b *AMQPObject) Send(e event.Event) {
	mp := make(map[string]interface{})

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})
	msg, err := json.Marshal(mp)
	if err != nil {
		log.Errorf("Failed to serialize event: %s", err.Error())
		return
	}
	err = b.amqpChannel.Publish(
		"",          // exchange
		b.queueName, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			Body:        msg,
		},
	)
	if err != nil {
		log.Errorf("Failed to send event: %s", err.Error())
		return
	}
}
