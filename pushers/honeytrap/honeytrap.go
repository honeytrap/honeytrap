package pushers

import (
	"fmt"

	pushers "github.com/honeytrap/honeytrap/pushers"
	api "github.com/honeytrap/honeytrap/pushers/api"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:honeytrap")

var (
	_ = pushers.Register("honeytrap", NewHoneytrapChannel())
)

type HoneytrapChannel struct {
	client *api.Client
}

func NewHoneytrapChannel() pushers.ChannelFunc {
	return func(conf map[string]interface{}) (pushers.Channel, error) {
		if _, ok := conf["host"]; !ok {
			return nil, fmt.Errorf("Host not set for channel honeytrap")
		}

		if _, ok := conf["token"]; !ok {
			return nil, fmt.Errorf("Token not set for channel honeytrap")
		}

		cl := api.New(&api.Config{
			Url:   conf["host"].(string),
			Token: conf["token"].(string),
		})
		return &HoneytrapChannel{cl}, nil
	}
}

func (hc HoneytrapChannel) Send(messages []*pushers.PushMessage) {
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
			log.Error("HoneytrapChannel: Error while preparing request: %s", err.Error())
			continue
		}

		var resp *http.Response
		if resp, err = hc.client.Do(req, nil); err != nil {
			log.Error("HoneytrapChannel: Error while sending message: %s", err.Error())
			continue
		}

		// for keep-alive
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error("HoneytrapChannel: Unexpected status code: %d", resp.StatusCode)
			continue
		}
	}

	log.Infof("HoneytrapChannel: Sent %d actions.", len(messages))
}
