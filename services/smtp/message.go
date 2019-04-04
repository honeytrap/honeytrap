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
package smtp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/mail"
)

// Message smtp message
type Message struct {
	Header mail.Header

	Buffer *bytes.Buffer

	Body *bytes.Buffer
}

func (m *Message) Read(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	msg, err := mail.ReadMessage(bytes.NewReader(buff))
	if err != nil {
		m.Body = bytes.NewBuffer(buff)
		return err
	}

	m.Header = msg.Header

	buff, err = ioutil.ReadAll(msg.Body)
	if err != nil {
		return err
	}

	m.Body = bytes.NewBuffer(buff)
	return err
}
