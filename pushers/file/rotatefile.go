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
package fschannel

import (
	"fmt"
	"os"
	"time"
)

func OpenRotateFile(name string, mode os.FileMode, maxSize int64) (*rotateFile, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return nil, err
	}

	offset, err := f.Seek(0, 2)
	if err != nil {
		return nil, err
	}

	rf := &rotateFile{
		f:       f,
		path:    name,
		pos:     offset,
		mode:    mode,
		maxSize: maxSize,
	}

	if offset < maxSize {
		return rf, nil
	} else if err := rf.rotate(); err != nil {
		return rf, err
	}

	return rf, nil
}

type rotateFile struct {
	f *os.File

	mode    os.FileMode
	path    string
	pos     int64
	maxSize int64
}

func (f *rotateFile) rotate() error {
	f.f.Sync()
	f.f.Close()

	now := time.Now()

	if err := os.Rename(f.path, fmt.Sprintf("%s.%s", f.path, now.Format("20060102150405"))); err != nil {
		return err
	}

	return f.reopen()
}

func (f *rotateFile) reopen() error {
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY, f.mode)
	if err != nil {
		return err
	}

	f.f = file
	f.pos = 0
	return nil
}

func (f *rotateFile) checkStat() (os.FileInfo, error) {
	return os.Stat(f.path)
}

func (f *rotateFile) Write(p []byte) (int, error) {
	if _, err := f.checkStat(); err == nil {
	} else if err := f.reopen(); err != nil {
		return 0, err
	}

	written := 0

	for f.pos+int64(len(p)) > f.maxSize {
		j := f.maxSize - int64(f.pos)

		for ; j > 0; j-- {
			// line endings windows?
			if p[j] == '\n' {
				break
			}
		}

		n, err := f.f.Write(p[:j])
		if err != nil {
			return n, err
		}

		written += n

		// rotate
		if err := f.rotate(); err != nil {
			return written, err
		}

		// skip \n
		written += 1

		p = p[j+1:]
	}

	n, err := f.f.Write(p)

	f.pos += int64(n)
	return written + n, err
}

func (f *rotateFile) Close() error {
	return f.f.Close()
}

func (f *rotateFile) Sync() error {
	return f.f.Sync()
}
