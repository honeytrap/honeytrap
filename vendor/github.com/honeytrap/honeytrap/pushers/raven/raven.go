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
package raven

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("raven", New)
)

var log = logging.MustGetLogger("channels/raven")

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	ch chan event.Event
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan event.Event, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	go c.run()

	return &c, nil
}

func Insecure(config *tls.Config) *tls.Config {
	config.InsecureSkipVerify = true
	return config
}

func (hc Backend) run() {
	tlsClientConfig := &tls.Config{}

	if hc.Insecure {
		tlsClientConfig = Insecure(tlsClientConfig)
	}

	d := &websocket.Dialer{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}

	for {
		func() {
			headers := http.Header{}
			headers.Set("User-Agent", fmt.Sprintf("Honeytrap/%s (%s; %s) %s", cmd.Version, runtime.GOOS, runtime.GOARCH, cmd.ShortCommitID))
			headers.Set("Authorization", fmt.Sprintf("Bearer %s", hc.Token))

			c, _, err := d.Dial(hc.Server, headers)
			if err != nil {
				log.Errorf("Error connecting to Raven server: %s: %s", hc.Server, err.Error())
				return
			}

			log.Debugf("Connected to Raven")
			defer log.Debugf("Connection to Raven lost")

			defer c.Close()

			readChan := make(chan []byte)

			go func(c *websocket.Conn) {
				defer close(readChan)

				for {
					_, r, err := c.ReadMessage()
					if err != nil {
						log.Errorf("Error received: %s", err.Error())
						return
					}
					readChan <- r
				}
			}(c)

			func(c *websocket.Conn) {
				for {
					select {
					case data, ok := <-readChan:
						if !ok {
							return
						}

						_ = data
					case evt := <-hc.ch:
						// we'll ignore heartbeats, those are generated within the protocol
						category := evt.Get("category")
						if category == "heartbeat" {
							continue
						}

						data, err := json.Marshal(evt)
						if err != nil {
							// handle errors
							log.Errorf("Error occurred while marshalling: %s", err.Error())
							continue
						}

						err = c.WriteMessage(websocket.BinaryMessage, data)
						if err != nil {
							log.Errorf("Could not write: %s", err.Error())
							return
						}
					}
				}
			}(c)
		}()

		time.Sleep(time.Second * 5)

		log.Info("Connection lost. Reconnecting in 5 seconds.")
	}
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc Backend) Send(message event.Event) {
	select {
	case hc.ch <- message:
	default:
		log.Errorf("Could not send more messages, channel full")
	}
}
