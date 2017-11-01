// +build linux

/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package netstack

import (
	"errors"
	"io"
	"net"

	"time"

	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/waiter"
)

type Conn struct {
	ep tcpip.Endpoint

	wq        *waiter.Queue
	waitEntry *waiter.Entry

	notifyCh chan struct{}
}

func (dc *Conn) Read(b []byte) (int, error) {
	for {
		v, err := dc.ep.Read(nil)
		if err == nil {
		} else if err == tcpip.ErrWouldBlock {
			<-dc.notifyCh
			continue
		} else {
			return 0, io.EOF
		}

		n := copy(b, v)
		v.TrimFront(n)
		return n, nil
	}
}

func (dc *Conn) Write(b []byte) (int, error) {
	v := buffer.NewViewFromBytes(b)
	if n, err := dc.ep.Write(v, nil); err != nil {
		return 0, errors.New("XXX")
	} else {
		return int(n), nil
	}
}

func (dc *Conn) Close() error {
	dc.wq.EventUnregister(dc.waitEntry)
	dc.ep.Close()
	return nil
}

func (dc *Conn) LocalAddr() net.Addr {
	la, _ := dc.ep.GetLocalAddress()
	ip := net.ParseIP(string(la.Addr))
	return &net.TCPAddr{
		IP:   ip,
		Port: int(la.Port),
	}
}

func (dc *Conn) RemoteAddr() net.Addr {
	ra, _ := dc.ep.GetRemoteAddress()
	ip := net.ParseIP(string(ra.Addr))
	return &net.TCPAddr{
		IP:   ip,
		Port: int(ra.Port),
	}
}

func (dc *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (dc *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (dc *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func newConn(wq *waiter.Queue, ep tcpip.Endpoint) net.Conn {
	// Create wait queue entry that notifies a channel.
	waitEntry, notifyCh := waiter.NewChannelEntry(nil)

	wq.EventRegister(&waitEntry, waiter.EventIn)

	return &Conn{
		ep:        ep,
		wq:        wq,
		notifyCh:  notifyCh,
		waitEntry: &waitEntry,
	}
}
