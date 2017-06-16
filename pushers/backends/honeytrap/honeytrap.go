package honeytrap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/api"
	"github.com/honeytrap/honeytrap/pushers/event"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:channels:honeytrap")

// TrapConfig defines the configuration used to setup a TrapBackend.
type TrapConfig struct {
	Host  string `toml:"host"`
	Token string `toml:"token"`
}

// TrapBackend defines a struct which implmenets the pushers.Backend
// interface for delivery honeytrap special messages.
type TrapBackend struct {
	client *api.Client
}

// New returns a new instance of a TrapBackend.
func New(config TrapConfig) TrapBackend {
	return TrapBackend{
		client: api.New(&api.Config{
			Url:   config.Host,
			Token: config.Token,
		}),
	}
}

// NewWith defines a function to return a pushers.Backend which delivers
// new messages to a giving underline honeytrap API defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig TrapConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.Host == "" {
		return nil, errors.New("honeytrap.TrapConfig Invalid: Host can not be empty")
	}

	if apiconfig.Token == "" {
		return nil, errors.New("honeytrap.TrapConfig Invalid: Token can not be empty")
	}

	return New(apiconfig), nil
}

func init() {
	pushers.RegisterBackend("honeytrap", NewWith)
}

// Send delivers all messages to the underline connection.
func (hc TrapBackend) Send(message *event.Event) {
	var err error
	var req *http.Request

	category := message["cateory"]
	sensor := message["sensor"]

	var jsData bytes.Buffer

	if err := json.NewEncoder(&jsData).Encode(message.Map()); err != nil {
		log.Errorf("HoneytrapBackend: Error while marshalling to json: %s", err.Error())
		return
	}

	if sensor == "honeytrap" && category == "ping" {
		req, err = hc.client.NewRequest("POST", "v1/ping", nil)
	} else {
		// TODO: workaround, need to update api
		req, err = hc.client.NewRequest("POST", "v1/action", &jsData)
	}

	if err != nil {
		log.Errorf("HoneytrapBackend: Error while preparing request: %s", err.Error())
		return
	}

	var resp *http.Response
	if resp, err = hc.client.Do(req, nil); err != nil {
		log.Errorf("HoneytrapBackend: Error while sending message: %s", err.Error())
		return
	}

	// for keep-alive
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("HoneytrapBackend: Unexpected status code: %d", resp.StatusCode)
		return
	}
}
