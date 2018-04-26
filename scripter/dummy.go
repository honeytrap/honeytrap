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

import (
	"bytes"
	"fmt"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/yuin/gopher-lua"
	"net"
	"time"
)

// New creates a lua scripter instance that handles the connection to all scripts
// A list where all scripts are stored in is generated
func Dummy(name string, options ...ScripterFunc) (Scripter, error) {
	l := &dummyScripter{
		name: name,
	}

	for _, optionFn := range options {
		optionFn(l)
	}

	return l, nil
}

// The scripter state to which scripter functions are attached
type dummyScripter struct {
	name string

	c pushers.Channel
}

// SetChannel sets the channel over which messages to the log and elasticsearch can be set
func (l *dummyScripter) SetChannel(c pushers.Channel) {
	l.c = c
}

// GetChannel gets the channel over which messages to the log and elasticsearch can be set
func (l *dummyScripter) GetChannel() pushers.Channel {
	return l.c
}

// Init initializes the scripts from a specific service
// The service name is given and the method will loop over all files in the scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *dummyScripter) Init(service string) error {
	return nil
}

//GetConnection returns a connection for the given ip-address, if no connection exists yet, create it.
func (l *dummyScripter) GetConnection(service string, conn net.Conn) ConnectionWrapper {
	return &ConnectionStruct{Service: service, Conn: nil}
}

// CanHandle checks whether scripter can handle incoming connection for the peeked message
// Returns true if there is one script able to handle the connection
func (l *dummyScripter) CanHandle(service string, message string) bool {
	return false
}

// GetScripts return the scripts for this scripter
func (l *dummyScripter) GetScripts() map[string]map[string]string {
	return nil
}

// GetScriptFolder return the folder where the scripts are located for this scripter
func (l *dummyScripter) GetScriptFolder() string {
	return fmt.Sprintf("%s", l.name)
}

// CleanConnections Check all connections removing all that haven't been used for more than 60 minutes to open up memory
func (l *dummyScripter) CleanConnections() {

}

type dummyConn struct {
	conn net.Conn

	//List of lua scripts running for this connection: directory/scriptname
	scripts map[string]map[string]*lua.LState

	connectionBuffer bytes.Buffer
}

//GetConn returns the connection for the SrcConn
func (c *dummyConn) GetConn() net.Conn {
	return c.conn
}

//SetStringFunction sets a function that is available in all scripts for a service
func (c *dummyConn) SetStringFunction(name string, getString func() string, service string) error {
	return nil
}

//SetFloatFunction sets a function that is available in all scripts for a service
func (c *dummyConn) SetFloatFunction(name string, getFloat func() float64, service string) error {
	return nil
}

//SetVoidFunction sets a function that is available in all scripts for a service
func (c *dummyConn) SetVoidFunction(name string, doVoid func(), service string) error {
	return nil
}

//GetParameters gets the stack parameters from lua to be used in Go functions
func (c *dummyConn) GetParameters(params []string, service string) (map[string]string, error) {
	return nil, nil
}

//HasScripts returns whether the scripts for a given service are loaded already
func (c *dummyConn) HasScripts(service string) bool {
	return false
}

//AddScripts adds scripts to a connection for a given service
func (c *dummyConn) AddScripts(service string, scripts map[string]string, folder string) error {
	return nil
}

// GetConnectionBuffer returns the buffer of the connection
func (c *dummyConn) GetConnectionBuffer() *bytes.Buffer {
	return nil
}

// Handle calls the handle method on the lua state with the message as the argument
func (c *dummyConn) Handle(service string, message string) (*Result, error) {
	return nil, nil
}

// GetLastUsed returns the time in milliseconds that this connection was called for the last time
func (c *dummyConn) GetLastUsed() time.Time {
	return time.Now()
}
