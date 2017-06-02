package server

import (
	"time"

	"github.com/honeytrap/honeytrap/pushers/message"
)

// ping delivers a ping event to the server indicate it's alive.
func (hc *Honeytrap) ping() error {
	hc.events.Send(message.BasicEvent{
		Sensor: message.PingSensor,
		Type:   message.PingEvent,
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
