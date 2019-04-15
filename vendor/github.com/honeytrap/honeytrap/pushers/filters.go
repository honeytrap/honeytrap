// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pushers

import (
	"regexp"

	"github.com/honeytrap/honeytrap/event"
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

// FilterFunc defines a function for event filtering.
type FilterFunc func(event.Event) bool

// RegexFilterFunc returns a function for filtering event values.
func RegexFilterFunc(field string, expressions []string) FilterFunc {
	matchers := make([]*regexp.Regexp, len(expressions))

	for i, match := range expressions {
		matchers[i] = regexp.MustCompile(match)
	}

	return func(e event.Event) bool {
		for _, rx := range matchers {
			val := e.Get(field)
			// Only return on successful match, continue loop otherwise
			if rx.MatchString(val) {
				return true
			}
		}
		return false
	}
}

// FilterChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criteria.
func FilterChannel(channel Channel, fn FilterFunc) Channel {
	return filterChannel{
		Channel:  channel,
		FilterFn: fn,
	}
}

type tokenChannel struct {
	Channel

	Token string
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc tokenChannel) Send(e event.Event) {
	mc.Channel.Send(event.Apply(e, event.Token(mc.Token)))
}

// TokenChannel returns a Channel to set token value.
func TokenChannel(channel Channel, token string) Channel {
	return tokenChannel{
		Channel: channel,
		Token:   token,
	}
}
