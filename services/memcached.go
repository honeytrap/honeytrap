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
package services

import (
	"bufio"
	"context"
	"encoding/hex"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("memcached", Memcached)
)

func Memcached(options ...ServicerFunc) Servicer {
	s := &memcachedService{
		limiter: NewLimiter(),
	}

	for _, o := range options {
		o(s)
	}

	return s
}

type memcachedServiceConfig struct {
}

type memcachedService struct {
	memcachedServiceConfig

	limiter *Limiter

	ch pushers.Channel
}

func (s *memcachedService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *memcachedService) Handle(ctx context.Context, conn net.Conn) error {
	b := bufio.NewReader(conn)

	// memcached behaves differently over UDP: it has an 8-bytes header
	if conn.RemoteAddr().Network() == "udp" {
		_, err := b.Discard(8)
		if err != nil {
			log.Error("Error processing UDP header: %s", err.Error())
		}
	}

	for {
		command, err := b.ReadBytes('\n')
		if err != nil {
			break
		}
		// Strip trailing \r\n
		sz := len(command)
		if sz >= 2 {
			command = command[:sz-2]
		}

		s.ch.Send(event.New(
			EventOptions,
			event.Category("memcached"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("memcached-command"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("memcached.command", string(command)),
			event.Custom("memcached.command-hex", hex.EncodeToString(command)),
		))

		// we return errors for udp connections, to prevent udp amplification
		if conn.RemoteAddr().Network() == "udp" {
			if s.limiter.Allow(conn.RemoteAddr()) {
				conn.Write([]byte("ERROR\r\n"))
			}

			return nil
		}

		if string(command) == "stats" {
			conn.Write([]byte(`
STAT pid 2080
STAT uptime 3151236
STAT time 1520550684
STAT version 1.4.13
STAT libevent 2.0.16-stable
STAT pointer_size 64
STAT rusage_user 371.247201
STAT rusage_system 1839.982991
STAT curr_connections 8
STAT total_connections 5547233
STAT connection_structures 55
STAT reserved_fds 20
STAT cmd_get 22076096
STAT cmd_set 21
STAT cmd_flush 3
STAT cmd_touch 0
STAT get_hits 22076066
STAT get_misses 30
STAT delete_misses 0
STAT delete_hits 0
STAT incr_misses 0
STAT incr_hits 0
STAT decr_misses 0
STAT decr_hits 0
STAT cas_misses 0
STAT cas_hits 0
STAT cas_badval 0
STAT touch_hits 0
STAT touch_misses 0
STAT auth_cmds 0
STAT auth_errors 0
STAT bytes_read 286857265
STAT bytes_written 129670828957
STAT limit_maxbytes 67108864
STAT accepting_conns 1
STAT listen_disabled_num 0
STAT threads 4
STAT conn_yields 0
STAT hash_power_level 16
STAT hash_bytes 524288
STAT hash_is_expanding 0
STAT expired_unfetched 0
STAT evicted_unfetched 0
STAT bytes 29828
STAT curr_items 5
STAT total_items 21
STAT evictions 0
STAT reclaimed 3
END\r\n
`))
		} else {
			conn.Write([]byte("ERROR\r\n"))
		}
	}

	return nil
}
