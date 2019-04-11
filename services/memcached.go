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
package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"

	"strconv"

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

type memcachedService struct {
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
		hdr := make([]byte, 8)
		_, err := b.Read(hdr)
		if err != nil {
			log.Error("Error processing UDP header: %s", err.Error())
		}

		_ = hdr
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
		if conn.RemoteAddr().Network() != "udp" {
		} else if !s.limiter.Allow(conn.RemoteAddr()) {
			log.Warningf("Rate limit exceeded for host: %s", conn.RemoteAddr())
			return nil
		}

		parts := bytes.Split(command, []byte{0x20})

		switch string(parts[0]) {
		case "flush_all":
			conn.Write([]byte(`OK\r\n`))
		case "stats":
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
		case "add":
			fallthrough
		case "replace":
			fallthrough
		case "prepend":
			fallthrough
		case "append":
			fallthrough
		case "cas":
			fallthrough
		case "set":
			if len(parts) < 5 {
				return fmt.Errorf("Invalid number of arguments: %s", string(command))
			}

			key := string(parts[1])
			flags := string(parts[2])
			expireTime := string(parts[3])
			byteCount := string(parts[4])

			v, err := strconv.Atoi(byteCount)
			if err != nil {
				return fmt.Errorf("Byte count is not a number: %s", string(command))
			}
			count := v

			buff := make([]byte, 80)

			n, err := b.Read(buff)
			if err != nil {
				return err
			}

			buff = buff[:n]

			// discard rest of payload
			count -= n

			b.Discard(count)

			s.ch.Send(event.New(
				EventOptions,
				event.Category("memcached"),
				event.Protocol(conn.RemoteAddr().Network()),
				event.Type(fmt.Sprintf("memcached-%s", string(parts[0]))),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("memcached.command", string(parts[0])),
				event.Custom("memcached.key", key),
				event.Custom("memcached.flags", flags),
				event.Custom("memcached.expire-time", expireTime),
				event.Custom("memcached.bytes", byteCount),
				event.Payload(buff),
			))

			conn.Write([]byte("STORED\r\n"))
		default:
			conn.Write([]byte("ERROR\r\n"))
		}
	}

	return nil
}
