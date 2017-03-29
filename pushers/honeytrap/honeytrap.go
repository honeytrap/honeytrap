package honeytrap

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	api "github.com/honeytrap/honeytrap/pushers/api"

	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:honeytrap")

// TrapChannel defines a struct which implmenets the pushers.Channel
// interface for delivery honeytrap special messages.
type TrapChannel struct {
	client *api.Client
}

// UnmarshalConfig attempts to unmarshal the provided value into the giving
// HoneytrapChannel.
func (hc *TrapChannel) UnmarshalConfig(m interface{}) error {
	conf, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("Expected to receive a map")
	}

	if _, ok := conf["host"]; !ok {
		return fmt.Errorf("Host not set for channel honeytrap")
	}

	if _, ok := conf["token"]; !ok {
		return fmt.Errorf("Token not set for channel honeytrap")
	}

	hc.client = api.New(&api.Config{
		Url:   conf["host"].(string),
		Token: conf["token"].(string),
	})

	return nil
}

// Send delivers all messages to the underline connection.
func (hc TrapChannel) Send(messages []*message.PushMessage) {
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
