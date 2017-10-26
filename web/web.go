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
	"encoding/json"
	"net/http"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers/eventbus"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	assets "github.com/honeytrap/honeytrap-web"
	logging "github.com/op/go-logging"

	_ "github.com/mattn/go-sqlite3"
)

var log = logging.MustGetLogger("honeytrap:web")

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = 1 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

func AcceptAllOrigins(r *http.Request) bool { return true }

type web struct {
	*http.Server

	config *config.Config

	eb *eventbus.EventBus

	// Registered connections.
	connections map[*connection]bool

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func New(options ...func(*web)) *web {
	handler := http.NewServeMux()

	// TODO(nl5887): make configurable
	server := &http.Server{
		Addr:    ":8089",
		Handler: handler,
	}

	hc := web{
		Server: server,

		eb: nil,

		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}

	for _, optionFn := range options {
		optionFn(&hc)
	}

	log.Infof("Web interface started: %s", "8089")

	sh := http.FileServer(&assetfs.AssetFS{
		Asset:     assets.Asset,
		AssetDir:  assets.AssetDir,
		AssetInfo: assets.AssetInfo,
		Prefix:    assets.Prefix,
	})

	handler.HandleFunc("/ws", hc.ServeWS)
	handler.Handle("/", sh)

	go hc.run()

	/*
		db, err := sql.Open("sqlite3", "./foo.db")
		if err != nil {

		}

		_, err = db.Exec(`
		  CREATE TABLE 'events' (
		      'uid' INTEGER PRIMARY KEY AUTOINCREMENT,
		      'username' VARCHAR(64) NULL,
		      'departname' VARCHAR(64) NULL,
		      'created' DATE NULL
		  );`)

		tx, err := db.Begin()
		if err != nil {
		}

		defer tx.Commit()

		stmt, err := db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
		if err != nil {
		}

		res, err := stmt.Exec("astaxie", "研发部门", "2012-12-09")
		if err != nil {
		}

		id, err := res.LastInsertId()
		if err != nil {
		}

		_ = id
	*/

	return &hc
}

type Metadata struct {
	Start time.Time
}

func (metadata Metadata) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	m["start"] = metadata.Start
	return json.Marshal(m)
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
		}
	}
}

func (web *web) Send(e event.Event) {
	for c := range web.connections {
		c.send <- &e
	}
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
		send: make(chan json.Marshaler),
	}

	log.Info("Connection upgraded.")
	defer func() {
		c.web.unregister <- c
		c.ws.Close()

		log.Info("Connection closed")
	}()

	web.register <- c

	go func() {
		c.send <- Metadata{
			Start: time.Now(),
		}
	}()

	go c.writePump()
	c.readPump()
}
