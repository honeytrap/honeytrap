package director

import "github.com/honeytrap/honeytrap/pushers/message"

// ContainerStoppedEvent returns a connection open event object giving the associated data values.
func ContainerStoppedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerStopped,
	}
}

// ContainerDialEvent returns a connection open event object giving the associated data values.
func ContainerDialEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerDial,
	}
}

// ContainerPcappedEvent returns a connection open event object giving the associated data values.
func ContainerPcappedEvent(c Container, data []byte, detail map[string]interface{}) message.Event {
	if detail == nil {
		detail = map[string]interface{}{}
	}

	detail["container"] = c.Detail()

	return message.Event{
		Details: detail,
		Data:    data,
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerPcaped,
	}
}

// ContainerCheckpointEvent returns a connection open event object giving the associated data values.
func ContainerCheckpointEvent(c Container, data []byte, detail map[string]interface{}) message.Event {
	if detail == nil {
		detail = map[string]interface{}{}
	}

	detail["container"] = c.Detail()

	return message.Event{
		Details: detail,
		Data:    data,
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerCheckpoint,
	}
}

// ContainerTarredEvent returns a connection open event object giving the associated data values.
func ContainerTarredEvent(c Container, data []byte, detail map[string]interface{}) message.Event {
	if detail == nil {
		detail = map[string]interface{}{}
	}

	detail["container"] = c.Detail()

	return message.Event{
		Details: detail,
		Data:    data,
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerTarred,
	}
}

// ContainerClonedEvent returns a connection open event object giving the associated data values.
func ContainerClonedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerPaused,
	}
}

// ContainerPausedEvent returns a connection open event object giving the associated data values.
func ContainerPausedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerPaused,
	}
}

// ContainerResumedEvent returns a connection open event object giving the associated data values.
func ContainerResumedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerResumed,
	}
}

// ContainerUnfrozenEvent returns a connection open event object giving the associated data values.
func ContainerUnfrozenEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerUnfrozen,
	}
}

// ContainerFrozenEvent returns a connection open event object giving the associated data values.
func ContainerFrozenEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerFrozen,
	}
}

// ContainerStartedEvent returns a connection open event object giving the associated data values.
func ContainerStartedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerStarted,
	}
}

// ContainerErrorEvent returns a connection open event object giving the associated data values.
func ContainerErrorEvent(c Container, data error) message.Event {
	return message.Event{
		Data:   data,
		Sensor: message.ContainersSensor,
		Type:   message.PingEvent,
	}
}

// ContainerDataWriteEvent returns a connection open event object giving the associated data values.
func ContainerDataWriteEvent(c Container, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Data:    data,
		Details: meta,
		Sensor:  message.ContainersSensor,
		Type:    message.DataWrite,
	}
}

// ContainerDataReadEvent returns a connection open event object giving the associated data values.
func ContainerDataReadEvent(c Container, data interface{}) message.Event {
	return message.Event{
		Data:   data,
		Sensor: message.ContainersSensor,
		Type:   message.DataRead,
	}
}