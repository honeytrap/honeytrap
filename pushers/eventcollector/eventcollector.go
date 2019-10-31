// Copyright 2019 Ubiwhere (https://dutchsec.com/)
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

package eventcollector

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	sarama "github.com/Shopify/sarama"
	"regexp"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("eventcollector", New)
)

var log = logging.MustGetLogger("channels/eventcollector")

// Backend defines a struct which provides a channel for delivery
// push messages to the EventCollector brokers (kafka).
type Backend struct {
	Config
	producer sarama.AsyncProducer
	ch chan map[string]interface{}
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"


func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	c := Backend{
		ch: ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.NoResponse

	producer, err := sarama.NewAsyncProducer(c.Brokers, config)
	if err != nil {
		return nil, err
	}

	c.producer = producer

	go c.run()
	return &c, nil
}

func printify(s string) string {
	o := ""
	for _, rune := range s {
		if !unicode.IsPrint(rune) {
			buf := make([]byte, 4)

			n := utf8.EncodeRune(buf, rune)
			o += fmt.Sprintf("\\x%s", hex.EncodeToString(buf[:n]))
			continue
		}

		o += string(rune)
	}
	return o
}

func (hc Backend) run() {
	defer hc.producer.AsyncClose()

	for e := range hc.ch {
		var params []string

		ProcessEvent(e)

		for k, v := range e {

			switch x := v.(type) {
			case net.IP:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case uint32, uint16, uint8, uint,
				int32, int16, int8, int:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case time.Time:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case string:
				params = append(params, fmt.Sprintf("%s=%s", k, printify(x)))
			default:
				params = append(params, fmt.Sprintf("%s=%#v", k, v))
			}
		}
		sort.Strings(params)
		message := &sarama.ProducerMessage{
			Topic: hc.Topic,
			Key:   nil,
			Value: sarama.StringEncoder(strings.Join(params, ", ")),
		}
		select {
		case hc.producer.Input() <- message:
			log.Debug("Produced message:\n", message)
		case err := <- hc.producer.Errors():
			log.Error("Failed to commit message: ", err)
		}
	}
}

// Send delivers the giving push messages into the internal elastic search endpoint.
func (hc Backend) Send(message event.Event) {
	mp := make(map[string]interface{})

	message.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	hc.ch <- mp
}

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

type MetadataSSH struct {
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
	SessionID string
	SourceIP string
	DestinationIP string
	SourcePort uint
	DestinationPort uint
	Username string
	Token string
	AuthAttempts []SSHSessionAuth
	AuthSuccess bool
	AuthFailCount uint
	Payload []byte
	Recording []byte
	NumberEvents uint
}

type SSHSessionAuth struct {
	AuthType string
	Password string
	PublicKey string
	PublicKeyType string
}

func ProcessEvent(e map[string]interface{}) []byte {
	switch e["category"] {
	case "ssh":
		session := ProcessEventSSH(e)
		ComposeEvent(session)
	}

	return nil
}

func ProcessEventSSH(e map[string]interface{}) SSHSession {
	var session SSHSession

	sessionID := fmt.Sprintf("%v", e["ssh.sessionid"])
	eventType := fmt.Sprintf("%v", e["type"])

	if s, ok := sshSessions[sessionID]; ok { // session already being handled
		session = s

	} else {
		session = SSHSession{
			SessionID: fmt.Sprintf("%v", e["ssh.sessionid"]),
			SourceIP: fmt.Sprintf("%v", e["source-ip"]),
			SourcePort: uint(e["source-port"].(int)),
			DestinationIP: fmt.Sprintf("%v", e["destination-ip"]),
			DestinationPort: uint(e["destination-port"].(int)),
			Username: fmt.Sprintf("%v", e["ssh.username"]),
			Token: fmt.Sprintf("%v", e["token"]),
		}
	}

	switch eventType {

	case "publickey-authentication":
		authAttempt := SSHSessionAuth{
			AuthType:      eventType,
			Password:      "",
			PublicKey:     fmt.Sprintf("%v", e["ssh.publickey"]),
			PublicKeyType: fmt.Sprintf("%v", e["ssh.publickey-type"]),
		}
		session.AuthAttempts = append(session.AuthAttempts, authAttempt)

	case "password-authentication":
		authAttempt := SSHSessionAuth{
			AuthType: eventType,
			Password: fmt.Sprintf("%v", e["ssh.password"]),
		}
		session.AuthAttempts = append(session.AuthAttempts, authAttempt)

	case "ssh-channel":
		session.AuthSuccess = true
		session.AuthFailCount = uint(len(session.AuthAttempts) - 1)

	case "ssh-request":
		session.Payload = append(session.Payload, []byte(fmt.Sprintf("%v", e["ssh.payload"]))...)

	case "ssh-session":
		fmt.Println("recording type: ", reflect.TypeOf(e["ssh.recording"]))
		sRecording := StripANSI(fmt.Sprintf("%v", e["ssh.recording"]))
		fmt.Printf("rec: %s", sRecording)
		session.Recording = append(session.Recording, sRecording...)
	}

	session.NumberEvents++
	return session
}



func StripANSI(str string) string {
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}


func ComposeEvent(sshSession SSHSession) []byte {

	var metadata interface{}
	var ok bool

	switch e["category"] {
	case "ssh":
		if metadata, ok = DigestSSH(e); !ok {
			log.Errorf("Failed to digest SSH parameters")
			return nil
		}
	default:
		return nil
	}
	fmt.Println("Got category ", e["category"])

	ecEvent := Event {
		EventID: 1,
		AgentType: "HONEYNET",
		Timestamp: fmt.Sprintf("%v", e["date"]),
		SourceIP: fmt.Sprintf("%v", e["source-ip"]),
		Count: 1,
		Type: "Notice",
		Priority: "Low",
		Name: "honeynet",
		Context: fmt.Sprintf("%v", e["category"]),
		Metadata: metadata,
	}

	ecEventJson, err := json.Marshal(ecEvent)
	if err != nil {
		log.Errorf("Failed to compose event: %s", err)
	}
	fmt.Println("\n\n-> Marshaled event:")
	fmt.Println(string(ecEventJson))
	fmt.Println("<- End of Marshaled event\n\n")

	return ecEventJson
}

func DigestSSH(e map[string]interface{}) (MetadataSSH, bool) {
	metadata := MetadataSSH {
		SessionID: fmt.Sprintf("%v", e["ssh.sessionid"]),
		SourcePort: uint(e["source-port"].(int)),
		DestinationIP: fmt.Sprintf("%v", e["destination-ip"]),
		DestinationPort: uint(e["destination-port"].(int)),
		Username: fmt.Sprintf("%v", e["ssh.username"]),
		Token: fmt.Sprintf("%v", e["token"]),
		AuthType: fmt.Sprintf("%v", e["type"]),
	}

	switch metadata.AuthType {
	case "publickey-authentication":
		metadata.PublicKey = fmt.Sprintf("%v", e["ssh.publickey"])
		metadata.PublicKeyType = fmt.Sprintf("%v", e["ssh.publickey-type"])
	case "password-authentication":
		metadata.Password = fmt.Sprintf("%v", e["ssh.password"])
	default:
		return metadata, false
	}
	return metadata, true
}