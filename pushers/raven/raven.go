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
	"time"

	"github.com/gorilla/websocket"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("raven", New)
)

var log = logging.MustGetLogger("channels/raven")

// RavenBackend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type RavenBackend struct {
	Config

	ch chan event.Event
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan event.Event, 100)

	c := RavenBackend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	go c.run()

	return &c, nil
}

func (hc RavenBackend) run() {
	d := &websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			//              RootCAs: capool,
			InsecureSkipVerify: true,
		},
	}

	for {
		func() {
			headers := http.Header{}
			headers.Set("Authorization", fmt.Sprintf("Bearer %s", hc.Token))

			c, _, err := d.Dial(hc.Server, headers)
			if err != nil {
				log.Error("dial:", err.Error())
				return
			}

			log.Debugf("Connected to Raven")
			defer log.Debugf("Connection to Raven lost")

			defer c.Close()

			readChan := make(chan []byte)

			go func(c *websocket.Conn) {
				defer close(readChan)

				for {
					if _, r, err := c.ReadMessage(); err != nil {
						log.Errorf("Error received: %s", err.Error())
						return
					} else {
						readChan <- r
					}
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
						if category := evt.Get("category"); category == "heartbeat" {
						} else if data, err := json.Marshal(evt); err != nil {
							// handle errors
							log.Errorf("Error occurred while marshalling: %s", err.Error())
							continue
						} else if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
							log.Errorf("Could not write: %s", err.Error())
							return
						} else {
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
func (hc RavenBackend) Send(message event.Event) {
	hc.ch <- message
}
