package pushers

import (
	"regexp"

	"github.com/honeytrap/honeytrap/pushers/event"
)

//==========================================================================================

// Filter defines an interface which exposes a method for filtering specific
// messages by specific boundaries.
type Filter interface {
	Filter(...event.Event) []event.Event
}

//==========================================================================================

// FilterGroup defines a slice of Filter object which used together to filter
// a series of events.
type FilterGroup []Filter

// Add adds the underline filter into the group.
func (fg *FilterGroup) Add(filter Filter) {
	*fg = append(*fg, filter)
}

// Filter returns a slice of messages that match the giving criterias from the
// provided events.
func (fg FilterGroup) Filter(events ...event.Event) []event.Event {
	if len(fg) == 0 {
		return events
	}

	for _, filter := range fg {
		events = filter.Filter(events...)
	}

	return events
}

//==========================================================================================

// RegExpFilterFunction defines the function used by the RegExpFilter
// to provide custom filtering validation for each provided event.Event.
type RegExpFilterFunction func(*regexp.Regexp, event.Event) bool

// SensorFilterFunc defines a function to validate a Pushevent.Sensor value
// based on a provided regular expression.
func SensorFilterFunc(rx *regexp.Regexp, message event.Event) bool {
	return rx.MatchString(message["sensor"].(string))
}

// TypeFilterFunc defines a function to validate a Pushevent.Category value
// based on a provided regular expression.
func TypeFilterFunc(rx *regexp.Regexp, message event.Event) bool {
	return rx.MatchString(message["type"].(string))
}

// CategoryFilterFunc defines a function to validate a Pushevent.Category value
// based on a provided regular expression.
func CategoryFilterFunc(rx *regexp.Regexp, message event.Event) bool {
	return rx.MatchString(message["category"].(string))
}

//==========================================================================================

// RegExpFilter defines a struct which implements the Filters interface and
// provides the ability to filter by a provided set of regular expression
// and a function which runs down all provided messages
type RegExpFilter struct {
	conditions []*regexp.Regexp
	validator  RegExpFilterFunction
}

// NewRegExpFilter returns a new instance of a RegExpFilter with the provided validator
// and regexp.Regexp matchers.
func NewRegExpFilter(fn RegExpFilterFunction, rx ...*regexp.Regexp) *RegExpFilter {
	var rxFilter RegExpFilter
	rxFilter.validator = fn
	rxFilter.conditions = rx
	return &rxFilter
}

// Filter returns a slice of messages passed in which passes the internal regular
// expressions criterias.
func (r *RegExpFilter) Filter(messages ...event.Event) []event.Event {
	if r.conditions == nil || len(r.conditions) == 0 {
		return messages
	}

	var filtered []event.Event

	{
	mloop:
		for _, message := range messages {
			for _, rx := range r.conditions {
				if !r.validator(rx, message) {
					continue mloop
				}
			}

			filtered = append(filtered, message)
		}
	}

	return filtered
}

//==========================================================================================

// MakeMatchers takes a giving slice of strings and returns a slice of regexp.Regexp.
func MakeMatchers(m ...string) []*regexp.Regexp {
	var matchers []*regexp.Regexp

	for _, match := range m {
		matchers = append(matchers, regexp.MustCompile(match))
	}

	return matchers
}
