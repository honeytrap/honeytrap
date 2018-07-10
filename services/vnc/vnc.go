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

func (s *vncService) SetDataDir(string) {}

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
