package slack_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/slack"
)

const (
	passed = "\u2713"
	failed = "\u2717"
)

var (
	blueChip = &pushers.PushMessage{
		Sensor:      "BlueChip",
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	ping = &pushers.PushMessage{
		Sensor:      "Ping",
		Category:    "Ping Notificiation",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	crum = &pushers.PushMessage{
		Sensor:      "Crum Stream",
		Category:    "WebRTC Crum Stream",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}
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
				t.Fatalf("\t%s\t Should have successfully failed to create new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully failed to create new MessageChannel.", passed)
		}

		t.Logf("\tWhen provided the correct Slack Webhook credentials")
		{

			// Setup test service and server for mocking.
			var service slackService

			server := httptest.NewServer(&service)
			host := server.URL + "/services"

			channel, err := newSlackChannel(map[string]interface{}{
				"host":  host,
				"token": "343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created new MessageChannel.", passed)

			channel.Send([]*pushers.PushMessage{blueChip})

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

		t.Logf("\tWhen provided Slack Webhook credentials with field filters")
		{
			// Setup test service and server for mocking.
			var service slackService

			server := httptest.NewServer(&service)
			host := server.URL + "/services"

			channel, err := newSlackChannel(map[string]interface{}{
				"host":  host,
				"token": "343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
				"fields": map[string]interface{}{
					"sensor":   "[^ping]",
					"category": "[^WebRTC]",
				},
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created new MessageChannel.", passed)

			response := make(map[string]interface{})

			channel.Send([]*pushers.PushMessage{ping, crum})
			if err := json.NewDecoder(&service.Body).Decode(&response); err != nil {
				t.Fatalf("\t%s\t Should have successfully failed to unmarshalled empty body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully failed to unmarshalled empty body.", passed)

			response = make(map[string]interface{})

			channel.Send([]*pushers.PushMessage{crum})
			if err := json.NewDecoder(&service.Body).Decode(&response); err != nil {
				t.Fatalf("\t%s\t Should have successfully failed to unmarshalled empty body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully failed to unmarshalled empty body.", passed)
		}

		t.Logf("\tWhen provided Slack Webhook credentials with field filters and channel selector")
		{
			// Setup test service and server for mocking.
			var service anySlackService

			server := httptest.NewServer(&service)
			host := server.URL + "/services"

			channel, err := newSlackChannel(map[string]interface{}{
				"host":  host,
				"token": "343HJUYFHGT/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
				"fields": map[string]interface{}{
					"sensor": "[ping|crum]",
				},
				"channels": []map[string]interface{}{
					{
						"field":   "sensor",
						"value":   "Ping",
						"channel": "ping-channel",
						"token":   "545JUHB89B/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
					},
					{
						"field":   "sensor",
						"value":   "Crum",
						"channel": "crum-channel",
						"token":   "JUC9OP90/B4545IO/VOOepdacxW9HG60eDfoFBiMF",
					},
				},
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created new MessageChannel: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created new MessageChannel.", passed)

			channel.Send([]*pushers.PushMessage{ping})

			pingResponse := make(map[string]interface{})
			if err := json.NewDecoder(&service.Body).Decode(&pingResponse); err != nil {
				t.Fatalf("\t%s\t Should have successfully unmarshalled request body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully unmarshalled requested body.", passed)

			if val, ok := pingResponse["channel"]; !ok || val != "ping-channel" {
				t.Fatalf("\t%s\t Should have successfully matched target channel for ping response but got %q.", failed, val)
			}
			t.Logf("\t%s\t Should have successfully matched target channel for ping response.", passed)

			if service.Token != "545JUHB89B/B4545IO/VOOepdacxW9HG60eDfoFBiMF" {
				t.Fatalf("\t%s\t Should have successfully matched request toekn for ping response.", failed)
			}
			t.Logf("\t%s\t Should have successfully matched request toekn for ping response.", passed)

			channel.Send([]*pushers.PushMessage{crum})

			crumResponse := make(map[string]interface{})
			if err := json.NewDecoder(&service.Body).Decode(&crumResponse); err != nil {
				t.Fatalf("\t%s\t Should have successfully unmarshalled empty body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully unmarshalled empty body.", passed)

			if val, ok := crumResponse["channel"]; !ok || val != "crum-channel" {
				t.Fatalf("\t%s\t Should have successfully matched target channel for crum response but got %q.", failed, val)
			}
			t.Logf("\t%s\t Should have successfully matched target channel for ping response.", passed)

			if service.Token != "JUC9OP90/B4545IO/VOOepdacxW9HG60eDfoFBiMF" {
				t.Fatalf("\t%s\t Should have successfully matched request toekn for ping response.", failed)
			}
			t.Logf("\t%s\t Should have successfully matched request toekn for ping response.", passed)

		}
	}
}
