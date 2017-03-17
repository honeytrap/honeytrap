package pushers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/elasticsearch"
	"github.com/honeytrap/honeytrap/pushers/fschannel"
	"github.com/honeytrap/honeytrap/pushers/honeytrap"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/honeytrap/honeytrap/pushers/slack"
)

// Waiter exposes a method to call a wait method to allow a channel finish it's
// operation.
type Waiter interface {
	Wait()
}

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send([]*message.PushMessage)
}

// Pusher defines a struct which implements a pusher to manage the loading and
// delivery of message.PushMessage.
type Pusher struct {
	config   *config.Config
	q        chan *message.PushMessage
	queue    []*message.PushMessage
	age      time.Duration
	channels []Channel
}

// New returns a new instance of Pusher.
func New(conf *config.Config) *Pusher {
	channels := []Channel{}

	for _, c := range conf.Channels {
		var ok bool
		var name string

		if name, ok = c["name"].(string); !ok {
			// TODO: add available channel names
			log.Errorf("Channel name not provided. Available channels are: %s", "")
			continue
		}

		switch name {
		case "elasticsearch":
			var elastic elasticsearch.ElasticSearchChannel
			if err := elastic.UnmarshalConfig(c); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			channels = append(channels, &elastic)
		case "file":
			fchan := fschannel.New()
			if err := fchan.UnmarshalConfig(c); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			channels = append(channels, fchan)
		case "honeytrap":
			var htrap honeytrap.HoneytrapChannel
			if err := htrap.UnmarshalConfig(c); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			channels = append(channels, &htrap)
		case "slack":
			var slackChannel slack.MessageChannel
			if err := slackChannel.UnmarshalConfig(c); err != nil {
				log.Errorf("Error initializing channel: %s", err.Error())
				continue
			}

			channels = append(channels, &slackChannel)
		}
	}

	return &Pusher{
		q:        make(chan *message.PushMessage),
		config:   conf,
		queue:    []*message.PushMessage{},
		channels: channels,
		age:      conf.Delays.PushDelay.Duration(),
	}
}

// Start defines a function which kickstarts the internal pusher call loop.
func (p *Pusher) Start() {
	go p.run()
}

func (p *Pusher) run() {
	log.Info("Pusher started")
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
	log.Info("Pusher exited")
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

// RecordPush implements a struct which hands record push functionality.
type RecordPush struct {
	config *config.Config
	queue  []*Record
	q      chan *Record
	age    time.Duration
}

// NewRecordPusher returns a new instance of a RecordPush.
func NewRecordPusher(conf *config.Config) *RecordPush {
	return &RecordPush{
		config: conf,
		queue:  []*Record{},
		q:      make(chan *Record),
		age:    conf.Delays.PushDelay.Duration(),
	}

}

// Push adds the giving data as Record notifications.
func (p *RecordPush) Push(to string, data []byte) {
	p.q <- &Record{to, data}
}

func (p *RecordPush) add(a *Record) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
}

// Run begins the run loop for the RecordPush structure.
func (p *RecordPush) Run() error {
	log.Info("RecordPusher started")
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
	log.Info("RecordPusher stopped")

	return nil
}

func (p *RecordPush) flush() {
	if len(p.queue) == 0 {
		return
	}

	go func(recs []*Record) {
		client := http.DefaultClient

		for _, rec := range recs {
			log.Info("Creating Http Req to %s", rec.Path)
			req, err := http.NewRequest("POST", rec.Path, bytes.NewBuffer(rec.Data))
			if err != nil {
				log.Error(err.Error())
				return
			}

			req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", p.config.Token))

			resp, err := client.Do(req)
			if err != nil {
				log.Error(err.Error())
				return
			}

			defer resp.Body.Close()

			b2, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err.Error())
				return
			}

			log.Info("Submission to (%s): %s with status %d", rec.Path, string(b2), resp.StatusCode)
		}
	}(p.queue)

	p.queue = []*Record{}
}
