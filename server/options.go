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
		} else if err != nil {
			// other error
			return err
		} else if data, err := ioutil.ReadFile(p); err == nil {
			uid = string(data)
		} else {
			return err
		}
		return nil
	}
}
