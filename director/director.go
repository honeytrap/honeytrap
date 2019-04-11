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

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) func(Director) error {
	return func(d Director) error {
		err := decoder.PrimitiveDecode(c, d)
		return err
	}
}
