package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dimfeld/httptreemux"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"

	logging "github.com/op/go-logging"
)

// Contains the different buckets used
var (
	log           = logging.MustGetLogger("Honeytrap")
	sessionBucket = []byte("sessions")
	eventsBucket  = []byte("events")
)

// Honeycast defines a struct which exposes methods to handle api related service
// responses.
type Honeycast struct {
	*httptreemux.TreeMux
	assets http.Handler
	bolted *Bolted
	config *config.Config
}

// NewHoneycast returns a new instance of a Honeycast struct.
func NewHoneycast(config *config.Config, assets *assetfs.AssetFS) *Honeycast {

	// Create the database we desire.
	// TODO: Should we really panic here, it makes sense to do that, since it's the server
	// right?
	bolted, err := NewBolted(fmt.Sprintf("%s-bolted", config.Token), string(sessionBucket), string(eventsBucket))
	if err != nil {
		log.Errorf("Failed to created BoltDB session: %+q", err)
		panic(err)
	}

	var hc Honeycast
	hc.config = config
	hc.TreeMux = httptreemux.New()
	hc.bolted = bolted

	// Register endpoints for all handlers.
	if hc.assets != nil {
		hc.assets = http.FileServer(assets)
	}

	hc.TreeMux.Handle("GET", "/", hc.Index)
	hc.TreeMux.Handle("GET", "/events", hc.Events)
	hc.TreeMux.Handle("GET", "/sessions", hc.Sessions)

	return &hc
}

// Send delivers the underline provided messages and stores them into the underline
// Honeycast database for retrieval through the API.
func (h *Honeycast) Send(msgs []message.PushMessage) {
	var sessions, events []message.Event

	// Seperate out the event types appropriately.
	for _, msg := range msgs {
		if !msg.Event {
			continue
		}

		event, ok := msg.Data.(message.Event)
		if !ok {
			continue
		}

		switch event.Type {
		case message.ConnectionStarted, message.ConnectionClosed:
			sessions = append(sessions, event)
		default:
			events = append(events, event)
		}
	}

	//  Batch save the events received for both sessions and events.
	if terr := h.bolted.Save(sessionBucket, sessions...); terr != nil {
		log.Errorf("honeycast : Failed to save session events to db: %+q", terr)
	}

	if terr := h.bolted.Save(eventsBucket, events...); terr != nil {
		log.Errorf("honeycast : Failed to save events to db: %+q", terr)
	}
}

// Sessions handles response for all `/sessions` target endpoint and returns all giving push
// messages returns the slice of data.
func (h *Honeycast) Sessions(w http.ResponseWriter, r *http.Request, params map[string]string) {
	totalInt := -1
	fromInt := -1

	if total, ok := params["total"]; ok {
		var err error
		totalInt, err = strconv.Atoi(total)
		if err != nil {
			log.Error("honeycast : Invalid Total Param: %+q", err)
			http.Error(w, "Invalid 'Total' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if from, ok := params["from"]; ok {
		var err error
		fromInt, err = strconv.Atoi(from)
		if err != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", err)
			http.Error(w, "Invalid 'From' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.bolted.Get(sessionBucket, fromInt, totalInt, func(result []message.Event) {

	})
}

// Events handles response for all `/events` target endpoint and returns all giving events
// and expects a giving filter paramter which will be used to filter out the needed events.
func (h *Honeycast) Events(w http.ResponseWriter, r *http.Request, params map[string]string) {
	totalInt := -1
	fromInt := -1

	if total, ok := params["total"]; ok {
		var err error
		totalInt, err = strconv.Atoi(total)
		if err != nil {
			log.Error("honeycast : Invalid Total Param: %+q", err)
			http.Error(w, "Invalid 'Total' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if from, ok := params["from"]; ok {
		var err error
		fromInt, err = strconv.Atoi(from)
		if err != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", err)
			http.Error(w, "Invalid 'From' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.bolted.Get(sessionBucket, fromInt, totalInt, func(result []message.Event) {

	})
}

// Index handles the servicing of index based requests for the giving service.
func (h *Honeycast) Index(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if h.assets != nil {
		h.assets.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
