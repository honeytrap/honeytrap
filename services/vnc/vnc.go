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
package vnc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"time"

	"image/png"

	logging "github.com/op/go-logging"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

var log = logging.MustGetLogger("services/vnc")

var (
	_ = services.Register("vnc", Vnc)
)

func Vnc(options ...services.ServicerFunc) services.Servicer {
	s := &vncService{}
	for _, o := range options {
		o(s)
	}

	if pwd, err := os.Getwd(); err != nil {
	} else if !filepath.IsAbs(s.ImagePath) {
		s.ImagePath = filepath.Join(pwd, s.ImagePath)
	}

	r, err := os.Open(s.ImagePath)
	if err != nil {
		log.Errorf("Could not open vnc image: %s", s.ImagePath)
		return nil
	}

	defer r.Close()

	im, err := png.Decode(r)
	if err != nil {
		log.Errorf("Could not decode png image: %s", s.ImagePath)
		return nil
	}

	s.li = &LockableImage{
		Img: im,
	}

	return s
}

type vncService struct {
	c pushers.Channel

	li *LockableImage

	ImagePath  string `toml:"image"`
	ServerName string `toml:"server-name"`
}

func (s *vncService) SetChannel(c pushers.Channel) {
	s.c = c
}

func (s *vncService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	bounds := s.li.Img.Bounds()

	c := newConn(bounds.Dx(), bounds.Dy(), conn)
	c.serverName = s.ServerName

	go c.serve()

	s.c.Send(event.New(
		event.Sensor("vnc"),
		event.Service("vnc"),
		event.Category("connect"),
		event.Type("connect"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
	))

	closec := make(chan bool)
	go func() {
		slide := 0
		tick := time.NewTicker(time.Second / 30)
		defer tick.Stop()

		haveNewFrame := false
		for {
			feed := c.Feed
			if !haveNewFrame {
				feed = nil
			}
			_ = feed
			select {
			case feed <- s.li:
				haveNewFrame = false
			case <-closec:
				return
			case <-tick.C:
				slide++

				haveNewFrame = true
			}
		}
	}()

	for e := range c.Event {
		_ = e
		/*
			s.c.Send(event.New(
				EventOptions,
				event.Category("vnc"),
				event.Type("vnc-event"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Payload(e)
			))
		*/
	}

	close(closec)

	return nil
}
