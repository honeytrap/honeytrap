package slack_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/slack"
)

const (
	passed = "\u2713"
	failed = "\u2717"
)

type slackService struct{}

func (slackService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/services/343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF" {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TestSlackPusher validates the operational correctness of the
// slack message channel which allows us delivery messages to slack channels
// as describe from the provided configuration.
func TestSlackPusher(t *testing.T) {

	t.Logf("Given the need to post messages to Slack Channels ")
	{

		newSlackChannel := slack.NewMessageChannel()

		t.Logf("\tWhen provided the invalid Slack Webhook credentials")
		{
			_, err := newSlackChannel(map[string]interface{}{
				"hosts":  "slack.com/services/",
				"tokens": "343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
			})

			if err == nil {
				t.Fatalf("\t%s\t Should have sucessfully failed to create a new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have sucessfully failed to create a new MessageChannel.", passed)
		}

		t.Logf("\tWhen provided the correct Slack Webhook credentials")
		{

			// Setup test server for mocking
			server := httptest.NewServer(slackService{})
			host := server.URL + "/services"

			channel, err := newSlackChannel(map[string]interface{}{
				"host":  host,
				"token": "343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have sucessfully created new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have sucessfully created new MessageChannel.", passed)

			channel.Send([]*pushers.PushMessage{
				&pushers.PushMessage{
					Sensor:      "BlueChip",
					Category:    "Chip Integrated",
					SessionID:   "4334334-3433434-34343-FUD",
					ContainerID: "56454-5454UDF-2232UI-34FGHU",
					Data:        "Hello World!",
				},
			})
		}

	}
}
