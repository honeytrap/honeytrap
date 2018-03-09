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
package fschannel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
)

var (
	_ = pushers.Register("file", New)
)

var (
	defaultMaxSize  = 1024 * 1024 * 1024
	defaultWaitTime = 5 * time.Second
	crtlline        = []byte("\r\n")
	log             = logging.MustGetLogger("channels/file")
)

// New returns a new instance of a FileBackend.
func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	fc := FileBackend{
		FileConfig: FileConfig{
			MaxSize: defaultMaxSize,
		},
		request: make(chan map[string]interface{}),
		closer:  make(chan struct{}),
	}

	for _, optionFn := range options {
		optionFn(&fc)
	}

	if fc.File == "" {
		return nil, errors.New("File channel: filename not set")
	}

	if path.IsAbs(fc.File) {
	} else if pwd, err := os.Getwd(); err == nil {
		fc.File = filepath.Join(pwd, fc.File)
	}

	fc.timeout = config.MakeDuration(fc.Timeout, uint64(defaultWaitTime))

	return &fc, nil
}

// FileConfig defines the config used to setup the FileBackend.
type FileConfig struct {
	MaxSize int    `toml:"maxsize"`
	File    string `toml:"filename"`
	Timeout string `toml:"timeout"`
}

// FileBackend defines a struct which implements the pushers.Pusher interface
// and allows us to write PushMessage updates into a giving file path. Mainly for
// the need to sync PushMessage to local files for persistence.
// File paths provided are either created with a append mode if they already
// exists else will be created. FileBackend will also restrict filesize to a max of 1gb by default else if
// there exists a max size set in configuration, then that will be used instead,
// also the old file will be renamed with the current timestamp and a new file created.
type FileBackend struct {
	FileConfig
	timeout time.Duration
	dest    *os.File
	request chan map[string]interface{}
	closer  chan struct{}
	wg      sync.WaitGroup
}

// Wait calls the internal waiter.
func (f *FileBackend) Wait() {
	f.wg.Wait()
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (f *FileBackend) Send(message event.Event) {
	if err := f.syncWrites(); err != nil {
		log.Errorf("Error syncing writes: %+q", err)
		return
	}

	mp := make(map[string]interface{})

	message.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	f.request <- mp
}

// syncWrites startups the channel procedure to listen for new writes to giving file.
func (f *FileBackend) syncWrites() error {
	if f.dest != nil && f.request != nil {
		return nil
	}

	// If the request channel has been niled but file is still opened,
	// close it.
	if f.dest != nil {
		f.dest.Sync()
		f.dest.Close()
	}

	if f.request == nil {
		f.request = make(chan map[string]interface{})
	}

	var err error

	f.dest, err = newFile(f.File, f.MaxSize)
	if err != nil {
		log.Errorf("Failed create destination file: %s", err)
		return err
	}

	f.wg.Add(1)
	go f.syncLoop()

	return nil
}

// syncLoop handles configuration of the giving loop for writing to file.
func (f *FileBackend) syncLoop() {
	defer f.wg.Done()

	ticker := time.NewTimer(f.timeout)
	var buf bytes.Buffer

	{
	writeSync:
		for {
			select {
			case <-ticker.C:
				f.dest.Close()
				f.dest = nil

				// Close request channel and nil it.
				f.request = nil

				break writeSync

			case req, ok := <-f.request:
				if !ok {
					f.dest.Close()
					f.dest = nil
					f.request = nil
					break writeSync
				}

				if err := json.NewEncoder(&buf).Encode(req); err != nil {
					log.Errorf("Failed to marshal PushMessage to JSON : %+q", err)
					continue writeSync
				}

				if buf.Len() < (500 * 1024) {
					continue
				}
			case <-time.After(time.Second):
			}

			if _, err := io.Copy(f.dest, &buf); err != nil && err != io.EOF {
				log.Errorf("Failed to copy data to File : %+q", err)
			}

			if err := f.dest.Sync(); err != nil {
				log.Errorf("Failed to sync Write to File : %+q", err)
			}

			// Reset the buffer for reuse.
			buf.Reset()
		}
	}
}

// newFile returns a new file with the giving target path and returns the
// new file object.
func newFile(targetPath string, maxSize int) (*os.File, error) {
	// Attempt to stat file, if it does not exists then create a new one.
	stat, err := os.Stat(targetPath)
	if err != nil {
		dest, err := os.OpenFile(targetPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}

		return dest, nil
	}

	if stat.IsDir() {
		return nil, errors.New("Only direct file paths allowed")
	}

	// if we dealing with a file still  below our max size, then
	// open file if already exists else
	if int(stat.Size()) <= maxSize {
		dest, err := os.OpenFile(targetPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}

		return dest, nil
	}

	if err := os.Rename(targetPath, fmt.Sprintf("%s-%s", targetPath, stat.ModTime().Format("20060102150405"))); err != nil {
		return nil, err
	}

	return os.OpenFile(targetPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
}
