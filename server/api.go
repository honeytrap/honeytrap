package server

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
)

// Honeycast defines a struct which exposes methods to handle api related service
// responses.
type Honeycast struct {
	*httptreemux.TreeMux
	assets http.Handler
	config *config.Config
}

// NewHoneycast returns a new instance of a Honeycast struct.
func NewHoneycast(config *config.Config, assets *assetfs.AssetFS) *Honeycast {
	var hc Honeycast
	hc.config = config
	hc.TreeMux = httptreemux.New()

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
	for _, msg := range msgs {
		if !msg.Event {
			continue
		}

		event, ok := msg.Data.(message.Event)
		if !ok {
			continue
		}

		switch event.Type {
		case message.ConnectionStarted:
		case message.ConnectionClosed:
		}
	}
}

// Sessions handles response for all `/sessions` target endpoint and returns all giving push
// messages returns the slice of data.
func (h *Honeycast) Sessions(w http.ResponseWriter, r *http.Request, params map[string]string) {

}

// Events handles response for all `/events` target endpoint and returns all giving events
// and expects a giving filter paramter which will be used to filter out the needed events.
func (h *Honeycast) Events(w http.ResponseWriter, r *http.Request, params map[string]string) {

}

// Index handles the servicing of index based requests for the giving service.
func (h *Honeycast) Index(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if h.assets != nil {
		h.assets.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
