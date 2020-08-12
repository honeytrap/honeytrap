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

//Package listener defines the honeytrap listener types.
package listener

import (
	"context"
	"net"
)

var (
	listeners = map[string]func(...func(Listener) error) (Listener, error){}
)

func Register(key string, fn func(...func(Listener) error) (Listener, error)) func(...func(Listener) error) (Listener, error) {
	listeners[key] = fn
	return fn
}

func Get(key string) (func(...func(Listener) error) (Listener, error), bool) {
	d := Dummy

	if fn, ok := listeners[key]; ok {
		return fn, true
	}

	return d, false
}

func Range(fn func(string)) {
	for k := range listeners {
		fn(k)
	}
}

type Listener interface {
	Start(ctx context.Context) error
	Close() error
	Accept() (net.Conn, error)
}
