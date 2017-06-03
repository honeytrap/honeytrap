package director

import "github.com/honeytrap/honeytrap/pushers/event"

// ContainerStoppedEvent returns a connection open event object giving the associated data values.
func ContainerStoppedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerStopped,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerDialEvent returns a connection open event object giving the associated data values.
func ContainerDialEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerDial,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)

}

// ContainerPcappedEvent returns a connection open event object giving the associated data values.
func ContainerPcappedEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerPcaped,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerCheckpointEvent returns a connection open event object giving the associated data values.
func ContainerCheckpointEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerCheckpoint,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerTarredEvent returns a connection open event object giving the associated data values.
func ContainerTarredEvent(c Container, data []byte, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerTarred,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerClonedEvent returns a connection open event object giving the associated data values.
func ContainerClonedEvent(c Container, name string, template string, ip string) event.Event {
	return event.New(
		event.ContainerCloned,
		event.ContainersSensor,
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
		event.ContainerPaused,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerResumedEvent returns a connection open event object giving the associated data values.
func ContainerResumedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerResumed,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerUnfrozenEvent returns a connection open event object giving the associated data values.
func ContainerUnfrozenEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerUnfrozen,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerFrozenEvent returns a connection open event object giving the associated data values.
func ContainerFrozenEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerFrozen,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(detail),
	)
}

// ContainerStartedEvent returns a connection open event object giving the associated data values.
func ContainerStartedEvent(c Container, detail map[string]interface{}) event.Event {
	return event.New(
		event.ContainerStarted,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		// event.CopyFrom(detail),
	)
}

// ContainerErrorEvent returns a connection open event object giving the associated data values.
func ContainerErrorEvent(c Container, data error) event.Event {
	return event.New(
		event.ContainerError,
		event.ErrorsSensor,
		event.Custom("container", c.Detail()),
		// event.CopyFrom(detail),
	)
}

// ContainerDataWriteEvent returns a connection open event object giving the associated data values.
func ContainerDataWriteEvent(c Container, data []byte, meta map[string]interface{}) event.Event {
	return event.New(
		event.DataWrite,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.CopyFrom(meta),
		event.Payload(data),
	)
}

// ContainerDataReadEvent returns a connection open event object giving the associated data values.
func ContainerDataReadEvent(c Container, data []byte) event.Event {
	return event.New(
		event.DataRead,
		event.ContainersSensor,
		event.Custom("container", c.Detail()),
		event.Payload(data),
		// event.CopyFrom(detail),
	)
}
