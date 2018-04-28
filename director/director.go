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
package director

import (
	"net"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	directors = map[string]func(...func(Director) error) (Director, error){}
)

func Register(key string, fn func(...func(Director) error) (Director, error)) func(...func(Director) error) (Director, error) {
	directors[key] = fn
	return fn
}

func Get(key string) (func(...func(Director) error) (Director, error), bool) {
	if fn, ok := directors[key]; ok {
		return fn, true
	}

	return nil, false
}

func GetAvailableDirectorNames() []string {
	var out []string
	for key := range directors {
		out = append(out, key)
	}
	return out
}

// Director defines an interface which exposes an interface to allow structures that
// implement this interface allow us to control containers which they provide.
type Director interface {
	Dial(net.Conn) (net.Conn, error)
	//Dial(network) (net.Conn, error)
	//DialUDP() (net.Conn, error)
	// Run(ctx context.Context)
}

type SetChanneler interface {
	SetChannel(pushers.Channel)
}

func WithChannel(channel pushers.Channel) func(Director) error {
	return func(d Director) error {
		if sc, ok := d.(SetChanneler); ok {
			sc.SetChannel(channel)
		}
		return nil
	}
}

func WithConfig(c toml.Primitive) func(Director) error {
	return func(d Director) error {
		err := toml.PrimitiveDecode(c, d)
		return err
	}
}
