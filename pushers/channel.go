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
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/event"
)

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send(event.Event)
}

type ChannelFunc func(...func(Channel) error) (Channel, error)

var (
	channels = map[string]ChannelFunc{}
)

func Range(fn func(string)) {
	for k := range channels {
		fn(k)
	}
}

func Register(key string, fn ChannelFunc) ChannelFunc {
	channels[key] = fn
	return fn
}

func WithConfig(c toml.Primitive) func(Channel) error {
	return func(d Channel) error {
		err := toml.PrimitiveDecode(c, d)
		return err
	}
}

func Get(key string) (ChannelFunc, bool) {
	d := Dummy

	if fn, ok := channels[key]; ok {
		return fn, true
	}

	return d, false
}
