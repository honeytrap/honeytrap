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
package server

import (
	"net"
	"sync"
)

func PeekConnection(conn net.Conn) *peekConnection {
	return &peekConnection{
		conn,
		[]byte{},
		sync.Mutex{},
	}
}

type peekConnection struct {
	net.Conn

	buffer []byte
	m      sync.Mutex
}

func (pc *peekConnection) Peek(p []byte) (int, error) {
	pc.m.Lock()
	defer pc.m.Unlock()

	n, err := pc.Conn.Read(p)

	pc.buffer = append(pc.buffer, p[:n]...)
	return n, err
}

func (pc *peekConnection) Read(p []byte) (n int, err error) {
	pc.m.Lock()
	defer pc.m.Unlock()

	// first serve from peek buffer
	if len(pc.buffer) > 0 {
		bn := copy(p, pc.buffer)
		pc.buffer = pc.buffer[bn:]
		return bn, nil
	}

	return pc.Conn.Read(p)
}
