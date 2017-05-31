package message

import (
	"fmt"
	"io"
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

// EventSensor defines a string type for all event sensors available.
type EventSensor string

// Contains a series of sensors constants.
const (
	ContainersSensor      EventSensor = "CONTAINER"
	ConnectionSensor      EventSensor = "CONNECTION"
	ServiceSensor         EventSensor = "SERVICE"
	SessionSensor         EventSensor = "SESSIONS"
	BasicSensor           EventSensor = "EVENTS"
	PingSensor            EventSensor = "PING"
	DataSensor            EventSensor = "DATA"
	ErrorsSensor          EventSensor = "ERRORS"
	DataErrorSensor       EventSensor = "DATA:ERROR"
	ConnectionErrorSensor EventSensor = "CONNECTION:ERROR"
)

// EventCategory defines a string type for for which is used to defined event category
// for different types.
type EventCategory string

// Event defines an interface which holds expect data received from event.
type Event interface {
	Message() string
	DataReader() io.Reader
	Fields() map[string]interface{}
	Identity() (EventCategory, EventType, EventSensor)
}

// BasicEvent defines a struct which contains definitive details about the operation of
// a giving event.
type BasicEvent struct {
	Date        time.Time              `json:"date"`
	Data        interface{}            `json:"data"`
	Category    EventCategory          `json:"category"`
	Sensor      EventSensor            `json:"sensor"`
	Details     map[string]interface{} `json:"details"`
	HostAddr    string                 `json:"hostAddr"`
	LocalAddr   string                 `json:"localAddr"`
	Type        EventType              `json:"event_type"`
	Ended       time.Time              `json:"ended,omitempty"`
	Token       string                 `json:"token,omitempty"`
	Started     time.Time              `json:"started,omitempty"`
	Location    string                 `json:"location,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ContainerID string                 `json:"container_id,omitempty"`
	Reader      io.Reader              `json:"reader"`
}

// DataReader returns a data reader if available for reading data from
// the giving event if available.
func (be BasicEvent) DataReader() io.Reader {
	return be.Reader
}

// Identity returns the Category, Type and Sensor associated with the event.
func (be BasicEvent) Identity() (EventCategory, EventType, EventSensor) {
	return be.Category, be.Type, be.Sensor
}

// Fields returns the event fields associated with the giving evevnt.
func (be BasicEvent) Fields() map[string]interface{} {
	return map[string]interface{}{
		"data":        be.Data,
		"type":        be.Type,
		"token":       be.Token,
		"sensor":      be.Sensor,
		"details":     be.Details,
		"hostAddr":    be.HostAddr,
		"location":    be.Location,
		"category":    be.Category,
		"localAddr":   be.LocalAddr,
		"sessionID":   be.SessionID,
		"containerID": be.ContainerID,
		"date":        be.Date.UTC().String(),
		"ended":       be.Ended.UTC().String(),
		"started":     be.Started.UTC().String(),
	}
}

// Message returns a the default Event message associated with the Event
func (be BasicEvent) Message() string {
	return fmt.Sprintf("Event occured with Sensor %q and Category %+q", be.Sensor, be.Category)
}

//====================================================================================

// EventSession is created to allow setting the sessionID of a event.
func EventSession(ev BasicEvent, sessionID string) BasicEvent {
	ev.SessionID = sessionID
	return ev
}

// EventContainer is created to allow setting the container of a event.
func EventContainer(ev BasicEvent, container string) BasicEvent {
	ev.ContainerID = container
	return ev
}

// EventLocation is created to allow setting the location of a event.
func EventLocation(ev BasicEvent, location string) BasicEvent {
	ev.Location = location
	return ev
}

// EventToken is created to allow setting the token of a event.
func EventToken(ev BasicEvent, token string) BasicEvent {
	ev.Token = token
	return ev
}

// EventCategoryType is created to allow setting the category of a event.
func EventCategoryType(ev BasicEvent, category string) BasicEvent {
	ev.Category = EventCategory(category)
	return ev
}

// EventDetail is created to allow setting the data of a event.
func EventDetail(ev BasicEvent, details map[string]interface{}) BasicEvent {
	ev.Details = details
	return ev
}

// EventData is created to allow setting the data of a event.
func EventData(ev BasicEvent, data interface{}) BasicEvent {
	ev.Data = data
	return ev
}
