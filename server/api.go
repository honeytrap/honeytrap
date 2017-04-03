package server

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/honeytrap/honeytrap/pushers/message"
)

// Honeycast defines a struct which exposes methods to handle api related service
// responses.
type Honeycast struct {
	*httptreemux.TreeMux
	assets http.Handler
}

// NewHoneycast returns a new instance of a Honeycast struct.
func NewHoneycast(assets *assetfs.AssetFS) *Honeycast {
	var hc Honeycast
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

// Deliver implements the pusher.EventDelivery interface and stores received events
// into the underline db.
func (h *Honeycast) Deliver(ev message.Event) {

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
