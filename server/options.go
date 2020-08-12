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

package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"

	_ "net/http/pprof"

	"github.com/honeytrap/honeytrap/storage"
	"github.com/pkg/profile"
	"github.com/rs/xid"

	"github.com/honeytrap/honeytrap/server/profiler"
)

type OptionFn func(*Honeytrap) error

func WithMemoryProfiler() OptionFn {
	return func(b *Honeytrap) error {
		b.profiler = profiler.New(profile.MemProfile)
		return nil
	}
}

func WithCPUProfiler() OptionFn {
	return func(b *Honeytrap) error {
		b.profiler = profiler.New(profile.CPUProfile)
		return nil
	}
}

func WithConfig(s string) (OptionFn, error) {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}

	return func(b *Honeytrap) error {
		return b.config.Load(bytes.NewBuffer(data))
	}, nil
}

func WithRemoteConfig(s string) (OptionFn, error) {
	resp, err := http.Get(s)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return func(b *Honeytrap) error {
		return b.config.Load(bytes.NewBuffer(body))
	}, nil
}

func WithDataDir(s string) (OptionFn, error) {
	var err error

	p, err := expand(s)
	if err != nil {
		return nil, err
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = os.Mkdir(p, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return func(b *Honeytrap) error {
		b.dataDir = p
		storage.SetDataDir(p)
		return nil
	}, nil
}

func expand(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func WithToken() OptionFn {
	uid := xid.New().String()

	return func(h *Honeytrap) error {
		h.token = uid

		p := h.dataDir
		p = path.Join(p, "token")

		if _, err := os.Stat(p); os.IsNotExist(err) {
			ioutil.WriteFile(p, []byte(uid), 0600)
		} else if err != nil /* other error */ {
			return err
		} else if data, err := ioutil.ReadFile(p); err != nil {
			return err
		} else {
			uid = string(data)
		}

		h.token = uid
		return nil
	}
}
