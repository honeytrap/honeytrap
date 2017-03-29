package server

import (
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ping delivers a ping event to the server indicate it's alive.
func (hc *Honeytrap) ping() error {
	hc.events.Deliver(message.Event{
		Sensor:   "Ping",
		Category: "Server",
		Type:     message.Ping,
	})
	return nil
}

// startPing initializes the ping runner.
func (hc *Honeytrap) startPing() {
	go func() {
		for {
			log.Debug("Yep, still alive")

			if err := hc.ping(); err != nil {
				log.Error(err.Error())
			}

			<-time.After(time.Second * 60)
		}
	}()
}
