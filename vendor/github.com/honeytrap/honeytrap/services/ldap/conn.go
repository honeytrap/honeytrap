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
package ldap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net"
)

type Conn struct {
	con net.Conn

	ConnReader *bufio.Reader

	isTLS bool
}

func NewConn(c net.Conn) *Conn {
	return &Conn{
		con:        c,
		ConnReader: bufio.NewReader(c),
		isTLS:      false,
	}
}

// StartTLS
func (c *Conn) StartTLS(config *tls.Config) error {

	if c.isTLS {
		return errors.New("TLS already established")
	}

	tc := tls.Server(c.con, config)

	if err := tc.Handshake(); err != nil {
		return err
	}

	c.con = tc
	c.ConnReader = bufio.NewReader(tc)
	c.isTLS = true

	return nil
}
