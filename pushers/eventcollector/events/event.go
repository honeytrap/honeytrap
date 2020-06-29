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
	"fmt"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/models"
	"github.com/op/go-logging"
	"regexp"
	"time"
)

var log = logging.MustGetLogger("channels/eventcollector/event")

var supportedServices = []string{"ssh", "telnet", "dns"}

var eventIDSeq int = 0

var Sessions = make(map[string]models.Session)
var Events = make(map[string]models.Event)
var eventModel = new(models.EventModel)
var sessionModel = new(models.SessionModel)


func ProcessEvent(e map[string]interface{}) (session models.Session, event models.Event, ok bool) {
	ok = true
	var eventMetadata interface{}
	var serviceMeta interface{}

	// restrict event processing to known services
	service := fmt.Sprintf("%v", e["category"])
	if !stringInSlice(service, supportedServices) {
		return session, event, false
	}

	eventIDSeq++
	event = models.Event{

		//EventID: fmt.Sprintf("%v", eventIDSeq),
		AgentType: "HONEYNET",
		Timestamp: ConvertDatePseudoISO8601(fmt.Sprintf("%v", e["date"])),
		SourceIP: fmt.Sprintf("%v", e["source-ip"]),
		Count: 0,
		Type: "Notice",
		Priority: "Low",
		Name: "honeynet",
		Context: service,
	}

	newSession := false

	// process session
	sessionID := fmt.Sprintf("%v", e[fmt.Sprintf("%v.sessionid", service)])
	// eventType := fmt.Sprintf("%v", e["type"])

	if s, ok := Sessions[sessionID]; ok {
		session = s
		session.UpdateDate = fmt.Sprintf("%v", e["date"])
		session.EventCount++
		log.Debugf("Handling previous registered session '%v'", sessionID)

	} else {
		newSession = true
		log.Debugf("Creating new handler for session '%v'", sessionID)
		session = models.Session{
			SessionID:       sessionID,
			Service:         service,
			SourceIP:        fmt.Sprintf("%v", e["source-ip"]),
			SourcePort:      uint(e["source-port"].(int)),
			DestinationIP:   fmt.Sprintf("%v", e["destination-ip"]),
			DestinationPort: uint(e["destination-port"].(int)),
			CreationDate:    fmt.Sprintf("%v", e["date"]),
			UpdateDate:      fmt.Sprintf("%v", e["date"]),
			EventCount:      1,
		}
	}

	switch service {
	case "ssh":
		serviceMeta, eventMetadata, ok = ProcessEventSSH(e)
		if !ok {
			log.Errorf("Failed to process event (SSH service) with session '%v'", sessionID)
		}

	case "telnet":
		serviceMeta, eventMetadata, ok = ProcessEventTelnet(e)
		if !ok {
			log.Errorf("Failed to process event (Telnet service) with session '%v'", sessionID)
		}
	}

	event.Metadata = eventMetadata
	fmt.Print(eventMetadata)
	session.ServiceMeta = serviceMeta
	Sessions[sessionID] = session
	eventid := fmt.Sprintf("%v", eventIDSeq)
	Events[eventid] = event

	if newSession {
		err := sessionModel.Create(session)
		if err != nil {
			log.Errorf("Failed to persist session %v: %v", sessionID, err)
		}
	} else {
		err := sessionModel.Update(sessionID, session)
		if err != nil {
			log.Errorf("Failed to update session %v: %v", sessionID, err)
		}
	}

	err := eventModel.Create(event)
	if err != nil {
		log.Errorf("Failed to persist event %v: %v", eventid, err)
	}
	return
}

func ConvertDatePseudoISO8601(date string) string {
	d, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Errorf("Failed to convert date: %v", err)
	}
	fd := fmt.Sprintf("%d-%02d-%02d %02d-%02d-%02d", d.Year(),  d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second())
	return fd
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StripANSI(str string) string {
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}


