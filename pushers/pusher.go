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
	Send([]message.PushMessage)
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

// ProxyPusher defines a decorator for the Pusher object which decorates it as a
// Channel.
type ProxyPusher struct {
	pusher *Pusher
}

// NewProxyPusher returns a new instance of a ProxyPusher
func NewProxyPusher(p *Pusher) *ProxyPusher {
	return &ProxyPusher{pusher: p}
}

// Send delivers the messages to the underline Pusher instance.
func (p ProxyPusher) Send(messages []message.PushMessage) {
	p.pusher.send(messages)
}

//=======================================================================================================

// Pusher defines a struct which implements a pusher to manage the loading and
// delivery of message.PushMessage.
type Pusher struct {
	config   *config.Config
	q        chan message.PushMessage
	queue    []message.PushMessage
	age      time.Duration
	channels []Channel
}

// New returns a new instance of Pusher.
func New(conf *config.Config) *Pusher {
	p := &Pusher{
		config: conf,
		queue:  []message.PushMessage{},
		q:      make(chan message.PushMessage),
		age:    conf.Delays.PushDelay.Duration(),
	}

	for _, cb := range conf.Channels {
		master := NewMasterChannel(conf)
		if err := master.UnmarshalConfig(cb); err != nil {
			log.Errorf("Failed to create channel for config [%#q]: %+q", cb, err)
			continue
		}

		p.channels = append(p.channels, master)
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

func (p *Pusher) send(messages []message.PushMessage) {
	for _, channel := range p.channels {
		channel.Send(messages)
	}
}

func (p *Pusher) flush() {
	if len(p.queue) == 0 {
		return
	}

	go p.send(p.queue)

	p.queue = []message.PushMessage{}
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
	p.q <- message.PushMessage{
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
	p.q <- message.PushMessage{
		Sensor:      sensor,
		Category:    category,
		SessionID:   sessionID,
		ContainerID: containerID,
		Data:        data,
	}
}

func (p *Pusher) add(a message.PushMessage) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
}

//=======================================================================================================
