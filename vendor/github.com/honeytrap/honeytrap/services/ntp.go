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
	"context"
	"io"
	"net"
	"os"

	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("ntp", NTP)
)

// Ntp is a placeholder
func NTP(options ...ServicerFunc) Servicer {
	s := &ntpService{}
	for _, o := range options {
		o(s)
	}
	return s
}

type ntpService struct {
	c pushers.Channel
}

func (s *ntpService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *ntpService) Handle(ctx context.Context, conn net.Conn) error {
	// TODO: implement protocol support
	_, err := io.Copy(os.Stdout, conn)
	return err
}
