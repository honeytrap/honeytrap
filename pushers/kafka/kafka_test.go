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
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

const config = `
[P]
brokers = ["%s"]
topic = "my_topic"
`

func TestChannelsKafkaSend(t *testing.T) {
	seedBroker := sarama.NewMockBroker(t, 1)
	defer seedBroker.Close()

	leader := sarama.NewMockBroker(t, 2)
	defer leader.Close()

	metadataResponse := new(sarama.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, sarama.ErrNoError)

	seedBroker.Returns(metadataResponse)

	prodSuccess := new(sarama.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, sarama.ErrNoError)

	leader.Returns(prodSuccess)

	s := struct {
		P toml.Primitive
	}{}

	md, err := toml.Decode(fmt.Sprintf(config, seedBroker.Addr()), &s)

	if err != nil {
		t.Error(err)
	}

	c, err := New(
		pushers.WithConfig(s.P, &md),
	)

	if err != nil {
		t.Error(err)
	}

	c.Send(event.New())

	kb := c.(*Backend)

	select {
	case <-kb.producer.Successes():
	case msg := <-kb.producer.Errors():
		t.Error(msg.Err)
	}

	close(kb.ch)

}
