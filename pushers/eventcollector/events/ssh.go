package events

import (
	"fmt"
	"github.com/honeytrap/honeytrap/pushers/eventcollector/models"
	"regexp"
	"strconv"
	"strings"
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
		sRecording := StripANSI(e["ssh.recording"].(string))
		sshSession.Recording = DigestRecording(sRecording)
		eventMetadataSSH.EventType = "session_report"
		eventMetadataSSH.Recording = sRecording

	}
	//

	return
}

func DigestRecording(recording string) (digest []models.SessionSSHRecording) {

	fmt.Printf("THE RECORDING BEFORE:\n%v\n\n", recording)

	// find matches based on regex
	re := regexp.MustCompile(`(?:\.wait\(.*?\)\.put\('(?P<content>(?:.|\r|\n)+?)'\))+?`)
	re_prompt := regexp.MustCompile(`(?:.+?@.+\$)`)
	names := re.SubexpNames()
	matches := re.FindAllStringSubmatch(recording, -1)

	tokens := []string{}
	fmt.Printf("RECORDING AFTER FORMAT:\n")
	for x, n := range matches {

		m := map[string]string{}
		for i, n := range n {
			m[names[i]] = n
		}

		tokens = append(tokens, m["content"])
		fmt.Printf("[%d] --> %v\n", x, m["content"])
	}

	//aggregate commands and separate them by "<br>"

	//commands := map[string]string{}
	input := false

	command := ""
	text := ""
	autoMessageCount := 0
	commandIndex := 0

	addCommand := func(index int, command string, output string) {
		commandDigest := models.SessionSSHRecording {
			Index: index,
			Command: command,
			Output:  output,
		}
		digest = append(digest, commandDigest)
	}

	for x, t := range tokens {

		log.Debugf("-> Processing token %d: %v\n", x, t)
		//check prompt
		promptMatch := re_prompt.FindStringSubmatch(t) // if matches prompt, switch to input mode
		if len(promptMatch) != 0 {
			input = true
			continue
		}

		index := strings.LastIndex(t, "<br/>")
		if index == -1 { // "<br/>" not found -> aggregate
			text += t
			continue

		} else { // <br/> found -> aggregate

			if input { // if in input mode -> create command
				command += text + t[0:index] // aggregate until <br/> placeholder
				command = strings.Replace(command, "<backspace>", "\b", -1)
				input = false

			} else {
				text = t[0:index]// aggregate until <br/> placeholder and sanitize
				text = strings.Replace(text, "<br/>", "\n", -1)

				if len(command) > 0 {
					//commands[command] = text
					addCommand(commandIndex, command, text)
					commandIndex += 1
					command = ""
				} else {
					autoMessageCount += 1
					addCommand(commandIndex, "", text)
					//commands[fmt.Sprintf("<Auto message #%d>", autoMessageCount)] = text
					autoMessageCount += 1
					commandIndex += 1

				}
				text = ""
			}

		}



	}

/*	fmt.Printf("COMMANDS:\n")

	for x, c := range commands {
		fmt.Printf("Command %s: %s\n", x, c)
	}

	i := 1
	for c, o := range commands {
		commandDigest := models.SessionSSHRecording {
			Index: i,
			Command: c,
			Output:  o,
		}
		i++
		digest = append(digest, commandDigest)
	}*/

	return digest
}





