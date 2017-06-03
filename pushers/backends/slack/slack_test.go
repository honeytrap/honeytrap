package slack_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/backends/slack"
	"github.com/honeytrap/honeytrap/pushers/event"
	"github.com/honeytrap/honeytrap/utils/tests"
)

const (
	passed = "\u2713"
	failed = "\u2717"
)

var (
	blueChip = event.New(
		event.Sensor("BlueChip"),
		event.Category("Chip Integrated"),
	)

	ping = event.New(
		event.Sensor("Ping"),
		event.Category("Ping Notification"),
	)

	crum = event.New(
		event.Sensor("Crum Stream"),
		event.Category("WebRTC Crum Stream"),
	)
)

type slackService struct {
	Body bytes.Buffer
}

func (s *slackService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/services/343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF" {
		w.WriteHeader(404)
		return
	}

	io.Copy(&s.Body, r.Body)
	r.Body.Close()

	w.WriteHeader(http.StatusCreated)
}

type anySlackService struct {
	Body  bytes.Buffer
	Token string
}

func (s *anySlackService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, "/services") {
		w.WriteHeader(404)
		return
	}

	s.Token = strings.TrimPrefix(r.URL.Path, "/services/")

	io.Copy(&s.Body, r.Body)
	r.Body.Close()

	w.WriteHeader(http.StatusCreated)
}

// TestSlackPusher validates the operational correctness of the
// slack message channel which allows us delivery messages to slack channels
// as describe from the provided configuration.
func TestSlackPusher(t *testing.T) {

	t.Logf("Given the need to post messages to Slack Backends ")
	{

		t.Logf("\tWhen provided the correct Slack Webhook credentials")
		{

			// Setup test service and server for mocking.
			var service slackService

			server := httptest.NewServer(&service)

			channel := slack.New(slack.Config{
				WebhookURL: server.URL + "/services/343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
			})

			channel.Send(blueChip)

			response := make(map[string]interface{})
			if err := json.NewDecoder(&service.Body).Decode(&response); err != nil {
				t.Fatalf("\t%s\t Should have successfully unmarshalled service body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully unmarshalled service body.", passed)

			attachments, ok := response["attachments"].([]interface{})
			if !ok {
				t.Fatalf("\t%s\t Should have successfully retrieved attachments.", failed)
			}
			t.Logf("\t%s\t Should have successfully retrieved attachments.", passed)

			if len(attachments) < 1 {
				t.Fatalf("\t%s\t Should have successfully retrieved 1 attachments.", failed)
			}
			t.Logf("\t%s\t Should have successfully retrieved 1 attachments.", passed)

			attached, ok := attachments[0].(map[string]interface{})
			if !ok {
				t.Fatalf("\t%s\t Should have successfully retrieved first attachement.", failed)
			}
			t.Logf("\t%s\t Should have successfully retrieved first attachement.", passed)

			fields, ok := attached["fields"].([]interface{})
			if !ok {
				t.Fatalf("\t%s\t Should have successfully retrieved fields.", failed)
			}
			t.Logf("\t%s\t Should have successfully retrieved fields.", passed)

			if len(fields) < 4 {
				t.Fatalf("\t%s\t Should have successfully retrieved 4 fields.", failed)
			}
			t.Logf("\t%s\t Should have successfully retrieved 4 fields.", passed)
		}
	}
}

func TestSlackGenerator(t *testing.T) {
	tomlConfig := `
	backend = "slack"
	webhookURL = "https://hooks.slack.com/services/KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"`

	var config toml.Primitive

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	var backend = struct {
		Backend string `toml:"backend"`
	}{}

	if err := meta.PrimitiveDecode(config, &backend); err != nil {
		tests.Failed("Should have successfully parsed backend name.")
	}
	tests.Passed("Should have successfully parsed backend name.")

	if backend.Backend != "slack" {
		tests.Failed("Should have properly unmarshalled value of config.Backend.")
	}
	tests.Passed("Should have properly unmarshalled value of config.Backend.")

	if _, err := slack.NewWith(meta, config); err != nil {
		tests.Failed("Should have successfully created new  backend: %+q.", err)
	}
	tests.Passed("Should have successfully created new  backend.")

}
