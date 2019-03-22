/*
This is an example of a Yara-based plugin.

It matches a few known patterns for UDP amplification attacks.
*/

package main

import (
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/transforms"
	"github.com/honeytrap/honeytrap/transforms/yara"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("udp-ampl-detector")

func Transform() transforms.TransformFunc {
	matcher, err := yara.NewMatcherFrom("/honeytrap/assets/yara-custom/amplification.yara")
	if err != nil {
		panic(err)
	}
	return func(state transforms.State, e event.Event, send func(event.Event)) {
		matches, err := matcher.GetMatches(e)
		if err != nil {
			log.Error(err.Error())
			return
		}

		// Send the original event, as well as an additional event for each match
		send(e)
		for _, match := range matches {
			send(event.New(
				event.Sensor("yara-custom"),
				event.Category("amplification"),
				event.Type(match.Rule),
			))
		}
	}
}
