package director

import "github.com/honeytrap/honeytrap/pushers/event"

// ContainerStoppedEvent returns a connection open event object giving the associated data values.
func ContainerStoppedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerStopped),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerDialEvent returns a connection open event object giving the associated data values.
func ContainerDialEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerDial),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)

}

// ContainerPcappedEvent returns a connection open event object giving the associated data values.
func ContainerPcappedEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerPcaped),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerCheckpointEvent returns a connection open event object giving the associated data values.
func ContainerCheckpointEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerCheckpoint),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerTarredEvent returns a connection open event object giving the associated data values.
func ContainerTarredEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerTarred),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerClonedEvent returns a connection open event object giving the associated data values.
func ContainerClonedEvent(c Container, name string, template string, ip string) event.Event {
	return event.New(
		event.Type(event.ContainerCloned),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(map[string]interface{}{
			"container-ip":       ip,
			"container-name":     name,
			"container-template": template,
		}),
	)
}

// ContainerPausedEvent returns a connection open event object giving the associated data values.
func ContainerPausedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerPaused),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerResumedEvent returns a connection open event object giving the associated data values.
func ContainerResumedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerResumed),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerUnfrozenEvent returns a connection open event object giving the associated data values.
func ContainerUnfrozenEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerUnfrozen),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerFrozenEvent returns a connection open event object giving the associated data values.
func ContainerFrozenEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerFrozen),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerStartedEvent returns a connection open event object giving the associated data values.
func ContainerStartedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.ContainerStarted),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		// event.CopyFrom(detail),
	)
}

// ContainerErrorEvent returns a connection open event object giving the associated data values.
func ContainerErrorEvent(c Container, data error) event.Event {
	return event.New(
		event.Type(event.ContainerError),
		event.Sensor(event.ErrorsSensor),
		event.Custom("container", c.Detail()),
		// event.CopyFrom(detail),
	)
}

// ContainerDataWriteEvent returns a connection open event object giving the associated data values.
func ContainerDataWriteEvent(c Container, data []byte, meta map[string]interface{}) event.Event {
	return event.New(
		event.Type(event.DataWrite),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.CopyFrom(meta),
		event.Payload(data),
	)
}

// ContainerDataReadEvent returns a connection open event object giving the associated data values.
func ContainerDataReadEvent(c Container, data []byte) event.Event {
	return event.New(
		event.Type(event.DataRead),
		event.Sensor(event.ContainersSensor),
		event.Custom("container", c.Detail()),
		event.Payload(data),
		// event.CopyFrom(detail),
	)
}
