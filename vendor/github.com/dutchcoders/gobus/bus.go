// MIT License

// Copyright (c) 2018 Dutchcoders

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package bus

import (
	"strings"
	"sync"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("bus")

var DefaultBus *Bus = New()

type BusFunc func(string, interface{}) error

type BusChan chan emitMessage

type Bus struct {
	subscriptions map[string]map[interface{}]BusChan

	m sync.Mutex
}

func New() *Bus {
	return &Bus{
		subscriptions: map[string]map[interface{}]BusChan{},
		m:             sync.Mutex{},
	}
}

func Subscribe(topic string, key interface{}, fn BusFunc) {
	DefaultBus.Subscribe(topic, key, fn)
}

func (b *Bus) Subscribe(topic string, key interface{}, fn BusFunc) {
	b.m.Lock()
	defer b.m.Unlock()

	if _, ok := b.subscriptions[topic]; !ok {
		b.subscriptions[topic] = map[interface{}]BusChan{}
	}

	ch := make(chan emitMessage)
	b.subscriptions[topic][key] = ch

	go func() {
		for msg := range ch {
			err := fn(msg.Subject, msg.Value)
			if err != nil {
				log.Error("Error emit: %s: %s", key, err.Error())
			}
		}
	}()
}

func SubscribeOnce(topic string, key interface{}, fn BusFunc) {
	DefaultBus.SubscribeOnce(topic, key, fn)
}

func (b *Bus) SubscribeOnce(topic string, key interface{}, fn BusFunc) {
	b.m.Lock()
	defer b.m.Unlock()

	if _, ok := b.subscriptions[topic]; !ok {
		b.subscriptions[topic] = map[interface{}]BusChan{}
	}

	ch := make(chan emitMessage)
	b.subscriptions[topic][key] = ch

	go func() {
		defer func() {
			b.m.Lock()
			defer b.m.Unlock()

			ch := b.subscriptions[topic][key]

			delete(b.subscriptions[topic], key)

			close(ch)
		}()

		msg := <-ch

		err := fn(msg.Subject, msg.Value)
		if err != nil {
			log.Error("Error emit: %s: %s", key, err.Error())
		}
	}()
}

func Unsubscribe(topic string, key interface{}) {
	DefaultBus.Unsubscribe(topic, key)
}

func (b *Bus) Unsubscribe(topic string, key interface{}) {
	b.m.Lock()
	defer b.m.Unlock()

	ch := b.subscriptions[topic][key]

	delete(b.subscriptions[topic], key)

	close(ch)
}

func Emit(topic string, o interface{}) {
	DefaultBus.Emit(topic, o)
}

type emitMessage struct {
	Subject string
	Value   interface{}
}

func (b *Bus) Emit(topic string, o interface{}) {
	for key, subscription := range b.subscriptions {
		parts := strings.Split(topic, ":")
		if parts[0] != key {
			continue
		}

		subject := ""
		if len(parts) > 1 {
			subject = parts[1]
		}

		for _, ch := range subscription {
			select {
			case ch <- emitMessage{
				Subject: subject,
				Value:   o,
			}:
			default:
				log.Error("Channel busy topic=%s", topic)
			}
		}
	}
}
