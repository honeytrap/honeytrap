package events

import (
	"fmt"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/models"
)

var	TelnetSessions map[string]models.SessionTelnet

func ProcessEventTelnet(e map[string]interface{}) (telnetSession models.SessionTelnet, eventMetadataTelnet models.EventMetadataTelnet, ok bool) {

	sessionID := fmt.Sprintf("%v", e["telnet.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	ok = true

	eventMetadataTelnet = models.EventMetadataTelnet{
		SessionID: sessionID,
		TransactionType: eventType,
	}

	if Sessions[sessionID].ServiceMeta == nil {
		telnetSession = models.SessionTelnet{
			Token:         fmt.Sprintf("%v", e["token"]),
		}
	} else {
		telnetSession = Sessions[sessionID].ServiceMeta.(models.SessionTelnet)
	}

	switch eventType {

	case "password-authentication":
		authAttempt := models.SessionTelnetAuth{
			AuthType: 	   eventType,
			Username: 	   fmt.Sprintf("%v", e["telnet.username"]),
			Password: 	   fmt.Sprintf("%v", e["telnet.password"]),
			Timestamp:	   fmt.Sprintf("%v", e["date"]),
		}
		telnetSession.AuthAttempts = append(telnetSession.AuthAttempts, authAttempt)
		eventMetadataTelnet.EventType = "auth_attempt_passwd"
		eventMetadataTelnet.Username = authAttempt.Username
		eventMetadataTelnet.Password = authAttempt.Password

	case "session":
		if !telnetSession.AuthSuccess {
			lastAuth := &telnetSession.AuthAttempts[len(telnetSession.AuthAttempts)-1]
			lastAuth.Success = true
			telnetSession.AuthSuccess = true
		}

		command := models.SessionTelnetCommand{
			Command:   fmt.Sprintf("%v", e["telnet.command"]),
			Timestamp: fmt.Sprintf("%v", e["date"]),
		}
		telnetSession.Commands = append(telnetSession.Commands, command)
		eventMetadataTelnet.EventType = "session_command"
	}

	return
}