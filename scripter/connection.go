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
package scripter

//ConnectionStruct
type ConnectionStruct struct {
	Service string
	Conn    ScrConn
}

// GetScrConn returns the ScrConn
func (w *ConnectionStruct) GetScrConn() ScrConn {
	return w.Conn
}

// Handle incoming message string
// Get all scripts for a given service and pass the string to each script
func (w *ConnectionStruct) Handle(message string) (string, error) {
	result, err := w.Conn.Handle(w.Service, message)

	if err != nil {
		log.Errorf("Error while handling scripts: %s", err)
	}

	if result != nil {
		return result.Content, nil
	}

	return "", nil
}

//SetStringFunction sets a string function for a connection
func (w *ConnectionStruct) SetStringFunction(name string, getString func() string) error {
	return w.Conn.SetStringFunction(name, getString, w.Service)
}

//SetFloatFunction sets a string function for a connection
func (w *ConnectionStruct) SetFloatFunction(name string, getFloat func() float64) error {
	return w.Conn.SetFloatFunction(name, getFloat, w.Service)
}

//SetVoidFunction sets a string function for a connection
func (w *ConnectionStruct) SetVoidFunction(name string, doVoid func()) error {
	return w.Conn.SetVoidFunction(name, doVoid, w.Service)
}

//GetParameters gets a parameter from a connection
func (w *ConnectionStruct) GetParameters(params []string) (map[string]string, error) {
	return w.Conn.GetParameters(params, w.Service)
}
