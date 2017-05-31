package message

import (
	"fmt"
	"time"
)

//====================================================================================

// Messager defines an interface that exposes a single method
// that returns a custom message.
type Messager interface {
	Message() string
}

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
	Operational          EventType = "OPERATIONAL:Event"
	OperationalAuth      EventType = "OPERATIONAL:AUTH"
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
	ContainerError       EventType = "CONTAINER:ERROR"
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

// EventCategory defines a string type for for which is used to defined event category
// for different types.
type EventCategory string

// Event defines a struct which contains definitive details about the operation of
// a giving event.
type Event struct {
	Date        time.Time              `json:"date"`
	Data        interface{}            `json:"data"`
	Category    EventCategory          `json:"category"`
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

// String returns a the default Event message associated with the Event
func (e Event) String() string {
	return fmt.Sprintf("Event occured with Sensor %q and Category %+q", e.Sensor, e.Category)
}

//====================================================================================

// EventSession is created to allow setting the sessionID of a event.
func EventSession(ev Event, sessionID string) Event {
	ev.SessionID = sessionID
	return ev
}

// EventContainer is created to allow setting the container of a event.
func EventContainer(ev Event, container string) Event {
	ev.ContainerID = container
	return ev
}

// EventLocation is created to allow setting the location of a event.
func EventLocation(ev Event, location string) Event {
	ev.Location = location
	return ev
}

// EventToken is created to allow setting the token of a event.
func EventToken(ev Event, token string) Event {
	ev.Token = token
	return ev
}

// EventCategoryType is created to allow setting the category of a event.
func EventCategoryType(ev Event, category string) Event {
	ev.Category = EventCategory(category)
	return ev
}

// EventDetail is created to allow setting the data of a event.
func EventDetail(ev Event, details map[string]interface{}) Event {
	ev.Details = details
	return ev
}

// EventData is created to allow setting the data of a event.
func EventData(ev Event, data interface{}) Event {
	ev.Data = data
	return ev
}
