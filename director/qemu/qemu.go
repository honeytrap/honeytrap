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

// Package qemu will implement a director for Qemu. Using this director
// firmware images could be run and used as target.
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
