package server

import "time"

func (hc *honeytrap) ping() error {
	hc.pusher.Push("honeytrap", "ping", "", "", nil)
	return nil
}

func (hc *honeytrap) startPing() {
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
