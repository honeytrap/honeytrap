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

func WithHomeDir(s string) (OptionFn, error) {
	var err error
	var usr *user.User
	if usr, err = user.Current(); err != nil {
		return nil, err
	}

	var p string
	if s == ".honeytrap" {
		p = path.Join(usr.HomeDir, s)
	} else {
		p, err = filepath.Abs(s)
	}

	_, err = os.Stat(p)

	switch {
	case err == nil:
		break
	case os.IsNotExist(err):
		if err = os.Mkdir(p, 0755); err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	return func(b *Honeytrap) error {
		b.homeDir = p
		storage.SetHomeDir(p)
		return nil
	}, nil
}

func WithToken() OptionFn {
	uid := xid.New().String()

	return func(h *Honeytrap) error {
		h.token = uid
		p := h.homeDir
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
