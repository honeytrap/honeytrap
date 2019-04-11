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
