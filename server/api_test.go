package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/director"
	"github.com/honeytrap/honeytrap/director/iodirector"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/honeytrap/honeytrap/server"
	"github.com/honeytrap/honeytrap/utils/tests"
)

const (
	passed = "\u2713"
	failed = "\u2717"
	dbName = "4534-pid"
)

var (
	conso = message.PushMessage{
		Sensor:      "Zu",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor: message.ServiceSensor,
			Type:   message.ServiceStarted,
		},
	}

	conlo = message.PushMessage{
		Sensor:      "Chip",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor: message.ConnectionSensor,
			Type:   message.ContainerCloned,
		},
	}

	conco = message.PushMessage{
		Sensor:      "Cuj",
		Event:       true,
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor: message.PingSensor,
			Type:   message.PingEvent,
		},
	}

	conzip = message.PushMessage{
		Sensor:      "Cuj Sip",
		Event:       true,
		Category:    "Integrated OS",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor: message.SessionSensor,
			Type:   message.ServiceStarted,
		},
	}

	contar = message.PushMessage{
		Sensor:      "Cuj Hul",
		Event:       true,
		Category:    "Integrated OS",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data: message.Event{
			Sensor: message.SessionSensor,
			Type:   message.ContainerCheckpoint,
		},
	}
)

func TestHoneycast(t *testing.T) {
	conf := &config.Config{Token: dbName}
	dir := iodirector.New(conf, nil)
	manager := director.NewContainerConnections()
	cast := server.NewHoneycast(conf, manager, dir)

	defer os.Remove(dbName + "-bolted.db")

	sm := httptest.NewServer(cast)

	cast.Send([]message.PushMessage{conso, conco, conlo, contar, conzip})

	t.Logf("Given the an instance of a Honeycast API ")
	{

		t.Logf("\t When retrieving events from the /sessions endpoints")
		{

			var event server.EventRequest
			event.Page = -1
			event.ResponsePerPage = 24

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(event); err != nil {
				t.Fatalf("\t%s\t Should have successfully created event body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created event body.", passed)

			req, err := http.NewRequest("GET", sm.URL+"/sessions", &buf)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created request.", passed)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully made request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully made request.", passed)

			defer res.Body.Close()

			var item server.EventResponse

			if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
				t.Fatalf("\t%s\t Should have successfully decoded response: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully decoded response.", passed)

			if len(item.Events) != 2 {
				t.Fatalf("\t%s\t Should have retrieved 2 event for /sessions: %d.", failed, len(item.Events))
			}
			t.Logf("\t%s\t Should have retrieved 2 event for /sessions.", passed)

			if item.Total != 2 {
				t.Fatalf("\t%s\t Should have total of 2 events in store: %d.", failed, item.Total)
			}
			t.Logf("\t%s\t Should have total of 2 events in store.", passed)
		}

		t.Logf("\t When retrieving events from the /events endpoints")
		{
			var event server.EventRequest
			event.Page = -1
			event.ResponsePerPage = 24

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(event); err != nil {
				t.Fatalf("\t%s\t Should have successfully created event body: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created event body.", passed)

			req, err := http.NewRequest("GET", sm.URL+"/events", &buf)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created request.", passed)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully made request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully made request.", passed)

			defer res.Body.Close()

			var item server.EventResponse

			if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
				t.Fatalf("\t%s\t Should have successfully decoded response: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully decoded response.", passed)

			if len(item.Events) != 5 {
				t.Fatalf("\t%s\t Should have retrieved 5 event for /events: %d.", failed, len(item.Events))
			}
			t.Logf("\t%s\t Should have retrieved 5 event for /events.", passed)

			if item.Total != 5 {
				t.Fatalf("\t%s\t Should have total of 5 events in store: %d.", failed, item.Total)
			}
			t.Logf("\t%s\t Should have total of 5 events in store.", passed)
		}

		t.Logf("\t When retrieving metric data from the /metrics/containers endpoints")
		{
			req, err := http.NewRequest("GET", sm.URL+"/metrics/containers", nil)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully created request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully created request.", passed)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully made request: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully made request.", passed)

			defer res.Body.Close()

			var item server.ContainerResponse

			if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
				t.Fatalf("\t%s\t Should have successfully decoded response: %q.", failed, err.Error())
			}
			t.Logf("\t%s\t Should have successfully decoded response.", passed)

			if item.Total != 0 {
				t.Fatalf("\t%s\t Should have retrieved 0 container metrics for /metrics/containers: %d.", failed, item.Total)
			}
			t.Logf("\t%s\t Should have retrieved 0 container metrics for /metrics/containers.", passed)
		}
	}
}

