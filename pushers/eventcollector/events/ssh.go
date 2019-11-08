package events

import (
	"fmt"
	strip "github.com/grokify/html-strip-tags-go"
	"strconv"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var	SSHSessions map[string]SSHSession


type EventMetadataSSH struct {
	EventType 		string 		`json:"event_type"`
	SourcePort 		uint 		`json:"source_port"`
	DestinationIP 	string 		`json:"dest_ip"`
	DestinationPort uint 		`json:"dest_port"`
	SessionID 		string 		`json:"session_id"`
	Username 		string 		`json:"username"`
	Token 			string 		`json:"token"`
	TransactionType string 		`json:"transaction_type"`  // publickey-authentication | password-authentication | request..
	PublicKey 		string 		`json:"public_key"`
	PublicKeyType 	string 		`json:"public_key_type"` // ssh-rsa | ...
	Password 		string 		`json:"password"`
	Recording 		string 		`json:"recording"`
}

type SSHSession struct {
	Token			string				`json:"token"`
	AuthAttempts	[]SSHSessionAuth	`json:"auth-attempts"`
	AuthSuccess		bool				`json:"auth-success"`
	AuthFailCount	uint				`json:"auth-fail-count"`
	Payload			string				`json:"payload"`
	Recording		string				`json:"recording"`
}

type SSHSessionAuth struct {
	Timestamp 		string 		`json:"timestamp"`
	AuthType 		string 		`json:"auth-type"`
	Username 		string 		`json:"username"`
	Password 		string 		`json:"password"`
	PublicKey 		string 		`json:"public-key"`
	PublicKeyType 	string 		`json:"public-key-type"`
	Success 		bool 		`json:"success"`
}


func ProcessEventSSH(e map[string]interface{}) (sshSession SSHSession, eventMetadataSSH EventMetadataSSH, ok bool) {

	sessionID := fmt.Sprintf("%v", e["ssh.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	ok = true

	srcPort, _ := strconv.ParseUint(fmt.Sprintf("%v", e["source-port"]), 10, 64)
	dstPort, _ := strconv.ParseUint(fmt.Sprintf("%v", e["destination-port"]), 10, 64)

	eventMetadataSSH = EventMetadataSSH{
		SessionID: sessionID,
		TransactionType: eventType,
		SourcePort: uint(srcPort),
		DestinationIP: fmt.Sprintf("%v", e["destination-ip"]),
		DestinationPort: uint(dstPort),

	}

	if Sessions[sessionID].ServiceMeta == nil {
		sshSession = SSHSession{
			Token:         fmt.Sprintf("%v", e["token"]),
		}
	} else {
		sshSession = Sessions[sessionID].ServiceMeta.(SSHSession)
	}


	switch eventType {

	case "publickey-authentication":
		authAttempt := SSHSessionAuth{
			AuthType:      eventType,
			Username:	   fmt.Sprintf("%v", e["ssh.username"]),
			Password:      "",
			PublicKey:     fmt.Sprintf("%v", e["ssh.publickey"]),
			PublicKeyType: fmt.Sprintf("%v", e["ssh.publickey-type"]),
			Timestamp:	   fmt.Sprintf("%v", e["date"]),
		}
		sshSession.AuthAttempts = append(sshSession.AuthAttempts, authAttempt)
		eventMetadataSSH.EventType = "auth_attempt_pubkey"
		eventMetadataSSH.Username = authAttempt.Username
		eventMetadataSSH.PublicKey = authAttempt.PublicKey
		eventMetadataSSH.PublicKeyType = authAttempt.PublicKeyType

	case "password-authentication":
		authAttempt := SSHSessionAuth{
			AuthType: 	   eventType,
			Username: 	   fmt.Sprintf("%v", e["ssh.username"]),
			Password: 	   fmt.Sprintf("%v", e["ssh.password"]),
			Timestamp:	   fmt.Sprintf("%v", e["date"]),
		}
		sshSession.AuthAttempts = append(sshSession.AuthAttempts, authAttempt)
		eventMetadataSSH.EventType = "auth_attempt_passwd"
		eventMetadataSSH.Username = authAttempt.Username
		eventMetadataSSH.Password = authAttempt.Password

	case "ssh-channel":
		sshSession.AuthSuccess = true
		sshSession.AuthFailCount = uint(len(sshSession.AuthAttempts) - 1)
		if len(sshSession.AuthAttempts) < 1 {
			log.Errorf("Handling ssh-channel with no previous auth attempts: %v", sshSession.AuthAttempts)
			ok = false
			break
		}

		lastAuth := &sshSession.AuthAttempts[len(sshSession.AuthAttempts)-1]
		lastAuth.Success = true

		eventMetadataSSH.Username = lastAuth.Username
		switch lastAuth.AuthType {
		case "publickey-authentication":
			eventMetadataSSH.EventType = "auth_success_pubkey"
			eventMetadataSSH.PublicKey = lastAuth.PublicKey
			eventMetadataSSH.PublicKeyType = lastAuth.PublicKeyType
		case "password-authentication":
			eventMetadataSSH.EventType = "auth_success_passwd"
			eventMetadataSSH.Password = lastAuth.Password
		}

	case "ssh-request":
		sshSession.Payload = fmt.Sprintf("%v%v", sshSession.Payload, e["ssh.payload"])

	case "ssh-session":
		sRecording := StripANSI(strip.StripTags(e["ssh.recording"].(string)))
		sshSession.Recording = fmt.Sprintf("%v%v", sshSession.Recording, sRecording)
		eventMetadataSSH.EventType = "session_report"
		eventMetadataSSH.Recording = sshSession.Recording
	}

	return
}
