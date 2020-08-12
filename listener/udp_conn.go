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

package listener

import (
	"net"
	"time"
)

type DummyUDPConn struct {
	Buffer []byte

	Laddr net.Addr
	Raddr *net.UDPAddr

	Fn func(b []byte, addr *net.UDPAddr) (int, error)
}

func (dc *DummyUDPConn) Read(b []byte) (int, error) {
	n := copy(b, dc.Buffer)
	dc.Buffer = dc.Buffer[n:]
	return n, nil
}

func (dc *DummyUDPConn) Write(b []byte) (int, error) {
	if dc.Fn == nil {
		return len(b), nil
	}

	return dc.Fn(b[:], dc.Raddr)
}

func (dc *DummyUDPConn) Close() error {
	return nil
}

func (dc *DummyUDPConn) LocalAddr() net.Addr {
	return dc.Laddr
}

func (dc *DummyUDPConn) RemoteAddr() net.Addr {
	return dc.Raddr
}

func (dc *DummyUDPConn) SetDeadline(t time.Time) error {
	return nil
}

func (dc *DummyUDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (dc *DummyUDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
