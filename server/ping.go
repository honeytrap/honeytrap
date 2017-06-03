package server

import (
	"time"

	"github.com/honeytrap/honeytrap/pushers/event"
)

// ping delivers a ping event to the server indicate it's alive.
func (hc *Honeytrap) ping() error {
	hc.events.Send(event.New(
		event.PingSensor,
		event.PingEvent,
	))
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
