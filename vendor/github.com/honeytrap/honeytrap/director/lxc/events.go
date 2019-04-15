// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package lxc

import (
	"net"

	"github.com/honeytrap/honeytrap/event"
)

// ContainerClonedEvent returns a connection open event object giving the associated data values.
func ContainerClonedEvent(name, template string) event.Event {
	return event.New(
		event.ContainerCloned,
		event.ContainersSensor,
		event.Custom("container-name", name),
		event.Custom("container-template", template),
	)
}

// ContainerUnfrozenEvent returns a connection open event object giving the associated data values.
func ContainerUnfrozenEvent(name string, ip net.IP) event.Event {
	return event.New(
		event.ContainerUnfrozen,
		event.ContainersSensor,
		event.Custom("container-name", name),
		event.Custom("container-ip", ip.String()),
	)
}

// ContainerStartedEvent returns a connection open event object giving the associated data values.
func ContainerStartedEvent(name string) event.Event {
	return event.New(
		event.ContainerStarted,
		event.ContainersSensor,
		event.Custom("container-name", name),
	)
}

// ContainerErrorEvent returns a connection open event object giving the associated data values.
func ContainerErrorEvent(name string, e error) event.Event {
	return event.New(
		event.ContainerError,
		event.ContainersSensor,
		event.Custom("container-name", name),
		event.Error(e),
	)
}
