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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
)

var (
	_ = pushers.Register("file", New)
)

var (
	defaultMaxSize  = int64(1024 * 1024 * 1024)
	defaultWaitTime = 5 * time.Second

	log = logging.MustGetLogger("channels/file")
)

// New returns a new instance of a FileBackend.
func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	fc := FileBackend{
		FileConfig: FileConfig{
			MaxSize: defaultMaxSize,
			Mode:    os.FileMode(0600),
		},
		request: make(chan map[string]interface{}),
	}

	for _, optionFn := range options {
		optionFn(&fc)
	}

	if fc.File == "" {
		return nil, errors.New("File channel: filename not set")
	}

	if fc.MaxSize < 1024 {
		return nil, errors.New("File channel: minimal max size is 1024")
	}

	if path.IsAbs(fc.File) {
	} else if pwd, err := os.Getwd(); err == nil {
		fc.File = filepath.Join(pwd, fc.File)
	}

	go fc.writeLoop()

	return &fc, nil
}

// FileConfig defines the config used to setup the FileBackend.
type FileConfig struct {
	MaxSize int64       `toml:"maxsize"`
	File    string      `toml:"filename"`
	Mode    os.FileMode `toml:"mode"`
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

	request chan map[string]interface{}
}

func (f *FileBackend) Close() {
	close(f.request)
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (f *FileBackend) Send(message event.Event) {
	mp := make(map[string]interface{})

	message.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	f.request <- mp
}

// syncLoop handles configuration of the giving loop for writing to file.
func (f *FileBackend) writeLoop() {
	dest, err := OpenRotateFile(f.File, f.Mode, f.MaxSize)
	if err != nil {
		log.Errorf("Failed create destination file: %s", err)
		return
	}

	defer dest.Close()

	var buf bytes.Buffer

	for {
		select {
		case req, ok := <-f.request:
			if !ok {
				return
			}

			if err := json.NewEncoder(&buf).Encode(req); err != nil {
				log.Errorf("Failed to marshal PushMessage to JSON : %+q", err)
				continue
			}

			if buf.Len() < (500 * 1024) {
				continue
			}
		case <-time.After(time.Second):
		}

		if _, err := io.Copy(dest, &buf); err != nil {
			log.Errorf("Failed to copy data to File : %+q", err)
		}

		if err := dest.Sync(); err != nil {
			log.Errorf("Failed to sync Write to File : %+q", err)
		}

		// Reset the buffer for reuse.
		buf.Reset()
	}
}
