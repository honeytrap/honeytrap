package pushers

import (
	"regexp"

	"github.com/honeytrap/honeytrap/pushers/event"
)

//==========================================================================================

// Filter defines an interface which exposes a method for filtering specific
// messages by specific boundaries.
type Filter interface {
	Filter(event.Event) bool
}

type filterChannel struct {
	Channel

	FilterFn FilterFunc
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc filterChannel) Send(e event.Event) {
	if !mc.FilterFn(e) {
		return
	}

	mc.Channel.Send(e)
}

type FilterFunc func(event.Event) bool

func RegexFilterFunc(field string, expressions []string) FilterFunc {
	matchers := make([]*regexp.Regexp, len(expressions))

	for i, match := range expressions {
		matchers[i] = regexp.MustCompile(match)
	}

	return func(e event.Event) bool {
		for _, rx := range matchers {
			val := e.Get(field)
			return rx.MatchString(val)
		}

		return false
	}
}

// FilterChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criterias.
func FilterChannel(channel Channel, fn FilterFunc) Channel {
	return filterChannel{
		Channel:  channel,
		FilterFn: fn,
	}
}
