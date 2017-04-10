package server

import (
	"bytes"
	"encoding/json"

	"github.com/gorilla/websocket"
)

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

// SocketTransport defines a struct which implements a message consumption and
// response transport for use over a websocket connection.
type SocketTransport struct{}

// HandleMessage defines a central method which provides the entry point which is used
// to respond to new messages.
func (so SocketTransport) HandleMessage(message []byte, conn *websocket.Conn) error {
	var newMessage Message

	if err := json.NewDecoder(bytes.NewBuffer(message)).Decode(&newMessage); err != nil {
		log.Errorf("Honeycast : Failed to decode message : %+q", err)
		return err
	}

	// We initially will only handle just two requests of getter types.
	// TODO: Handle NewSessions and NewEvents somewhere else.
	switch newMessage.Type {
	case FetchEvents:

	case FetchSessions:

	default:
		return so.DeliverMessage(Message{
			Type:    ErrorResponse,
			Payload: "Unknown Request Type",
		}, conn)
	}

	return nil
}

// DeliverMessage defines a method which handles the delivery of a message to a giving
// websocket.Conn.
func (so SocketTransport) DeliverMessage(message Message, conn *websocket.Conn) error {
	var bu bytes.Buffer

	if err := json.NewEncoder(&bu).Encode(message); err != nil {
		log.Errorf("Honeycast : Failed to decode message : %+q", err)
		return err
	}

	return conn.WriteMessage(websocket.BinaryMessage, bu.Bytes())
}
