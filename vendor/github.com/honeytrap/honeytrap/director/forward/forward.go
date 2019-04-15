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
package forward

import (
	"errors"
	"fmt"
	"net"

	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = director.Register("forward", New)
)

func New(options ...func(director.Director) error) (director.Director, error) {
	d := &forwardDirector{
		eb: pushers.MustDummy(),
	}

	for _, optionFn := range options {
		optionFn(d)
	}

	return d, nil
}

type forwardDirector struct {
	eb pushers.Channel

	Host string `toml:"host"`
}

func (d *forwardDirector) SetChannel(eb pushers.Channel) {
	d.eb = eb
}

func (d *forwardDirector) Dial(conn net.Conn) (net.Conn, error) {
	host := d.Host
	protocol := ""
	port := ""

	if ta, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		port = fmt.Sprintf("%d", ta.Port)
		protocol = "tcp"
	} else if ta, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		port = fmt.Sprintf("%d", ta.Port)
		protocol = "udp"
	} else {
		return nil, errors.New("Unsupported protocol")
	}

	// port is being overruled
	if h, v, err := net.SplitHostPort(host); err == nil {
		host = h
		port = v
	}

	return net.Dial(protocol, net.JoinHostPort(host, port))
}
