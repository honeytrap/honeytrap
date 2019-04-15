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
	"context"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

// Dummy is a placeholder
func Dummy(options ...ServicerFunc) Servicer {
	s := &dummyService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type dummyService struct {
	c pushers.Channel
}

func (s *dummyService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *dummyService) Handle(ctx context.Context, conn net.Conn) error {
	b := bufio.NewReader(conn)
	for {
		line, err := b.ReadBytes('\n')
		if err != nil { // EOF, or worse
			break
		}

		s.c.Send(event.New(
			SensorLow,
			event.Category("echo"),
			event.Payload([]byte(line)),
		))

		conn.Write(line)
	}

	return nil
}
