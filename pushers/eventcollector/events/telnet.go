package events

import (
	"fmt"
)

var	TelnetSessions map[string]TelnetSession

type EventMetadataTelnet struct {
	EventType 		string 		`json:"event_type"`
	SourcePort 		uint 		`json:"source_port"`
	DestinationIP 	string 		`json:"dest_ip"`
	DestinationPort uint 		`json:"dest_port"`
	SessionID 		string 		`json:"session_id"`
	Username 		string 		`json:"username"`
	Token 			string 		`json:"token"`
	TransactionType string 		`json:"transaction_type"` // password-authentication | request..
	//PublicKey 		string 		`json:"public_key"`
	//PublicKeyType 	string 		`json:"public_key_type"` // ssh-rsa | ...
	Password 		string 		`json:"password"`
	Recording 		string 		`json:"recording"`
}

type TelnetSession struct {
	Token			string				`json:"token"`
	AuthAttempts	[]TelnetSessionAuth	`json:"auth-attempts"`
	AuthSuccess		bool				`json:"auth-success"`
	AuthFailCount	uint				`json:"auth-fail-count"`
	Commands		[]TelnetCommand		`json:"commands"`
	Payload			string				`json:"payload"`
	Recording		string				`json:"recording"`
}

type TelnetSessionAuth struct {
	Timestamp 		string 		`json:"timestamp"`
	AuthType 		string 		`json:"auth-type"`
	Username 		string 		`json:"username"`
	Password 		string 		`json:"password"`
	//PublicKey 		string 		`json:"public-key"`
	//PublicKeyType 	string 		`json:"public-key-type"`
	Success 		bool 		`json:"success"`
}

type TelnetCommand struct {
	Command 		string		`json:"command"`
	Timestamp 		string 		`json:"timestamp"`
}


func ProcessEventTelnet(e map[string]interface{}) (telnetSession TelnetSession, eventMetadataTelnet EventMetadataTelnet, ok bool) {

	sessionID := fmt.Sprintf("%v", e["telnet.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	ok = true

	eventMetadataTelnet = EventMetadataTelnet{
		SessionID: sessionID,
		TransactionType: eventType,
	}

	if Sessions[sessionID].ServiceMeta == nil {
		telnetSession = TelnetSession{
			Token:         fmt.Sprintf("%v", e["token"]),
		}
	} else {
		telnetSession = Sessions[sessionID].ServiceMeta.(TelnetSession)
	}


	switch eventType {

	/*
	case "publickey-authentication":
		authAttempt := TelnetSessionAuth{
			AuthType:      eventType,
			Username:	   fmt.Sprintf("%v", e["ssh.username"]),
			Password:      "",
			//PublicKey:     fmt.Sprintf("%v", e["ssh.publickey"]),
			//PublicKeyType: fmt.Sprintf("%v", e["ssh.publickey-type"]),
			Timestamp:	   fmt.Sprintf("%v", e["date"]),
		}
		telnetSession.AuthAttempts = append(telnetSession.AuthAttempts, authAttempt)
		eventMetadataTelnet.EventType = "auth_attempt_pubkey"
		eventMetadataTelnet.Username = authAttempt.Username
		//eventMetadataTelnet.PublicKey = authAttempt.PublicKey
		//eventMetadataTelnet.PublicKeyType = authAttempt.PublicKeyType
*/
	case "password-authentication":
		authAttempt := TelnetSessionAuth{
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

		command := TelnetCommand{
			Command:   fmt.Sprintf("%v", e["telnet.command"]),
			Timestamp: fmt.Sprintf("%v", e["date"]),
		}
		telnetSession.Commands = append(telnetSession.Commands, command)
		eventMetadataTelnet.EventType = "session_command"
	}

	return
}