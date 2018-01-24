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
	"context"
	"net"

	"bufio"
	"io"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

var (
	_ = services.Register("redis", REDIS)
)

func REDIS(options ...services.ServicerFunc) services.Servicer {
	s := &redisService{
		redisServiceConfig: redisServiceConfig{
			Version: "4.0.6",
		},
	}
	for _, o := range options {
		o(s)
	}
	return s
}

type redisServiceConfig struct {
	Version string `toml:"version"`
}

type redisService struct {
	redisServiceConfig

	ch pushers.Channel
}

func (s *redisService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *redisService) Handle(ctx context.Context, conn net.Conn) error {

	defer conn.Close()

	br := bufio.NewReader(conn)

	for {
		cmd, err := br.ReadString('\n')

		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		answer, closeConn := REDISHandler(s, cmd)

		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("redis"),
			event.Type("redis-command"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("command", cmd),
		))

		if closeConn == true {
			break
		} else {
			_, err = conn.Write([]byte(answer))
		}
	}

	return nil

}
