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
package splunk

import (
	"net/http"

	"time"

	hec "github.com/fuyufjh/splunk-hec-go"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
	"github.com/honeytrap/honeytrap/storers"
)

var (
	_ = pushers.Register("splunk", New)
)

var log = logging.MustGetLogger("channels:splunk")

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	ch chan map[string]interface{}
	Storer storers.Storer
}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	go c.run()

	return &c, nil
}

func (hc Backend) run() {
	log.Debug("Splunk indexer started...")
	defer log.Debug("Splunk indexer stopped...")

	client := hec.NewCluster(
		hc.Config.Endpoints,
		hc.Config.Token,
	)

	client.SetHTTPClient(&http.Client{Transport: &http.Transport{
		TLSClientConfig: hc.tlsConfig,
	}})

	batch := []*hec.Event{}

	count := 0
	for {
		select {
		case doc := <-hc.ch:
			event := hec.NewEvent(doc)
			event.SetTime(time.Now())

			batch = append(batch, event)
			if len(batch) < 10 {
				continue
			}
		case <-time.After(time.Second * 10):
		}

		if len(batch) == 0 {
			continue
		}

		if err := client.WriteBatch(batch); err != nil {
			log.Errorf("Error indexing: %s", err.Error())
		} else {
			count += len(batch)

			log.Infof("Bulk indexing: %d total %d", len(batch), count)

			batch = []*hec.Event{}
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

// Must be called as a goroutine in order to be non-blocking
func (hc Backend) SendFile(file []byte) {
	ref := hc.Storer.Push(file)
	hc.ch <- ref.ToMap()
}

func (hc Backend) SetStorer(storer storers.Storer) {
	hc.Storer = storer
}