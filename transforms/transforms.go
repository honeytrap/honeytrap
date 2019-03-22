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
package transforms

import (
	"fmt"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/plugins"
	"github.com/honeytrap/honeytrap/pushers"
)

type transformChannel struct {
	destination pushers.Channel
	fn          TransformFunc
	connState   map[string]State // Map<IP, Map<Key, Value>>
	// todo(capacitorset): implement an expiration policy for old data
}

func (c transformChannel) Send(input event.Event) {
	var state State
	if input.Has("source-ip") {
		ip := input.Get("source-ip")
		if _, ok := c.connState[ip]; !ok {
			c.connState[ip] = make(State)
		}
		state = c.connState[ip]
	} else {
		// There's no connection state if there's no connection, so pass an empty state
		// Plugins shouldn't use the state for non-connection events anyway.
		state = make(State)
	}
	c.fn(state, input, c.destination.Send)
}

func Transform(dest pushers.Channel, fn TransformFunc) pushers.Channel {
	return transformChannel{destination: dest, fn: fn, connState: make(map[string]State)}
}

type State map[string]interface{}

type TransformFunc func(state State, e event.Event, send func(event.Event))

var staticTransforms = make(map[string]TransformFunc)

// Registers a static transform.
func Register(name string, fn TransformFunc) int {
	staticTransforms[name] = fn
	// The return value is unused, but it allows for `var _ = Register("name", handler)`
	return 0
}

// Gets a static or dynamic transform, giving priority to static ones.
func Get(name, folder string) (TransformFunc, error) {
	staticPl, ok := staticTransforms[name]
	if ok {
		return staticPl, nil
	}

	// todo: add Lua support (issue #272)
	sym, found, err := plugins.Get(name, "Transform", folder)
	if !found {
		return nil, fmt.Errorf("Transform %s not found", name)
	}
	if err != nil {
		return nil, err
	}
	return sym.(func() TransformFunc)(), nil
}

func MustGet(name, folder string) TransformFunc {
	out, err := Get(name, folder)
	if err != nil {
		panic(err.Error())
	}
	return out
}
