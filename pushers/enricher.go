package pushers

import "github.com/honeytrap/honeytrap/pushers/event"

// Enricher defines  struct which provides a series of
// data which will be applied to events which pass through
// it's call.
type Enricher struct {
	applications []event.Option
	channel      Channel
}

// NewEnricher returns a new instance of a Enricher.
func NewEnricher(channel Channel, options ...event.Option) *Enricher {
	return &Enricher{
		applications: options,
		channel:      channel,
	}
}

// Send delivers the provided event to the underline channel after
// applying the giving options.
func (e Enricher) Send(ev event.Event) {
	ev = event.Apply(ev, e.applications...)
	e.channel.Send(ev)
}
