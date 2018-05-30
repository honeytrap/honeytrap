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
