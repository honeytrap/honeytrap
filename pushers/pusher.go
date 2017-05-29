package pushers

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
)

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send(...message.Event)
}

//=======================================================================================================

// ChannelGenerator defines a function type which returns a Channel created
// from a primitive.
type ChannelGenerator func(toml.MetaData, toml.Primitive) (Channel, error)

// TODO(alex): Decide if we need a mutex to secure things concurrently.
// We assume it will never be read/written to concurrently.
var backends = struct {
	b map[string]ChannelGenerator
}{
	b: make(map[string]ChannelGenerator),
}

// RegisterBackend adds the giving generator to the global generator lists.
func RegisterBackend(name string, generator ChannelGenerator) {
	backends.b[name] = generator
}

// NewBackend returns a new Channel of the giving name with the provided toml.Primitive.
func NewBackend(name string, meta toml.MetaData, primi toml.Primitive) (Channel, error) {
	log.Info("honeytrap.Pusher : Creating %q Backend : %#q", name, primi)

	maker, ok := backends.b[name]
	if !ok {
		return nil, fmt.Errorf("Backend %q maker not found", name)
	}

	return maker(meta, primi)
}

//=======================================================================================================

// Pusher defines a struct which implements a pusher to manage the loading and
// delivery of message.Event.
type Pusher struct {
	config   *config.Config
	q        chan message.Event
	queue    []message.Event
	age      time.Duration
	channels []Channel
}

// New returns a new instance of Pusher.
func New(conf *config.Config) *Pusher {
	p := &Pusher{
		config: conf,
		queue:  []message.Event{},
		q:      make(chan message.Event),
		age:    conf.Delays.PushDelay.Duration(),
	}

	for _, cb := range conf.Channels {
		channels, err := MakeFilter(conf, cb)
		if err != nil {
			log.Info("honeytrap.Pusher : Failed creating Filter channels : %#q", cb)
			continue
		}

		p.channels = append(p.channels, channels...)
	}

	return p
}

// Start defines a function which kickstarts the internal pusher call loop.
func (p *Pusher) Start() {
	go p.run()
}

func (p *Pusher) run() {
	log.Info("Pusher started")
	defer log.Info("Pusher exited")

	for {
		select {
		case <-time.After(p.age):
			p.flush()
		case a := <-p.q:
			p.add(a)
		}
	}

	// TODO: We need to figure out where the call to Run stops,
	// 1. Does it stop after the call to time.After?
	// 2. Does it not stop at all, hence this code becomes unreachable.
}

// Send delivers the giving events to the underline channels.
func (p *Pusher) Send(messages ...message.Event) {
	for index, item := range messages {
		item.Date = time.Now()
		item.Token = p.config.Token

		messages[index] = item
	}

	for _, channel := range p.channels {
		channel.Send(messages...)
	}
}

func (p *Pusher) flush() {
	if len(p.queue) == 0 {
		return
	}

	go p.Send(p.queue...)

	p.queue = []message.Event{}
}

// TODO: Cannot we do the following
// switch (a.(type)) {
// case Action:
//  url = "/action"
// case Record:
// url = "/blabla"
//}

// Push adds the giving data as part of a single PushMessage to be published to
// the pushers backends.
func (p *Pusher) Push(sensor, category, containerID, sessionID string, data interface{}) {
	p.q <- message.Event{
		Sensor:      sensor,
		Category:    category,
		SessionID:   sessionID,
		ContainerID: containerID,
		Data:        data,
	}
}

// PushFile adds the giving data as push notifications for a file data.
// TODO: implement PushFile instead of RecordPush
func (p *Pusher) PushFile(sensor, category, containerID, sessionID string, filename string, data []byte) {
	p.q <- message.Event{
		Sensor:      sensor,
		Category:    category,
		SessionID:   sessionID,
		ContainerID: containerID,
		Data:        data,
	}
}

func (p *Pusher) add(a message.Event) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
}

//=======================================================================================================
