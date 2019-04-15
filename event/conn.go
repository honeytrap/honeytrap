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
package event

import (
	"net"
)

type Conn struct {
	net.Conn

	options []Option
}

func (ec *Conn) Options() Option {
	return NewWith(ec.options...)
}

func WithConn(conn net.Conn, options ...Option) *Conn {
	if innerConn, ok := conn.(*Conn); ok {
		innerConn.options = append(innerConn.options, options...)
		return innerConn
	}

	return &Conn{
		Conn:    conn,
		options: options,
	}
}
