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
	"fmt"
	"net"
	"strconv"

	"bufio"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("services/redis")

var (
	_ = services.Register("redis", REDIS)
)

func REDIS(options ...services.ServicerFunc) services.Servicer {
	s := &redisService{
		redisServiceConfig: redisServiceConfig{
			Version: "4.0.6",
			Os:      "Linux 4.9.49-moby x86_64",
		},
	}
	for _, o := range options {
		o(s)
	}
	return s
}

type redisServiceConfig struct {
	Version string `toml:"version"`

	Os string `toml:"os"`
}

type redisService struct {
	redisServiceConfig

	ch pushers.Channel
}

func (s *redisService) SetChannel(c pushers.Channel) {
	s.ch = c
}

type redisDatum struct {
	DataType byte
	Content  interface{}
}

func (d *redisDatum) ToString() (value string, success bool) {
	switch d.DataType {
	case 0x2b:
		fallthrough
	case 0x24:
		return d.Content.(string), true
	default:
		return "", false
	}
}

func parseRedisData(scanner *bufio.Scanner) (redisDatum, error) {
	success := scanner.Scan()
	if !success {
		err := scanner.Err()
		if err == nil {
			err = fmt.Errorf("eof")
		}
		return redisDatum{}, err
	}
	cmd := scanner.Text()
	if len(cmd) == 0 {
		return redisDatum{}, nil
	}
	dataType := cmd[0]
	if dataType == 0x2a { // 0x2a = '*', introduces an array
		n, err := strconv.ParseUint(cmd[1:], 10, 64)
		if err != nil {
			return redisDatum{}, fmt.Errorf("Error parsing command array size: %s", err.Error())
		}
		var items []interface{}
		for i := uint64(0); i < n; i++ {
			item, err := parseRedisData(scanner)
			if err != nil {
				return redisDatum{}, err
			}
			items = append(items, item)
		}
		return redisDatum{DataType: dataType, Content: items}, nil
	} else if dataType == 0x2b { // 0x2a = '+', introduces a simple string
		return redisDatum{DataType: dataType, Content: cmd[1:]}, nil
	} else if dataType == 0x24 { // 0x24 = '$', introduces a bulk string
		// Read (and ignore) string length
		_, err := strconv.ParseUint(cmd[1:], 10, 64)
		if err != nil {
			return redisDatum{}, err
		}
		scanner.Scan()
		str := scanner.Text()
		return redisDatum{DataType: dataType, Content: str}, nil
	} else if dataType == 0x3a { // 0x3a = ':', introduces an integer
		n, err := strconv.ParseUint(cmd[1:], 10, 64)
		return redisDatum{DataType: dataType, Content: n}, err
	} else {
		return redisDatum{}, fmt.Errorf("Unexpected data type: %q", dataType)
	}
}

func (s *redisService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for {
		datum, err := parseRedisData(scanner)
		if err != nil {
			if err.Error() != "eof" {
				log.Error(err.Error())
			}
			break
		}

		// Dirty hack to ignore "empty" packets (\r\n with no Redis content)
		if datum.DataType == 0x00 {
			continue
		}
		// Redis commands are sent as an array of strings, so expect that
		if datum.DataType != 0x2a {
			log.Error("Expected array, got data type %q", datum.DataType)
			break
		}
		items := datum.Content.([]interface{})
		firstItem := items[0].(redisDatum)
		command, success := firstItem.ToString()
		if !success {
			log.Error("Expected a command string, got something else (type=%q)", firstItem.DataType)
			break
		}
		answer, closeConn := s.REDISHandler(command, items[1:])

		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("redis"),
			event.Type("redis-command"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("redis.command", command),
		))

		if closeConn {
			break
		}
		_, err = conn.Write([]byte(answer))
		if err != nil {
			log.Error("error writing response: %s", err.Error())
			break
		}
	}

	return nil

}
