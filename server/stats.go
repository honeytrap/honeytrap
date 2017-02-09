package server

import "net/http"

type statsHandler struct {
}

func (th *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Honeytrap"))

	// template
}

func (hc *honeytrap) startStatsServer() {
	log.Infof("Stats server Listening on port: %s", hc.config.Web.Port)

	mux := http.NewServeMux()

	th := &statsHandler{}
	mux.Handle("/", th)

	go http.ListenAndServe(hc.config.Web.Port, mux)
}
