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

/*

[service.redis]
type="redis"
version="2.8.4"

[[port]]
port="tcp/6379"
services=["redis"]

*/

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

		payload := []byte{}
		cmd := ""
		for i := 0; i < len(items); i++ {
			Item := items[i].(redisDatum)
			command, success := Item.ToString()
			if !success {
				log.Error("Expected a command string, got something else (type=%q)", Item.DataType)
				break
			}
			if i == 0 {
				cmd = command
			}
			payload = append(payload, command+" "...)
		}

		answer := s.REDISHandler(cmd, items[1:])

		s.ch.Send(event.New(
			services.EventOptions,
			event.Category("redis"),
			event.Type(cmd),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("redis.command", cmd),
			event.Payload(payload),
		))
		_, err = conn.Write([]byte(answer))
		if err != nil {
			log.Error("Error writing response: %s", err.Error())
			return err
		}
	}

	return nil

}
