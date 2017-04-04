package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dimfeld/httptreemux"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
)

// Contains values for use.
const (
	ResponsePerPageHeader = "response_per_page"
	PageHeader            = "page"
)

// Contains the different buckets used
var (
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
	hc.TreeMux.Handle("GET", "/ws", hc.Websocket)
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

// Websocket handles response for all `/sessions` target endpoint and returns all giving push
// messages returns the slice of data.
func (h *Honeycast) Websocket(w http.ResponseWriter, r *http.Request, params map[string]string) {
}

// Sessions handles response for all `/sessions` target endpoint and returns all giving push
// messages returns the slice of data.
func (h *Honeycast) Sessions(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var responsePerPage int
	var pageNumber int

	total, err := h.bolted.GetSize(sessionBucket)
	if err != nil {
		log.Error("honeycast : Operation Failed : %+q", err)
		http.Error(w, "Operation Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if perpage, ok := params[ResponsePerPageHeader]; ok {
		var err error

		responsePerPage, err = strconv.Atoi(perpage)
		if err != nil {
			log.Error("honeycast : Invalid ResponsePerPage Param: %+q", err)
			http.Error(w, "Invalid 'ResponsePerPage' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		responsePerPage = -1
	}

	if page, ok := params[PageHeader]; ok {
		var err error

		pageNumber, err = strconv.Atoi(page)
		if err != nil {
			log.Error("honeycast : Invalid 'Page' Param: %+q", err)
			http.Error(w, "Invalid 'From' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		pageNumber = -1
	}

	var items []message.Event
	var terr error

	if responsePerPage == -1 || pageNumber == -1 {
		items, terr = h.bolted.Get(sessionBucket, -1, -1)
		if terr != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", terr)
			http.Error(w, "Invalid 'From' parameter: "+terr.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		length := responsePerPage * pageNumber

		if length >= total {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		items, terr = h.bolted.Get(sessionBucket, (length/2)+1, length)
		if terr != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", terr)
			http.Error(w, "Invalid 'From' parameter: "+terr.Error(), http.StatusInternalServerError)
			return
		}

	}

	var bu bytes.Buffer
	if jserr := json.NewEncoder(&bu).Encode(items); jserr != nil {
		log.Error("honeycast : Invalid 'From' Param: %+q", jserr)
		http.Error(w, "Invalid 'From' parameter: "+jserr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bu.Bytes())
}

// Events handles response for all `/events` target endpoint and returns all giving events
// and expects a giving filter paramter which will be used to filter out the needed events.
func (h *Honeycast) Events(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var responsePerPage int
	var pageNumber int

	total, err := h.bolted.GetSize(sessionBucket)
	if err != nil {
		log.Error("honeycast : Operation Failed : %+q", err)
		http.Error(w, "Operation Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if perpage, ok := params[ResponsePerPageHeader]; ok {
		var err error

		responsePerPage, err = strconv.Atoi(perpage)
		if err != nil {
			log.Error("honeycast : Invalid ResponsePerPage Param: %+q", err)
			http.Error(w, "Invalid 'ResponsePerPage' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		responsePerPage = -1
	}

	if page, ok := params[PageHeader]; ok {
		var err error

		pageNumber, err = strconv.Atoi(page)
		if err != nil {
			log.Error("honeycast : Invalid 'Page' Param: %+q", err)
			http.Error(w, "Invalid 'From' parameter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		pageNumber = -1
	}

	var items []message.Event
	var terr error

	if responsePerPage == -1 || pageNumber == -1 {

		items, terr = h.bolted.Get(eventsBucket, -1, -1)
		if terr != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", terr)
			http.Error(w, "Invalid 'From' parameter: "+terr.Error(), http.StatusInternalServerError)
			return
		}

	} else {

		length := responsePerPage * pageNumber

		if length >= total {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		items, terr = h.bolted.Get(eventsBucket, (length/2)+1, length)
		if terr != nil {
			log.Error("honeycast : Invalid 'From' Param: %+q", terr)
			http.Error(w, "Invalid 'From' parameter: "+terr.Error(), http.StatusInternalServerError)
			return
		}

	}

	var bu bytes.Buffer
	if jserr := json.NewEncoder(&bu).Encode(items); jserr != nil {
		log.Error("honeycast : Invalid 'From' Param: %+q", jserr)
		http.Error(w, "Invalid 'From' parameter: "+jserr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bu.Bytes())
}

// Index handles the servicing of index based requests for the giving service.
func (h *Honeycast) Index(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if h.assets != nil {
		h.assets.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
