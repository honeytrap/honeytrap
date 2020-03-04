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
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	sarama "github.com/Shopify/sarama"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/events"
	"io/ioutil"
	"unicode"
	"unicode/utf8"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("eventcollector", New)
	eventAPIStarted bool = false
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

	if !eventAPIStarted {
		go events.StartAPI()
		eventAPIStarted = true
	}

	ch := make(chan map[string]interface{}, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}


	config := sarama.NewConfig()

	if c.SecurityProtocol == "SSL" {

		tlsConfig, err := NewTLSConfig(c.SSLCertFile, c.SSLKeyFile, c.SSLCAFile)
		if err != nil {
			log.Errorf("Unable to create TLS configuration: %v", err)
			return nil, err
		}

		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true

		if len(c.SSLPassword) > 0 {
			config.Net.SASL.Password = c.SSLPassword
		}
	}

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



func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
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

		// process event
		_, evt, ok := events.ProcessEvent(e)
		if !ok {
			continue
		}

		// convert event to json
		eventJ, err := json.Marshal(evt)
		if err != nil {
			log.Errorf("Failed to marshall event: %v", err)
			continue
		}

		// send event to event collector broker
		message := &sarama.ProducerMessage{
			Topic: hc.Topic,
			Key:   nil,
			Value: sarama.StringEncoder(eventJ),
		}
		select {
		case hc.producer.Input() <- message:
			log.Debug("Produced message:\n", message)
		case err := <- hc.producer.Errors():
			log.Error("Failed to commit message: ", err)
		}
	}
}

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
