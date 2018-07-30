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
	"context"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("services")

var (
	services = map[string]func(...ServicerFunc) Servicer{}
)

type ServicerFunc func(Servicer) error

func Register(key string, fn func(...ServicerFunc) Servicer) func(...ServicerFunc) Servicer {
	services[key] = fn
	return fn
}

func Range(fn func(string)) {
	for k := range services {
		fn(k)
	}
}

func Get(key string) (func(...ServicerFunc) Servicer, bool) {
	d := Dummy

	if fn, ok := services[key]; ok {
		return fn, true
	}

	return d, false
}

type CanHandlerer interface {
	CanHandle([]byte) bool
}

type Servicer interface {
	Handle(context.Context, net.Conn) error

	SetChannel(pushers.Channel)
}

func WithChannel(eb pushers.Channel) ServicerFunc {
	return func(d Servicer) error {
		d.SetChannel(eb)
		return nil
	}
}

type Proxier interface {
	SetDirector(director.Director)
}

func WithDirector(d director.Director) ServicerFunc {
	return func(s Servicer) error {
		if p, ok := s.(Proxier); ok {
			p.SetDirector(d)
		}
		return nil
	}
}

type TomlDecoder interface {
	PrimitiveDecode(primValue toml.Primitive, v interface{}) error
}

func WithConfig(c toml.Primitive, decoder TomlDecoder) ServicerFunc {
	return func(s Servicer) error {
		err := decoder.PrimitiveDecode(c, s)
		return err
	}
}

var (
	SensorLow = event.Sensor("services")

	EventOptions = event.NewWith(
		SensorLow,
	)
)

/*

var (
	_ = director.Register("low", New)
)

// New will configure the low interaction director
func New(options ...func(director.Director)) (director.Director, error) {
	d := &lowDirector{
		eb: pushers.Dummy(),
		l:  listener.MustDummy(),
	}

	for _, optionFn := range options {
		optionFn(d)
	}

	return d, nil
}

type lowDirector struct {
	l listener.Listener

	eb pushers.Channel
}

func (d *lowDirector) SetChannel(eb pushers.Channel) {
	d.eb = eb
}

func (d *lowDirector) SetListener(l listener.Listener) {
	d.l = l
}

// Run will start the low interaction director
func (d *lowDirector) Run(ctx context.Context) {
	log.Info("LowDirector started...")
	defer log.Info("LowDirector finished...")

	fns := map[string]func(net.Conn){
		// "23":   d.Telnet(),
		"8022": d.SSH(),
		"8080": d.HTTP(),
	}

	for {
		conn, err := d.l.Accept()
		if err != nil {
			panic(err)
		}

		log.Debugf("Connection received from: %s", conn.LocalAddr())

		go func() {
			defer conn.Close()

			_, port, _ := net.SplitHostPort(conn.LocalAddr().String())

			fn, ok := fns[port]
			if !ok {
				return
			}

			fn(conn)
		}()
	}
}

*/
