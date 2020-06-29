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
package ssh

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

func NewTypeWriterReadCloser(r io.ReadWriteCloser) *TypeWriterReadWriteCloser {
	return &TypeWriterReadWriteCloser{ReadWriteCloser: r, time: time.Now()}
}

type TypeWriterReadWriteCloser struct {
	io.ReadWriteCloser

	time   time.Time
	buffer bytes.Buffer
}

func sanitize(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "<br/>", -1)
	s = strings.Replace(s, "'", "\\'", -1)
	s = strings.Replace(s, "\b", "<backspace>", -1)
	return s
}

func (lr *TypeWriterReadWriteCloser) Write(p []byte) (n int, err error) {
	return lr.ReadWriteCloser.Write(p)
}

func (lr *TypeWriterReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = lr.ReadWriteCloser.Read(p)

	now := time.Now()
	lr.buffer.WriteString(fmt.Sprintf(".wait(%d)", int(now.Sub(lr.time).Seconds()*1000)))
	lr.buffer.WriteString(fmt.Sprintf(".put('%s')", sanitize(string(p[:n]))))
	lr.time = now

	log.Debugf("typed: %v", sanitize(string(p[:n])))
	return n, err
}

func (lr *TypeWriterReadWriteCloser) String() string {
	return lr.buffer.String()
}

func (lr *TypeWriterReadWriteCloser) Close() error {
	return lr.ReadWriteCloser.Close()
}