func TestHoneycastWebsocket(t *testing.T) {
	conf := &config.Config{Token: dbName}
	dir := iodirector.New(conf, nil)
	manager := director.NewContainerConnections()
	cast := server.NewHoneycast(conf, manager, dir)

	defer os.Remove(dbName + "-bolted.db")

	sm := httptest.NewServer(cast)

	cast.Send([]message.PushMessage{conso, conco, conlo, conzip, contar})

	t.Logf("Given the an instance of a Honeycast API ")
	{
		t.Logf("\t When connecting with websocket to the /ws endpoints ")
		{

			wsPath := sm.URL + "/ws"
			wsURL, _ := url.Parse(wsPath)
			wsURL.Scheme = "ws"

			conn, _, err := connectWS(wsURL.String(), nil)
			if err != nil {
				tests.Failed("Should have successfully connected with a websocket client to %q: %+q.", wsURL.String(), err)
			}
			tests.Passed("Should have successfully connected with a websocket client")

			defer conn.WriteMessage(websocket.CloseMessage, nil)
			defer conn.Close()

			{
				if serr := send(conn, server.FetchSessions, nil); serr != nil {
					tests.Failed("Should have successfully delivered message to a websocket server")
				}
				tests.Passed("Should have successfully delivered message to a websocket server")

				response, err := read(conn)
				if err != nil {
					tests.Failed("Should have successfully read response from websocket client")
				}
				tests.Passed("Should have successfully read response from websocket client")

				if response.Type != server.FetchSessionsReply {
					t.Logf("\t\t Response: %+#q", response)
					tests.Failed("Should have successfully received FetchSessionReply from websocket server")
				}
				tests.Passed("Should have successfully received FetchSessionReply from websocket server")
			}

			{
				if serr := send(conn, server.FetchEvents, nil); serr != nil {
					tests.Failed("Should have successfully delivered message to a websocket server")
				}
				tests.Passed("Should have successfully delivered message to a websocket server")

				response, err := read(conn)
				if err != nil {
					tests.Failed("Should have successfully read response from websocket client")
				}
				tests.Passed("Should have successfully read response from websocket client")

				if response.Type != server.FetchEventsReply {
					t.Logf("\t\t Response: %+#q", response)
					tests.Failed("Should have successfully received FetchEventsReply from websocket server")
				}
				tests.Passed("Should have successfully received FetchEventsReply from websocket server")
			}

			{
				if serr := send(conn, server.MessageType(20), nil); serr != nil {
					tests.Failed("Should have successfully delivered message to a websocket server")
				}
				tests.Passed("Should have successfully delivered message to a websocket server")

				response, err := read(conn)
				if err != nil {
					tests.Failed("Should have successfully read response from websocket client")
				}
				tests.Passed("Should have successfully read response from websocket client")

				if response.Type != server.ErrorResponse {
					t.Logf("\t\t Response: %+#q", response)
					tests.Failed("Should have successfully received ErrorResponse from websocket server")
				}
				tests.Passed("Should have successfully received ErrorResponse from websocket server")
			}

		}
	}
}

var dailer = websocket.Dialer{}

// connectWS connects to the giving URL to create a websocket connection.
func connectWS(url string, headers map[string]string) (*websocket.Conn, *http.Response, error) {
	header := make(http.Header)

	for key, val := range headers {
		header.Set(key, val)
	}

	return dailer.Dial(url, header)
}

func send(conn *websocket.Conn, messageType server.MessageType, payload interface{}) error {
	tests.Info("Sending Command: %d - Data: %+q", messageType, payload)

	var bu bytes.Buffer

	if err := json.NewEncoder(&bu).Encode(server.Message{
		Payload: payload,
		Type:    messageType,
	}); err != nil {
		return err
	}

	return conn.WriteMessage(websocket.BinaryMessage, bu.Bytes())
}

// read reads the incoming response from the websocket connection.
func read(conn *websocket.Conn) (server.Message, error) {
	var message server.Message

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return server.Message{}, err
	}

	var bu bytes.Buffer
	bu.Write(msg)

	if err := json.NewDecoder(&bu).Decode(&message); err != nil {
		return server.Message{}, err
	}

	return message, nil
}
