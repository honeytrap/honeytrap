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
