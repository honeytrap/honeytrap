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
	logging "github.com/op/go-logging"
	"os/user"
	"plugin"
	"path"
)

var log = logging.MustGetLogger("honeytrap:channels")

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
	for k, _ := range channels {
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
	if fn, ok := channels[key]; ok {
		return fn, true
	}
	/*
	luaPl, ok := readfile(name)
	if ok {
		return lua.New(luaPl), nil
	}
*/

	// messy, todo: fix/choose path
	// https://stackoverflow.com/a/17617721
	usr, _ := user.Current()
	home := usr.HomeDir
	dynamicPl, err := plugin.Open(path.Join(home, ".honeytrap", key+".so"))
	if err != nil {
		log.Errorf("Couldn't load dynamic plugin: %s", err.Error())
		return nil, false
	}
	sym, err := dynamicPl.Lookup("Channel")
	if err != nil {
		log.Errorf("Couldn't lookup Channel symbol: %s", err.Error())
		return nil, false
	}

	return sym.(ChannelFunc), true
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
