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
package agent

import (
	"io"
	"net"
	"sync"
	"time"
)

var (
	noDeadline = make(<-chan time.Time)
)

type agentConnection struct {
	Laddr net.Addr
	Raddr net.Addr

	buff   []byte
	closed bool

	readTimeout  time.Time
	writeTimeout time.Time

	in chan []byte

	out chan interface{}

	m sync.Mutex
}

func (dc *agentConnection) receive(data []byte) {
	dc.m.Lock()
	defer dc.m.Unlock()

	if dc.closed {
		return
	}

	dc.buff = append(dc.buff, data...)

	select {
	case dc.in <- []byte{}: // v.Payload {
	default:
	}
}

func (dc *agentConnection) Read(b []byte) (int, error) {
	dc.m.Lock()
	if len(dc.buff) != 0 {
		n := copy(b[:], dc.buff[0:])
		dc.buff = dc.buff[n:]
		dc.m.Unlock()
		return n, nil
	}
	dc.m.Unlock()

	after := noDeadline

	if !dc.readTimeout.IsZero() {
		after = time.After(time.Until(dc.readTimeout))
	}

	select {
	case <-after:
		return 0, ErrTimeout
	case _, ok := <-dc.in:
		if !ok {
			log.Errorf("Error reading from channel, return EOF")
			return 0, io.EOF
		}

		dc.m.Lock()
		n := copy(b[:], dc.buff[0:])
		dc.buff = dc.buff[n:]
		dc.m.Unlock()

		return n, nil
	}
}

func (dc *agentConnection) Write(b []byte) (int, error) {
	dc.m.Lock()
	defer dc.m.Unlock()

	payload := make([]byte, len(b))

	copy(payload, b)

	p := ReadWriteTCP{
		Laddr:   dc.LocalAddr(),
		Raddr:   dc.RemoteAddr(),
		Payload: payload[:],
	}

	after := noDeadline
	if !dc.writeTimeout.IsZero() {
		after = time.After(time.Until(dc.writeTimeout))
	}

	select {
	case <-after:
		return 0, ErrTimeout
	case dc.out <- p:
	}

	return len(b), nil
}

func (dc *agentConnection) Close() error {
	dc.m.Lock()
	defer dc.m.Unlock()

	if dc.closed {
		return nil
	}

	p := EOF{
		Laddr: dc.LocalAddr(),
		Raddr: dc.RemoteAddr(),
	}

	dc.out <- p

	dc.closed = true
	close(dc.in)

	return nil
}

func (dc *agentConnection) LocalAddr() net.Addr {
	return dc.Laddr
}

func (dc *agentConnection) RemoteAddr() net.Addr {
	return dc.Raddr
}

func (dc *agentConnection) SetDeadline(t time.Time) error {
	dc.SetReadDeadline(t)
	dc.SetWriteDeadline(t)
	return nil
}

func (dc *agentConnection) SetReadDeadline(t time.Time) error {
	dc.readTimeout = t
	return nil
}

func (dc *agentConnection) SetWriteDeadline(t time.Time) error {
	dc.writeTimeout = t
	return nil
}
