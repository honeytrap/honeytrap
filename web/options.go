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
package web

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/eventbus"
)

func WithEventBus(bus *eventbus.EventBus) func(*web) error {
	return func(w *web) error {
		w.SetEventBus(bus)
		return nil
	}
}

func WithDataDir(dataDir string) func(*web) error {
	return func(w *web) error {
		w.dataDir = dataDir

		if filepath.IsAbs(dataDir) {
		} else if pwd, err := os.Getwd(); err != nil {
		} else {
			w.dataDir = filepath.Join(pwd, dataDir)
		}

		return nil
	}
}

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) func(*web) error {
	return func(d *web) error {
		err := decoder.PrimitiveDecode(c, d)
		return err
	}
}
