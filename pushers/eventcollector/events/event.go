// Copyright 2019 Ubiwhere (https://www.ubiwhere.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("channels/eventcollector/event")


var eventIDSeq int = 0

type Event struct {
	EventID uint `json:"event_id"`
	AgentType string  `json:"agent_type"`
	Timestamp string `json:"timestamp"` // ISO 8601
	SourceIP string `json:"sourceip"`
	Count uint `json:"count"`
	Type string `json:"type"`
	Priority string `json:"priority"`
	Name string `json:"name"`
	Context string `json:"context"`
	Metadata interface{} `json:"metadata"`
}

func ProcessEvent(event map[string]interface{}) []byte {

	var event_metadata interface{}

	switch event["category"] {
	case "ssh":
		event_metadata = ProcessEventSSH(event)

	}

	return ComposeEvent(event, event_metadata)
}


func ComposeEvent(event map[string]interface{}, metadata interface{}) []byte {
	eventIDSeq++
	ecEvent := Event{
		EventID: uint(eventIDSeq),
		AgentType: "HONEYNET",
		Timestamp: ConvertDatePseudoISO8601(fmt.Sprintf("%v", event["date"])),
		Count: 1,
		Type: "Notice",
		Priority: "Low",
		Name: "honeynet",
		Context: fmt.Sprintf("%v", event["category"]),
		Metadata: metadata,
	}

	if val, ok := event["source-ip"]; ok {
		ecEvent.SourceIP = fmt.Sprintf("%v", val)
	}

	ecEventJson, err := json.Marshal(ecEvent)
	if err != nil {
		log.Errorf("Failed to compose event: %s", err)
	}

	return ecEventJson
}

func ConvertDatePseudoISO8601(date string) string {
	d, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Errorf("Failed to convert date: %v", err)
	}
	fd := fmt.Sprintf("%d-%02d-%02d %02d-%02d-%02d", d.Year(),  d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second())
	return fd
}