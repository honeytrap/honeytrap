package yara

import (
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/services"
	"testing"
)

func TestAlwaysTrue(t *testing.T) {
	rule := "rule foo { condition: true }"
	c, err := NewCompiler()
	if err != nil {
		panic(err)
	}
	err = c.AddRulesFrom(rule)
	if err != nil {
		panic(err)
	}
	m, err := NewMatcher(c)
	if err != nil {
		panic(err)
	}
	evts := []event.Event{
		event.New(),
		event.New(services.EventOptions),
		event.New(event.Custom("foo", "")),
		event.New(event.Custom("foo", "bar")),
	}
	for _, evt := range evts {
		if !m.MustMatch(evt) {
			t.Fail()
		}
	}
}

func TestAlwaysFalse(t *testing.T) {
	rule := "rule foo { condition: false }"
	c, err := NewCompiler()
	if err != nil {
		panic(err)
	}
	err = c.AddRulesFrom(rule)
	if err != nil {
		panic(err)
	}
	m, err := NewMatcher(c)
	if err != nil {
		panic(err)
	}
	evts := []event.Event{
		event.New(),
		event.New(services.EventOptions),
		event.New(event.Custom("foo", "")),
		event.New(event.Custom("foo", "bar")),
	}
	for _, evt := range evts {
		if m.MustMatch(evt) {
			t.Fail()
		}
	}
}