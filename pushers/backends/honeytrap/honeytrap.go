package honeytrap

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/honeytrap/honeytrap/pushers/api"

	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:honeytrap")

// TrapConfig defines the configuration used to setup a TrapChannel.
type TrapConfig struct {
	Host  string `toml:"host"`
	Token string `toml:"token"`
}

// TrapChannel defines a struct which implmenets the pushers.Channel
// interface for delivery honeytrap special messages.
type TrapChannel struct {
	client *api.Client
}

// New returns a new instance of a TrapChannel.
func New(config TrapConfig) TrapChannel {
	return TrapChannel{
		client: api.New(&api.Config{
			Url:   config.Host,
			Token: config.Token,
		}),
	}
}

// Send delivers all messages to the underline connection.
func (hc TrapChannel) Send(messages []message.PushMessage) {
	// TODO:
	// req, err := hc.client.NewRequest("POST", "v1/action/{sensor}/{type}", actions)

	for _, message := range messages {
		var err error
		var req *http.Request

		if message.Sensor == "honeytrap" && message.Category == "ping" {
			req, err = hc.client.NewRequest("POST", "v1/ping", nil)
		} else {
			// TODO: workaround, need to update api
			req, err = hc.client.NewRequest("POST", "v1/action",
				[]interface{}{
					message.Data,
				},
			)
		}

		if err != nil {
			log.Errorf("HoneytrapChannel: Error while preparing request: %s", err.Error())
			continue
		}
		var resp *http.Response
		if resp, err = hc.client.Do(req, nil); err != nil {
			log.Errorf("HoneytrapChannel: Error while sending message: %s", err.Error())
			continue
		}

		// for keep-alive
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Errorf("HoneytrapChannel: Unexpected status code: %d", resp.StatusCode)
			continue
		}
	}

	log.Infof("HoneytrapChannel: Sent %d actions.", len(messages))
}
