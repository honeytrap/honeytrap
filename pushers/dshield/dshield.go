/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package dshield

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"

	"time"

	"github.com/honeytrap/honeytrap/cmd"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

	"io"

	"encoding/base64"

	"crypto/rand"

	logging "github.com/op/go-logging"
)

var (
	_ = pushers.Register("dshield", New)
)

var log = logging.MustGetLogger("channels/dshield")

/*
Configuration example:

[channel.dshield]
type="dshield"
user_id="{userid}"
api_key="{api_key}"
*/

// Backend defines a struct which provides a channel for delivery
// push messages to an elasticsearch api.
type Backend struct {
	Config

	MyIP string

	ch chan json.Marshaler
}

var types = []string{"email", "firewall", "sshlogin", "telnetlogin", "404report", "httprequest", "webhoneypot"}

func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan json.Marshaler, 100)

	c := Backend{
		ch: ch,
		Config: Config{
			URL: "https://www.dshield.org/",
		},
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	myIP, err := GetMyIP()
	if err != nil {
		return nil, err
	}

	log.Debug("DShield retrieved cientip: %s", myIP)

	c.MyIP = myIP

	if c.UserID == "" {
		log.Warning("DShield userid not set.")
	}

	if c.APIKey == "" {
		log.Warning("DShield apikey not set.")
	}

	go c.run()

	return &c, nil
}

func Insecure(config *tls.Config) *tls.Config {
	config.InsecureSkipVerify = true
	return config
}

func GetMyIP() (string, error) {
	resp, err := http.Get("https://www.dshield.org/api/myip?json")
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Could not retrieve ip.")
	}

	result := struct {
		IP string `json:"ip"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.IP, nil

}

func (hc Backend) MakeAuthHeader() (string, error) {
	nonce := make([]byte, 8)

	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	nonceStr := base64.StdEncoding.EncodeToString(nonce)

	key := append(nonce, fmt.Sprintf("%s", hc.UserID)...)

	message, err := base64.StdEncoding.DecodeString(hc.APIKey)
	if err != nil {
		return "", err
	}

	sig := hmac.New(sha256.New, key)
	sig.Write(message)

	digest := base64.StdEncoding.EncodeToString(sig.Sum(nil))
	return fmt.Sprintf("ISC-HMAC-SHA256 credentials=%s nonce=%s userid=%s", digest, nonceStr, hc.UserID), nil
}

func (hc Backend) run() {

	tlsClientConfig := &tls.Config{}

	if hc.Insecure {
		tlsClientConfig = Insecure(tlsClientConfig)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsClientConfig,
		},
	}

	docs := make([]json.Marshaler, 0)

	send := func(docs []json.Marshaler) {
		if len(docs) == 0 {
			return
		}

		authHeader := ""
		if val, err := hc.MakeAuthHeader(); err == nil {
			authHeader = val
		} else {
			log.Errorf("Error creating DShield authentication header: %s", err.Error())
		}

		l := Submit{
			AuthHeader: authHeader,
			Type:       "multiple",
			Logs:       docs,
		}

		pr, pw := io.Pipe()

		hash := sha1.New()

		r := io.TeeReader(pr, hash)
		r = io.TeeReader(r, os.Stdout)

		go func(l Submit) {
			var err error

			defer pw.CloseWithError(err)

			if err := json.NewEncoder(pw).Encode(l); err != nil {
				log.Errorf("Error json encoding: %s", err.Error())
			}
		}(l)

		req, err := http.NewRequest(http.MethodPost, "https://www.dshield.org/submitapi/", r)
		if err != nil {
			log.Errorf("Could create new request: %s", err.Error())
			return
		}

		req.Header.Set("User-Agent", fmt.Sprintf("Honeytrap/%s (%s; %s) %s", cmd.Version, runtime.GOOS, runtime.GOARCH, cmd.ShortCommitID))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("Could not submit event to DShield: %s", err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Errorf("Could not submit event to DShield: %d", resp.StatusCode)
			return
		}

		if !hc.Debug {
		} else if val, err := httputil.DumpResponse(resp, true); err == nil {
			log.Debug(string(val))
		}

		// verify hash!
		fmt.Printf("%x\n", hash.Sum(nil))
	}

	for {
		select {
		case doc := <-hc.ch:
			docs = append(docs, doc)

			if len(docs) < 10 {
				continue
			}

			send(docs)

			docs = make([]json.Marshaler, 0)
		case <-time.After(time.Second * 2):
			send(docs)

			docs = make([]json.Marshaler, 0)
		}
	}
}

func (hc Backend) send(msg json.Marshaler) {
	select {
	case hc.ch <- msg:
	default:
		log.Errorf("Could not send more messages, channel full")
	}
}

// Send delivers the giving push messages into dshield endpoint.
func (hc Backend) Send(message event.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Error sending event: %+v", r)
		}
	}()

	switch message.Get("category") {
	case "telnet":
		if message.Get("type") != "password-authentication" {
		} else {
			evt := &SSHEvent{}

			message.Range(func(key, val interface{}) bool {
				switch key.(string) {
				case "date":
					evt.Date = val.(time.Time)
				case "source-ip":
					evt.SourceIP = val.(string)
				case "destination-ip":
					evt.DestinationIP = val.(string)
				case "source-port":
					evt.SourcePort = val.(int)
				case "destination-port":
					evt.DestinationPort = val.(int)
				case "telnet.username":
					evt.Username = val.(string)
				case "telnet.password":
					evt.Password = val.(string)
				}

				return true
			})

			hc.send(evt)
		}
	case "ssh":
		if message.Get("type") != "password-authentication" {
		} else {
			evt := &SSHEvent{}

			message.Range(func(key, val interface{}) bool {
				switch key.(string) {
				case "date":
					evt.Date = val.(time.Time)
				case "source-ip":
					evt.SourceIP = val.(string)
				case "destination-ip":
					evt.DestinationIP = val.(string)
				case "source-port":
					evt.SourcePort = val.(int)
				case "destination-port":
					evt.DestinationPort = val.(int)
				case "ssh.username":
					evt.Username = val.(string)
				case "ssh.password":
					evt.Password = val.(string)
				}

				return true
			})

			hc.send(evt)
		}
	case "http":
		// not yet supported
	default:
	}
}
