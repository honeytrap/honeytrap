package message

import "fmt"

//====================================================================================

// PushMessage defines a struct which contains specific data relating to
// different messages to provide Push notifications for the pusher api.
type PushMessage struct {
	Sensor      string
	Category    string
	SessionID   string
	ContainerID string
	Data        interface{}
}

//====================================================================================

// EventType defines a int type for all event types available.
type EventType int

// contains different sets of possible events type.
const (
	// Process based events.
	ProcessBegin = iota + 1
	ProcessEnd

	Ping

	NewConnection
	ConnectionStarted
	ConnectionClosed
	ConnectionRequest
	ConnectionResponse
	ConnectionError

	ServiceStarted
	ServiceEnded

	// Container based events.
	ContainerClone
	ContainerStarted
	ContainerFrozen
	ContainerUnfrozen
	ContainerStopped
	ContainerTarBackup
	ContainerDataPacket
	ContainerDataCheckpoint

	// SSH based events.
	SSHSessionBegin
	SSHSessionEnd

	// Authentication events.
	Login = iota + 30
	Logout
)

// Event defines a struct which contains definitive details about the operation of
// a giving event.
type Event struct {
	Sensor      string                 `json:"sensor"`
	Category    string                 `json:"category"`
	Type        EventType              `json:"event_type"`
	Data        interface{}            `json:"data"`
	Details     map[string]interface{} `json:"details"`
	SessionID   string                 `json:"session_id,omitempty"`
	ContainerID string                 `json:"container_id,omitempty"`
}

// String returns a stringified version of the event.
func (e Event) String() string {
	return fmt.Sprintf("Event %d occured with for Sensor[%q] in Category[%q]. Data[%#q] - Detail[%#q]", e.Type, e.Sensor, e.Category, e.Data, e.Details)
}

//====================================================================================
