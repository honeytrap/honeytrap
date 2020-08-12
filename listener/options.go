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

package listener

import (
	"net"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
)

type AddAddresser interface {
	AddAddress(net.Addr)
}

func WithAddress(protocol, address string) func(Listener) error {
	return func(l Listener) error {
		if a, ok := l.(AddAddresser); ok {
			if protocol == "tcp" {
				addr, _ := net.ResolveTCPAddr(protocol, address)
				a.AddAddress(addr)
			} else if protocol == "udp" {
				addr, _ := net.ResolveUDPAddr(protocol, address)
				a.AddAddress(addr)
			}
		}
		return nil
	}
}

type SetChanneler interface {
	SetChannel(pushers.Channel)
}

func WithChannel(channel pushers.Channel) func(Listener) error {
	return func(d Listener) error {
		if sc, ok := d.(SetChanneler); ok {
			sc.SetChannel(channel)
		}
		return nil
	}
}

type TLSConfigurer interface {
	AddTLSConfig(port int, certFile, keyFile string)
}

func WithTLSConfig(port int, certFile, keyFile string) func(Listener) error {
	return func(l Listener) error {
		if tc, ok := l.(TLSConfigurer); ok {
			tc.AddTLSConfig(port, certFile, keyFile)
		}
		return nil
	}
}

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) func(Listener) error {
	return func(d Listener) error {
		err := decoder.PrimitiveDecode(c, d)
		return err
	}
}
