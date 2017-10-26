/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package director

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
