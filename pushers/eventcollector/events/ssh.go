package events

import (
	"fmt"
	"regexp"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

type EventMetadataSSH struct {
	SourcePort uint `json:"source_port"`
	DestinationIP string `json:"dest_ip"`
	DestinationPort uint `json:"dest_port"`
	SessionID string `json:"session_id"`
	Username string `json:"username"`
	Token string `json:"token"`
	AuthType string `json:"auth_type"`  // publickey-authentication | password-authentication
	PublicKey string `json:"public_key"`
	PublicKeyType string `json:"public_key_type"` // ssh-rsa | ...
	Password string `json:"password"`
	ChannelState string `json:"channel_state"` // open |
}

var sshSessions map[string]SSHSession

type SSHSession struct {
	SessionID string `json:"session-id"`
	SourceIP string `json:"source-ip"`
	DestinationIP string `json:"destination-ip"`
	SourcePort uint `json:"source-port"`
	DestinationPort uint `json:"destination-port"`
	Token string `json:"token"`
	AuthAttempts []SSHSessionAuth `json:"auth-attempts"`
	AuthSuccess bool `json:"auth-success"`
	AuthFailCount uint `json:"auth-fail-count"`
	Payload []byte `json:"payload"`
	Recording []byte `json:"recording"`
	EventCount uint `json:"event-count"`
	CreationDate string `json:"creation-date"`
	LastUpdateDate string `json:"last-update-date"`
}

type SSHSessionAuth struct {
	Timestamp string `json:"timestamp"`
	AuthType string `json:"auth-type"`
	Username string `json:"username"`
	Password string `json:"password"`
	PublicKey string `json:"public-key"`
	PublicKeyType string `json:"public-key-type"`
}



func ProcessEventSSH(e map[string]interface{}) EventMetadataSSH {
	var session SSHSession

	sessionID := fmt.Sprintf("%v", e["ssh.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	if s, ok := sshSessions[sessionID]; ok { // session already being handled
		session = s
		session.LastUpdateDate = fmt.Sprintf("%v", e["date"])

	} else {
		session = SSHSession{
			SessionID: fmt.Sprintf("%v", e["ssh.sessionid"]),
			SourceIP: fmt.Sprintf("%v", e["source-ip"]),
			SourcePort: uint(e["source-port"].(int)),
			DestinationIP: fmt.Sprintf("%v", e["destination-ip"]),
			DestinationPort: uint(e["destination-port"].(int)),
			Token: fmt.Sprintf("%v", e["token"]),
			CreationDate: fmt.Sprintf("%v", e["date"]),
			LastUpdateDate: fmt.Sprintf("%v", e["date"]),
		}
	}

	switch eventType {

	case "publickey-authentication":
		authAttempt := SSHSessionAuth{
			AuthType:      eventType,
			Username:	   fmt.Sprintf("%v", e["ssh.username"]),
			Password:      "",
			PublicKey:     fmt.Sprintf("%v", e["ssh.publickey"]),
			PublicKeyType: fmt.Sprintf("%v", e["ssh.publickey-type"]),
		}
		session.AuthAttempts = append(session.AuthAttempts, authAttempt)

	case "password-authentication":
		authAttempt := SSHSessionAuth{
			AuthType: eventType,
			Username: fmt.Sprintf("%v", e["ssh.username"]),
			Password: fmt.Sprintf("%v", e["ssh.password"]),
		}
		session.AuthAttempts = append(session.AuthAttempts, authAttempt)

	case "ssh-channel":
		session.AuthSuccess = true
		session.AuthFailCount = uint(len(session.AuthAttempts) - 1)

	case "ssh-request":
		session.Payload = append(session.Payload, []byte(fmt.Sprintf("%v", e["ssh.payload"]))...)

	case "ssh-session":
		sRecording := StripANSI(fmt.Sprintf("%v", e["ssh.recording"]))
		session.Recording = append(session.Recording, sRecording...)
	}

	session.EventCount++

	eventMetadata, ok := digestMetadata(e)
	if !ok {
		log.Errorf("Failed to digest SSH metadata")
	}

	return eventMetadata
}

func StripANSI(str string) string {
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}

func digestMetadata(e map[string]interface{}) (EventMetadataSSH, bool) {
	metadata := EventMetadataSSH{
		SessionID:       fmt.Sprintf("%v", e["ssh.sessionid"]),
		SourcePort:      uint(e["source-port"].(int)),
		DestinationIP:   fmt.Sprintf("%v", e["destination-ip"]),
		DestinationPort: uint(e["destination-port"].(int)),
		Token:           fmt.Sprintf("%v", e["token"]),
		TransactionType:        fmt.Sprintf("%v", e["type"]),
	}

	switch metadata.AuthType {
	case "publickey-authentication":
		metadata.PublicKey = fmt.Sprintf("%v", e["ssh.publickey"])
		metadata.PublicKeyType = fmt.Sprintf("%v", e["ssh.publickey-type"])
		metadata.Username = fmt.Sprintf("%v", e["ssh.username"])
	case "password-authentication":
		metadata.Password = fmt.Sprintf("%v", e["ssh.password"])
		metadata.Username = fmt.Sprintf("%v", e["ssh.username"])
	default:
		return metadata, false
	}
	return metadata, true
}