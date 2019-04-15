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
package eventbus

import (
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

// EventBus defines a structure which provides a pubsub bus where message.Events
// are sent along it's wires for delivery
type EventBus struct {
	subscribers []pushers.Channel
}

// NewEventBus returns a new instance of a EventBus.
func New() *EventBus {
	return &EventBus{}
}

// Subscribe adds the giving channel to the list of subscribers for the giving bus.
func (eb *EventBus) Subscribe(channel pushers.Channel) error {
	eb.subscribers = append(eb.subscribers, channel)
	return nil
}

// Send deliverers the slice of messages to all subscribers.
func (eb *EventBus) Send(e event.Event) {
	for _, subscriber := range eb.subscribers {
		subscriber.Send(e)
	}
}
