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
package rabbitmq

import (
	"fmt"
	"io"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"github.com/streadway/amqp"
	"encoding/json"
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

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
		"", // exchange
		b.queueName, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			Body: msg,
		},
	)
	if err != nil {
		log.Errorf("Failed to send event: %s", err.Error())
		return
	}
}
