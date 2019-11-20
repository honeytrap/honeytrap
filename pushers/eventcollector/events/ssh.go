package events

import (
	"fmt"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/models"
	"strconv"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"


func ProcessEventSSH(e map[string]interface{}) (sshSession models.SessionSSH, eventMetadataSSH models.EventMetadataSSH, ok bool) {

	sessionID := fmt.Sprintf("%v", e["ssh.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	ok = true

	srcPort, _ := strconv.ParseUint(fmt.Sprintf("%v", e["source-port"]), 10, 64)
	dstPort, _ := strconv.ParseUint(fmt.Sprintf("%v", e["destination-port"]), 10, 64)

	eventMetadataSSH = models.EventMetadataSSH{
		SessionID: sessionID,
		TransactionType: eventType,
		SourcePort: uint(srcPort),
		DestinationIP: fmt.Sprintf("%v", e["destination-ip"]),
		DestinationPort: uint(dstPort),

	}

	if Sessions[sessionID].ServiceMeta == nil {
		sshSession = models.SessionSSH{
			Token:         fmt.Sprintf("%v", e["token"]),
		}
	} else {
		sshSession = Sessions[sessionID].ServiceMeta.(models.SessionSSH)
	}


	switch eventType {

	case "publickey-authentication":
		authAttempt := models.SessionSSHAuth{
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
		authAttempt := models.SessionSSHAuth{
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
		eventMetadataSSH.EventType = "session_handshake"

	case "ssh-session":
		sRecording := StripANSI(strip.StripTags(e["ssh.recording"].(string)))
		sshSession.Recording = fmt.Sprintf("%v%v", sshSession.Recording, sRecording)
		eventMetadataSSH.EventType = "session_report"
		eventMetadataSSH.Recording = sshSession.Recording
	}

	return
}
