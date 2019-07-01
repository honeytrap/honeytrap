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
package fschannel_test

import (
	"fmt"
	"path"
	"testing"
	"time"

	"os"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	fschannel "github.com/honeytrap/honeytrap/pushers/file"
)

func WithPath(path string) func(pushers.Channel) error {
	return func(f pushers.Channel) error {
		fc := f.(*fschannel.FileBackend)
		fc.File = path
		return nil
	}
}

func WithMaxSize(maxSize int64) func(pushers.Channel) error {
	return func(f pushers.Channel) error {
		fc := f.(*fschannel.FileBackend)
		fc.MaxSize = maxSize
		return nil
	}
}

func TestStress(t *testing.T) {
	dir := os.TempDir()

	dir = path.Join(dir, fmt.Sprintf("honeytrap-%d", time.Now().Unix()))

	if err := os.Mkdir(dir, 0700); err != nil {
		t.Errorf("Could not make directory: %s: %s", dir, err)
	}

	t.Logf("Using temp directory %s", dir)

	channel, err := fschannel.New(WithPath(path.Join(dir, "test.log")), WithMaxSize(1024*1024))
	if err != nil {
		t.Fatalf("Error creating new file channel: %s", err)
	}

	for i := 0; i < 1000000; i++ {
		channel.Send(event.New())
	}

	channel.(*fschannel.FileBackend).Close()

	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Could not remove directory: %s: %s", dir, err)
	}
}
