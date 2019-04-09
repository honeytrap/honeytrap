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
package qemu

import (
	"errors"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = director.Register("qemu", New)
)

func New(options ...func(director.Director) error) (director.Director, error) {
	d := &qemuDirector{
		eb: pushers.MustDummy(),
	}

	for _, optionFn := range options {
		optionFn(d)
	}

	return d, nil
}

type qemuDirector struct {
	eb pushers.Channel
}

func (d *qemuDirector) SetChannel(eb pushers.Channel) {
	d.eb = eb
}

func (d *qemuDirector) Dial(conn net.Conn) (net.Conn, error) {
	return nil, errors.New("Qemu director not implemented yet")
}
