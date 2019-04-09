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
package pulsar

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/websocket"

	"encoding/base64"
	"encoding/json"

	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
)

var (
	_ = pushers.Register("pulsar", New)
)

var (
	log = logging.MustGetLogger("pulsar")
)

type Message struct {
	Properties map[string]string
	Data       []byte
	ID         string
	Key        string
	Timestamp  time.Time
}

type producerMessage struct {
	Payload             string            `json:"payload"`
	Properties          map[string]string `json:"properties,omitempty"`
	Context             *string           `json:"context,omitempty"`
	Key                 *string           `json:"key,omitempty"`
	ReplicationClusters []string          `json:"replicationClusters,omitempty"`
}

type producerAckMessage struct {
	Result    string  `json:"result"`
	MessageID *string `json:"messageId"`
	ErrorMsg  *string `json:"errorMsg"`
	Context   *string `json:"context"`
}

type producer struct {
	Config

	ch        chan Message
	ws        *websocket.Conn
	queueName string
}

type Config struct {
	URL string `toml:"url"` // Like "amqp://guest:guest@localhost:5672/"

	Insecure bool `toml:"insecure"`
}

func Insecure(config *tls.Config) *tls.Config {
	config.InsecureSkipVerify = true
	return config
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan Message, 100)

	p := producer{
		Config: Config{},
		ch:     ch,
	}

	for _, optionFn := range options {
		optionFn(&p)
	}

	if p.URL == "" {
		return nil, fmt.Errorf("Pulsar producer URL not set")
	}

	go p.run()

	return &p, nil
}

func (p *producer) run() {
	tlsClientConfig := &tls.Config{}

	if p.Insecure {
		tlsClientConfig = Insecure(tlsClientConfig)
	}

	d := &websocket.Dialer{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}

	headers := http.Header{}
	headers.Set("User-Agent", fmt.Sprintf("Honeytrap/%s (%s; %s) %s", cmd.Version, runtime.GOOS, runtime.GOARCH, cmd.ShortCommitID))

	for {
		func() {
			log.Infof("Connecting to Pulsar...")

			ws, _, err := d.Dial(p.URL, headers)
			if err != nil {
				log.Errorf("Error connecting to Pulsar: %s: %s", p.URL, err.Error())
				return
			}

			log.Infof("Connected to Pulsar...")

			done := make(chan struct{})

			go func(c *websocket.Conn) {
				defer close(done)

				for {
					ack := producerAckMessage{}
					if err := ws.ReadJSON(&ack); err == io.EOF {
						break
					} else if err != nil {
						log.Errorf("Error reading message: %s", err.Error())
						break
					} else if ack.Result != "ok" {
						log.Errorf("Error ack result: %s", *ack.ErrorMsg)
					}
				}
			}(ws)

			for {
				select {
				case <-done:
					return
				case m := <-p.ch:
					if err := ws.WriteJSON(producerMessage{
						Payload:    base64.StdEncoding.EncodeToString(m.Data),
						Properties: m.Properties,
						Key:        &m.Key,
					}); err != nil {
						log.Errorf("Error writing message: %s", err.Error())
						continue
					}
				}
			}
		}()

		time.Sleep(time.Second * 5)

		log.Errorf("Connection lost. Reconnecting in 5 seconds.")
	}
}

func (p *producer) Send(e event.Event) {
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

	p.ch <- Message{
		Data: msg,
	}
}
