package server

import (
	"net/http"

	web "github.com/honeytrap/honeytrap-web"

	"fmt"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	// logging "github.com/op/go-logging"
	"os"
)

// startStatsServer starts the http server for handling request.
func (hc *Honeytrap) startStatsServer() {
	log.Infof("Stats server Listening on port: %s", hc.config.Web.Port)

	staticHandler := http.FileServer(
		&assetfs.AssetFS{
			Asset:    web.Asset,
			AssetDir: web.AssetDir,
			AssetInfo: func(path string) (os.FileInfo, error) {
				return os.Stat(path)
			},
			Prefix: web.Prefix,
		})

	mux := http.NewServeMux()

	mux.Handle("/", staticHandler)

	if hc.config.Web.Path != "" {
		log.Debug("Using static file path: ", hc.config.Web.Path)

		// check local css first
		// TODO: What is this for and why are we assigning here.
		staticHandler = http.FileServer(http.Dir(hc.config.Web.Path))
	}

	fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap server started, listening on address %s.", hc.config.Web.Port)))

	defer func() {
		fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap server stopped.")))
	}()

	if err := http.ListenAndServe(hc.config.Web.Port, mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
