package filters

import (
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/transforms"
	"regexp"
)

// FieldRegex filters by applying a regex to a field.
func FieldRegex(field string, expressions []string) transforms.TransformFunc {
	matchers := make([]*regexp.Regexp, len(expressions))

	for i, match := range expressions {
		matchers[i] = regexp.MustCompile(match)
	}

	return func(e event.Event) []event.Event {
		for _, rx := range matchers {
			val := e.Get(field)
			// Only return on successful match, continue loop otherwise
			if rx.MatchString(val) {
				return []event.Event{e}
			}
		}
		return []event.Event{}
	}
}
