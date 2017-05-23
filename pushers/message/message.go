package message

import (
	"fmt"
	"time"
)

//====================================================================================

// PushMessage defines a struct which contains specific data relating to
// different messages to provide Push notifications for the pusher api.
type PushMessage struct {
	Event       bool
	Sensor      string
	Category    string
	SessionID   string
	ContainerID string
	Data        interface{}
}

//====================================================================================

// EventType defines a string type for all event types available.
type EventType string

// contains different sets of possible events type.
const (
	PingEvent            EventType = "PING"
	DataRequest          EventType = "DATA:REQUEST"
	DataRead             EventType = "DATA:READ"
	DataWrite            EventType = "DATA:WRITE"
	ServiceEnded         EventType = "SERVICE:ENDED"
	ServiceStarted       EventType = "SERVICE:STARTED"
	ConnectionOpened     EventType = "CONNECTION:OPENED"
	ConnectionClosed     EventType = "CONNECTION:CLOSED"
	UserSessionOpened    EventType = "SESSION:USER:OPENED"
	UserSessionClosed    EventType = "SESSION:USER:CLOSED"
	ConnectionReadError  EventType = "CONNECTION:ERROR:READ"
	ConnectionWriteError EventType = "CONNECTION:ERROR:WRITE"
	ContainerStarted     EventType = "CONTAINER:STARTED"
	ContainerFrozen      EventType = "CONTAINER:FROZEN"
	ContainerDial        EventType = "CONTAINER:DIAL"
	ContainerUnfrozen    EventType = "CONTAINER:UNFROZEN"
	ContainerCloned      EventType = "CONTAINER:CLONED"
	ContainerStopped     EventType = "CONTAINER:STOPPED"
	ContainerPaused      EventType = "CONTAINER:PAUSED"
	ContainerResumed     EventType = "CONTAINER:RESUMED"
	ContainerTarred      EventType = "CONTAINER:TARRED"
	ContainerCheckpoint  EventType = "CONTAINER:CHECKPOINT"
	ContainerPcaped      EventType = "CONTAINER:PCAPED"
)

// Contains a series of sensors constants.
const (
	ContainersSensor      = "CONTAINER"
	ConnectionSensor      = "CONNECTION"
	ServiceSensor         = "SERVICE"
	SessionSensor         = "SESSIONS"
	EventSensor           = "EVENTS"
	PingSensor            = "PING"
	DataSensor            = "DATA"
	ErrorsSensor          = "ERRORS"
	DataErrorSensor       = "DATA:ERROR"
	ConnectionErrorSensor = "CONNECTION:ERROR"
)

// Event defines a struct which contains definitive details about the operation of
// a giving event.
type Event struct {
	Date        time.Time              `json:"date"`
	Data        interface{}            `json:"data"`
	Sensor      string                 `json:"sensor"`
	Details     map[string]interface{} `json:"details"`
	HostAddr    string                 `json:"host_addr"`
	LocalAddr   string                 `json:"local_addr"`
	Type        EventType              `json:"event_type"`
	Ended       time.Time              `json:"ended,omitempty"`
	Token       string                 `json:"token,omitempty"`
	Started     time.Time              `json:"started,omitempty"`
	Location    string                 `json:"location,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ContainerID string                 `json:"container_id,omitempty"`
}

// String returns a stringified version of the event.
func (e Event) String() string {
	return fmt.Sprintf("Event %q occured with for Sensor[%q], Data[%#q] - Detail[%#q]", e.Type, e.Sensor, e.Data, e.Details)
}

//====================================================================================
