package director

import "github.com/honeytrap/honeytrap/pushers/message"

// ContainerStoppedEvent returns a connection open event object giving the associated data values.
func ContainerStoppedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Message: "Container has been stopped",
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerStopped,
	}
}

// ContainerDialEvent returns a connection open event object giving the associated data values.
func ContainerDialEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Message: "New Container net.Conn created",
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
		Message: "Container network data stored",
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
		Message: "Container is getting checkpoint to save history",
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
		Message: "Container data is getting tarred",
	}
}

// ContainerClonedEvent returns a connection open event object giving the associated data values.
func ContainerClonedEvent(c Container, name string, template string, ip string) message.Event {
	return message.Event{
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerPaused,
		Message: "Container is cloned",
		Details: map[string]interface{}{
			"container-ip":       ip,
			"container-name":     name,
			"container-template": template,
		},
	}
}

// ContainerPausedEvent returns a connection open event object giving the associated data values.
func ContainerPausedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerPaused,
		Message: "Container has just being paused",
	}
}

// ContainerResumedEvent returns a connection open event object giving the associated data values.
func ContainerResumedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerResumed,
		Message: "Container has just being resumed",
	}
}

// ContainerUnfrozenEvent returns a connection open event object giving the associated data values.
func ContainerUnfrozenEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerUnfrozen,
		Message: "Container has just being unfrozen",
	}
}

// ContainerFrozenEvent returns a connection open event object giving the associated data values.
func ContainerFrozenEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerFrozen,
		Message: "Container has just being frozen",
	}
}

// ContainerStartedEvent returns a connection open event object giving the associated data values.
func ContainerStartedEvent(c Container, detail map[string]interface{}) message.Event {
	return message.Event{
		Details: detail,
		Data:    c.Detail(),
		Sensor:  message.ContainersSensor,
		Type:    message.ContainerStarted,
		Message: "Container has just being started",
	}
}

// ContainerErrorEvent returns a connection open event object giving the associated data values.
func ContainerErrorEvent(c Container, data error) message.Event {
	return message.Event{
		Data:    data,
		Sensor:  message.ErrorsSensor,
		Type:    message.ContainerError,
		Message: "Container has just faced an error",
	}
}

// ContainerDataWriteEvent returns a connection open event object giving the associated data values.
func ContainerDataWriteEvent(c Container, data interface{}, meta map[string]interface{}) message.Event {
	return message.Event{
		Data:    data,
		Details: meta,
		Sensor:  message.ContainersSensor,
		Type:    message.DataWrite,
		Message: "Container has just written new data",
	}
}

// ContainerDataReadEvent returns a connection open event object giving the associated data values.
func ContainerDataReadEvent(c Container, data interface{}) message.Event {
	return message.Event{
		Data:    data,
		Sensor:  message.ContainersSensor,
		Type:    message.DataRead,
		Message: "Container has just read more data",
	}
}
