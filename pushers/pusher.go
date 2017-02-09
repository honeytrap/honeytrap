package pushers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/honeytrap/honeytrap/config"
)

type ChannelFunc func(conf map[string]interface{}) (Channel, error)

var (
	availableChannels = map[string]ChannelFunc{}
)

type Channel interface {
	Send([]*PushMessage)
}

func RegisterChannel(name string, chanFn ChannelFunc) ChannelFunc {
	availableChannels[name] = chanFn
	return chanFn
}

type Pusher struct {
	config   *config.Config
	q        chan *PushMessage
	queue    []*PushMessage
	age      time.Duration
	channels []Channel
}

func New(conf *config.Config) *Pusher {
	channels := []Channel{}

	for _, c := range conf.Channels {
		var ok bool
		var name string
		if name, ok = c["name"].(string); !ok {
			// TODO: add available channel names
			log.Errorf("Channel name not provided. Available channels are: %s")
			continue
		}

		var ac ChannelFunc
		if ac, ok = availableChannels[name]; !ok {
			log.Errorf("Unknown channel: %s", name)
		}

		channel, err := ac(c)
		if err != nil {
			log.Error("Error initializing channel: %s", err.Error())
			continue
		}

		channels = append(channels, channel)
	}

	return &Pusher{
		q:        make(chan *PushMessage),
		config:   conf,
		queue:    []*PushMessage{},
		channels: channels,
		age:      conf.Delays.PushDelay.Duration(),
	}
}

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
	log.Info("Pusher exited")
}

func (p *Pusher) send(messages []*PushMessage) {
	for _, channel := range p.channels {
		channel.Send(messages)
	}
}

func (p *Pusher) flush() {
	if len(p.queue) == 0 {
		return
	}

	go p.send(p.queue)

	p.queue = []*PushMessage{}
}

// TODO: Cannot we do the following
// switch (a.(type)) {
// case Action:
//  url = "/action"
// case Record:
// url = "/blabla"
//}

type PushMessage struct {
	Sensor      string
	Category    string
	SessionID   string
	ContainerID string
	Data        interface{}
}

func (p *Pusher) Push(sensor, category, containerID, sessionID string, data interface{}) {
	p.q <- &PushMessage{
		sensor,
		category,
		sessionID,
		containerID,
		data,
	}
}

// TODO: implement PushFile instead of RecordPush
func (p *Pusher) PushFile(sensor, category, containerID, sessionID string, filename string, data []byte) {
	p.q <- &PushMessage{
		sensor,
		category,
		sessionID,
		containerID,
		data,
	}
}

func (p *Pusher) add(a *PushMessage) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
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

type RecordPush struct {
	config *config.Config
	queue  []*Record
	q      chan *Record
	age    time.Duration
}

func (p *RecordPush) Push(to string, data []byte) {
	p.q <- &Record{to, data}
}

func NewRecordPusher(conf *config.Config) *RecordPush {
	return &RecordPush{
		config: conf,
		queue:  []*Record{},
		q:      make(chan *Record),
		age:    conf.Delays.PushDelay.Duration(),
	}

}

func (p *RecordPush) add(a *Record) {
	p.queue = append(p.queue, a)

	if len(p.queue) > 20 {
		p.flush()
	}
}

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

	log.Info("RecordPusher stopped")

	return nil
}
