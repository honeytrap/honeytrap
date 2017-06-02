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
