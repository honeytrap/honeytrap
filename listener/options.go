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
package listener

import (
	"net"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
)

type AddAddresser interface {
	AddAddress(net.Addr)
}

func WithAddress(protocol, address string) func(Listener) error {
	return func(l Listener) error {
		if a, ok := l.(AddAddresser); ok {
			if protocol == "tcp" {
				addr, _ := net.ResolveTCPAddr(protocol, address)
				a.AddAddress(addr)
			} else if protocol == "udp" {
				addr, _ := net.ResolveUDPAddr(protocol, address)
				a.AddAddress(addr)
			}
		}
		return nil
	}
}

type SetChanneler interface {
	SetChannel(pushers.Channel)
}

func WithChannel(channel pushers.Channel) func(Listener) error {
	return func(d Listener) error {
		if sc, ok := d.(SetChanneler); ok {
			sc.SetChannel(channel)
		}
		return nil
	}
}

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) func(Listener) error {
	return func(d Listener) error {
		err := decoder.PrimitiveDecode(c, d)
		return err
	}
}
