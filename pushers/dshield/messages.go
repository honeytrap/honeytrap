// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
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
package dshield

import (
	"encoding/json"

	"time"
)

type Submit struct {
	AuthHeader string `json:"authheader"`

	Type string           `json:"type"`
	Logs []json.Marshaler `json:"logs"`
}

type HTTPEvent struct {
	Date time.Time

	SourceIP        string
	DestinationIP   string
	SourcePort      int
	DestinationPort int

	Method    string
	UserAgent string
	URL       string
}

func (e *HTTPEvent) MarshalJSON() ([]byte, error) {
	headers := []string{}

	val := map[string]interface{}{
		"type":      "httprequest",
		"time":      e.Date.Unix(),
		"sip":       e.SourceIP,
		"dip":       e.DestinationIP,
		"sport":     e.SourcePort,
		"dport":     e.DestinationPort,
		"headers":   headers,
		"method":    e.Method,
		"url":       e.URL,
		"useragent": e.UserAgent,
	}

	return json.Marshal(val)
}

type SSHEvent struct {
	Date time.Time

	SourceIP        string
	DestinationIP   string
	SourcePort      int
	DestinationPort int

	Username string
	Password string
}

func (e *SSHEvent) MarshalJSON() ([]byte, error) {
	val := map[string]interface{}{
		"type": "sshlogin",

		"time": e.Date.Unix(),

		"sip":   e.SourceIP,
		"dip":   e.DestinationIP,
		"sport": e.SourcePort,
		"dport": e.DestinationPort,

		"username": e.Username,
		"password": e.Password,
	}

	return json.Marshal(val)
}
