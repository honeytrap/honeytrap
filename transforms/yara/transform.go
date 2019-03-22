package yara

import (
	"strings"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/transforms"
)

func Yara(source string) transforms.TransformFunc {
	c, err := NewCompiler()
	if err != nil {
		panic(err)
	}
	err = c.AddRulesFrom(source)
	if err != nil {
		panic(err)
	}
	m, err := NewMatcher(c)
	if err != nil {
		panic(err)
	}
	return func(state transforms.State, e event.Event, send func(event.Event)) {
		matches, err := m.GetMatches(e)
		if err != nil {
			panic(err)
		}
		for _, match := range matches {
			// Duplicate the event and add Yara metadata
			extendedEvt := event.New(
				event.MergeFrom(event.ToMap(e)),
				event.Custom("yara.rule", match.Rule),
				event.Custom("yara.tags", strings.Join(match.Tags, ",")),
			)
			send(extendedEvt)
		}
	}
}