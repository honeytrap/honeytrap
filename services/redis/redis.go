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
package redis

import (
	"bufio"
	"context"

	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("services/redis")

var (
	_ = services.Register("redis", REDIS)
)

func REDIS(options ...services.ServicerFunc) services.Servicer {

	s := &redisService{
		RedisServiceConfig: RedisServiceConfiguration{},
	}

	for _, o := range options {
		o(s)
	}

	s.RedisServiceConfig, errList = s.configureRedisService()
	if len(errList) != 0 {
		for field, reason := range errList {
			log.Errorf("Could not add [%s]: %s", field, reason)
		}
	}

	return s
}

type redisService struct {
	RedisServiceConfig RedisServiceConfiguration `toml:"config"`
	ch                 pushers.Channel
}

func (s *redisService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *redisService) Handle(ctx context.Context, conn net.Conn) error {

	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {

		cmd := scanner.Text()

		answer, closeConn := s.REDISHandler(cmd)

		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("redis"),
			event.Type("redis-command"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("redis.command", cmd),
		))

		if closeConn {
			break
		} else {
			_, err := conn.Write([]byte(answer))
			if err != nil {
				log.Error("error writing response: %s", err.Error())
			}
		}
	}

	return nil
}
