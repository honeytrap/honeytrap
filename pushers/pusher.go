package pushers

import (
	"errors"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/elasticsearch"
	"github.com/honeytrap/honeytrap/pushers/fschannel"
	"github.com/honeytrap/honeytrap/pushers/honeytrap"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/honeytrap/honeytrap/pushers/slack"
)

//=======================================================================================================

// Waiter exposes a method to call a wait method to allow a channel finish it's
// operation.
type Waiter interface {
	Wait()
}

//=======================================================================================================

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send([]*message.PushMessage)
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
func (p ProxyPusher) Send(messages []*message.PushMessage) {
	p.pusher.send(messages)
}

//=======================================================================================================

// Pusher defines a struct which implements a pusher to manage the loading and
// delivery of message.PushMessage.
type Pusher struct {
	config   *config.Config
	q        chan *message.PushMessage
	queue    []*message.PushMessage
	age      time.Duration
	backends map[string]Channel
	channels []Channel
}

// New returns a new instance of Pusher.
func New(conf *config.Config) *Pusher {
	backends := make(map[string]Channel)

	for name, c := range conf.Backends {
		backend, ok := c.(map[string]interface{})
		if !ok {
			log.Errorf("Found key %q with non-map config value: %#v", name, c)
			continue
		}

		key, ok := backend["key"].(string)
		if !ok {
			key = name
		}

		// Check if key already exists and panic.
		if _, ok := backends[key]; ok {
			// TODO: should log instead of panic here?
			log.Errorf("Found key %q already used for previous backend", key)
			continue
		}

		switch name {
		case "elasticsearch":
			var elastic elasticsearch.ElasticSearchChannel
			if err := elastic.UnmarshalConfig(backend); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			backends[key] = &elastic
		case "file":
			fchan := fschannel.New()
			if err := fchan.UnmarshalConfig(backend); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())

				continue
			}

			backends[key] = fchan
		case "honeytrap":
			var htrap honeytrap.HoneytrapChannel
			if err := htrap.UnmarshalConfig(backend); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			backends[key] = &htrap
		case "slack":
			var slackChannel slack.MessageChannel
			if err := slackChannel.UnmarshalConfig(backend); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			backends[key] = &slackChannel
		}
	}

	p := &Pusher{
		config:   conf,
		backends: backends,
		queue:    []*message.PushMessage{},
		q:        make(chan *message.PushMessage),
		age:      conf.Delays.PushDelay.Duration(),
	}

	for _, cb := range conf.Channels {
		master := NewMasterChannel(p)
		if err := master.UnmarshalConfig(cb); err != nil {
			log.Errorf("Failed to create channel for config [%#q]: %+q", cb, err)
			continue
		}

		p.channels = append(p.channels, master)
	}

	return p
}

// ErrBackendNotFound defines an error returned when a backend key
// does not match the registered keys.
var ErrBackendNotFound = errors.New("Backend not found")

// GetBackend returns a giving backend registered with the pusher.
func (p *Pusher) GetBackend(key string) (Channel, error) {
	if channel, ok := p.backends[key]; ok {
		return channel, nil
	}

	return nil, ErrBackendNotFound
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

func (p *Pusher) send(messages []*message.PushMessage) {
	// TODO: Should we do some waitgroup here to ensure all channels
	// properly finish?
	for _, channel := range p.channels {
		channel.Send(messages)
	}
}

func (p *Pusher) flush() {
	if len(p.queue) == 0 {
		return
	}

	go p.send(p.queue)

	p.queue = []*message.PushMessage{}
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
	p.q <- &message.PushMessage{
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
	p.q <- &message.PushMessage{
		Sensor:      sensor,
		Category:    category,
		SessionID:   sessionID,
		ContainerID: containerID,
		Data:        data,
	}
}

func (p *Pusher) add(a *message.PushMessage) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
}

//=======================================================================================================
