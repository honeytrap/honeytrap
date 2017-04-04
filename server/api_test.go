package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/honeytrap/honeytrap/server"

	web "github.com/honeytrap/honeytrap-web"
)

const (
	passed = "\u2713"
	failed = "\u2717"
	dbName = "4534-pid"
)

var (
	conso = message.PushMessage{
		Sensor:      "Zu",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor:   "Rack",
			Category: "Wonderbat",
			Type:     message.ConnectionStarted,
		},
	}

	conlo = message.PushMessage{
		Sensor:      "Chip",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor:   "Fasmit",
			Category: "Wonderbat",
			Type:     message.ContainerClone,
		},
	}

	conco = message.PushMessage{
		Sensor:      "Cuj",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor:   "Crednur",
			Category: "Wonderbat",
			Type:     message.ConnectionClosed,
		},
	}
)

func TestHoneycast(t *testing.T) {
	conf := &config.Config{Token: dbName}
	cast := server.NewHoneycast(conf, &assetfs.AssetFS{
		Asset:     web.Asset,
		AssetDir:  web.AssetDir,
		AssetInfo: web.AssetInfo,
		Prefix:    web.Prefix,
	})

	defer os.Remove(dbName + "-bolted.db")

	server := httptest.NewServer(cast)

	cast.Send([]message.PushMessage{conso, conco, conlo})

	t.Logf("Given the an instance of a Honeycast API ")
	{
		t.Logf("\t When retrieving events from the /sessions endpoints")
		{

			req, err := http.NewRequest("GET", server.URL+"/sessions", nil)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created request.", passed)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully made request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully made request.", passed)

			defer res.Body.Close()

			var items []message.Event

			if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
				t.Fatalf("\t%s\t Should have successfully decoded response: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully decoded response.", passed)

			if len(items) != 2 {
				t.Fatalf("\t%s\t Should have retrieved 2 event for sessions: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have retrieved 2 event for sessions.", passed)
		}

		t.Logf("\t When retrieving events from the /events endpoints")
		{

			req, err := http.NewRequest("GET", server.URL+"/events", nil)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created request.", passed)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully made request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully made request.", passed)

			defer res.Body.Close()

			var items []message.Event

			if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
				t.Fatalf("\t%s\t Should have successfully decoded response: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully decoded response.", passed)

			if len(items) != 1 {
				t.Fatalf("\t%s\t Should have retrieved 1 event for sessions: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have retrieved 1 event for sessions.", passed)
		}

	}
}
