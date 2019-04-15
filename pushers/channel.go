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

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) func(Channel) error {
	return func(d Channel) error {
		err := decoder.PrimitiveDecode(c, d)
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
