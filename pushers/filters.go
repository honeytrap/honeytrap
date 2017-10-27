/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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
			return rx.MatchString(val)
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
