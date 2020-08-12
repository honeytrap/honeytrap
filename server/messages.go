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

package server

// MessageType defines a int type used to represent message type requests
// incoming.
type MessageType int

// contains possible set of expected Message requests and response types.
const (
	FetchSessions = iota + 1
	FetchSessionsReply

	FetchEvents
	FetchEventsReply

	NewSessions
	NewEvents

	ErrorResponse
)

// Message defines a generic message type send over the wire with a websocket
// request and response.
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// ErrorPayload defines a type which is delievered when an error occurs for a request
// or action which was not valid or failed.
type ErrorPayload struct {
	Request MessageType `json:"request"`
	Error   string      `json:"error"`
}
