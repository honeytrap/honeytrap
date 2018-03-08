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
package web

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers/eventbus"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	assets "github.com/honeytrap/honeytrap-web"
	logging "github.com/op/go-logging"
	maxminddb "github.com/oschwald/maxminddb-golang"
)

var log = logging.MustGetLogger("web")

func AcceptAllOrigins(r *http.Request) bool { return true }

func download(url string, dest string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", geoLiteURL, nil)
	if err != nil {
		return err
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	gzf, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzf.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, gzf)
	if err != nil {
		return err
	}

	return nil
}

const geoLiteURL = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz"

type web struct {
	config *config.Config

	dataDir string

	ListenAddress string `toml:"listen"`
	Enabled       bool   `toml:"enabled"`

	eb *eventbus.EventBus

	start time.Time

	eventCh   chan event.Event
	messageCh chan json.Marshaler

	// Registered connections.
	connections map[*connection]bool

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	hotCountries *SafeArray
	events       *SafeArray
}

func New(options ...func(*web) error) (*web, error) {
	hc := web{
		eb: nil,

		start: time.Now(),

		ListenAddress: "127.0.0.1:8089",
		Enabled:       true,

		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),

		eventCh:   nil,
		messageCh: make(chan json.Marshaler),

		hotCountries: NewSafeArray(),
		events:       NewLimitedSafeArray(1000),
	}

	for _, optionFn := range options {
		if err := optionFn(&hc); err != nil {
			return nil, err
		}
	}

	return &hc, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (web *web) SetEventBus(eb *eventbus.EventBus) {
	eb.Subscribe(web)
}

func (web *web) Start() {
	if !web.Enabled {
		return
	}

	handler := http.NewServeMux()

	server := &http.Server{
		Addr:    web.ListenAddress,
		Handler: handler,
	}

	sh := http.FileServer(&assetfs.AssetFS{
		Asset:     assets.Asset,
		AssetDir:  assets.AssetDir,
		AssetInfo: assets.AssetInfo,
		Prefix:    assets.Prefix,
	})

	handler.HandleFunc("/ws", web.ServeWS)
	handler.Handle("/", sh)

	eventCh := make(chan event.Event)

	go func(ch chan event.Event) {
		for evt := range ch {
			web.events.Append(evt)

			web.messageCh <- Data("event", evt)

			isoCode := evt.Get("source.country.isocode")
			if isoCode == "" {
				continue
			}

			found := false

			web.hotCountries.Range(func(v interface{}) bool {
				hotCountry := v.(*HotCountry)

				if hotCountry.ISOCode != isoCode {
					return true
				}

				hotCountry.Last = time.Now()
				hotCountry.Count++

				found = true
				return false
			})

			if !found {
				web.hotCountries.Append(&HotCountry{
					ISOCode: isoCode,
					Count:   1,
					Last:    time.Now(),
				})
			}

			web.messageCh <- Data("hot_countries", web.hotCountries)
		}
	}(eventCh)

	eventCh = resolver(web.dataDir, eventCh)
	eventCh = filter(eventCh)

	web.eventCh = eventCh

	go web.run()

	go func() {
		log.Infof("Web interface started: %s", web.ListenAddress)

		server.ListenAndServe()
	}()
}

func (web *web) run() {
	for {
		select {
		case c := <-web.register:
			web.connections[c] = true
		case c := <-web.unregister:
			if _, ok := web.connections[c]; ok {
				delete(web.connections, c)

				close(c.send)
			}
		case msg := <-web.messageCh:
			for c := range web.connections {
				c.send <- msg
			}
		}
	}
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (msg Message) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	m["type"] = msg.Type
	m["data"] = msg.Data
	return json.Marshal(m)
}

func Data(type_ string, data interface{}) json.Marshaler {
	return &Message{
		Type: type_,
		Data: data,
	}
}

func filter(outCh chan event.Event) chan event.Event {
	ch := make(chan event.Event)
	go func() {
		for {
			evt := <-ch

			if category := evt.Get("category"); category == "heartbeat" {
				continue
			}

			outCh <- evt
		}
	}()

	return ch
}

func resolver(dataDir string, outCh chan event.Event) chan event.Event {
	db_path := path.Join(dataDir, "GeoLite2-Country.mmdb")

	_, err := os.Stat(db_path)
	if os.IsNotExist(err) {
		if err = download(geoLiteURL, db_path); err != nil {
			log.Fatal(err)
			return outCh
		} else {
		}
	}

	ch := make(chan event.Event)
	go func() {
		db, err := maxminddb.Open(db_path)
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		for {
			evt := <-ch

			v := evt.Get("source-ip")
			if v == "" {
				outCh <- evt
				continue
			}

			ip := net.ParseIP(v)

			var record struct {
				Country struct {
					ISOCode string `maxminddb:"iso_code"`
				} `maxminddb:"country"`
			}

			if err = db.Lookup(ip, &record); err != nil {
				log.Error("Error looking up country for: %s", err.Error())

				outCh <- evt
				continue
			}

			evt.Store("source.country.isocode", record.Country.ISOCode)
			outCh <- evt
		}
	}()

	return ch
}

func (web *web) Send(evt event.Event) {
	web.eventCh <- evt
}

func (web *web) ServeWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Could not upgrade connection: %s", err.Error())
		return
	}

	c := &connection{
		ws:   ws,
		web:  web,
		send: make(chan json.Marshaler, 100),
	}

	log.Info("Connection upgraded.")
	defer func() {
		c.web.unregister <- c
		c.ws.Close()

		log.Info("Connection closed")
	}()

	web.register <- c

	c.send <- Data("metadata", Metadata{
		Start:         web.start,
		Version:       cmd.Version,
		ReleaseTag:    cmd.ReleaseTag,
		CommitID:      cmd.CommitID,
		ShortCommitID: cmd.ShortCommitID,
	})

	c.send <- Data("events", web.events)
	c.send <- Data("hot_countries", web.hotCountries)

	go c.writePump()
	c.readPump()
}
