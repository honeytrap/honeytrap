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
package lua

import (
	"fmt"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/op/go-logging"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"time"
)

var log = logging.MustGetLogger("scripter/lua")

var (
	_ = scripter.Register("lua", New)
)

// New creates a lua scripter instance that handles the connection to all scripts
// A list where all scripts are stored in is generated
func New(name string, options ...scripter.ScripterFunc) (scripter.Scripter, error) {
	l := &luaScripter{
		name: name,
	}

	for _, optionFn := range options {
		optionFn(l)
	}

	log.Infof("Using folder: %s", l.Folder)

	l.scripts = map[string]map[string]string{}
	l.connections = map[string]*luaConn{}
	l.canHandleStates = map[string]map[string]*lua.LState{}

	return l, nil
}

// The scripter state to which scripter functions are attached
type luaScripter struct {
	name string

	Folder       string        `toml:"folder"`
	CleanupTimer time.Duration `toml:"cleanupTimer"`

	//Source of the states, initialized per connection: directory/scriptname
	scripts map[string]map[string]string
	//List of connections keyed by 'ip'
	connections map[string]*luaConn
	//Lua states to check whether the connection can be handled with the script
	canHandleStates map[string]map[string]*lua.LState

	c pushers.Channel
}

// SetChannel sets the channel over which messages to the log and elasticsearch can be set
func (l *luaScripter) SetChannel(c pushers.Channel) {
	l.c = c
}

// GetChannel gets the channel over which messages to the log and elasticsearch can be set
func (l *luaScripter) GetChannel() pushers.Channel {
	return l.c
}

// Init initializes the scripts from a specific service
// The service name is given and the method will loop over all files in the scripts folder with the given service name
// All of these scripts are then loaded and stored in the scripts map
func (l *luaScripter) Init(service string) error {
	fileNames, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%s", l.Folder, l.name, service))
	if err != nil {
		return err
	}

	// TODO: Load basic lua functions from shared context
	l.connections = map[string]*luaConn{}
	l.scripts[service] = map[string]string{}
	l.canHandleStates[service] = map[string]*lua.LState{}

	for _, f := range fileNames {
		// Skip files that are not lua scripts
		if f.IsDir() || filepath.Ext(f.Name()) != ".lua" {
			continue
		}

		sf := fmt.Sprintf("%s/%s/%s/%s", l.Folder, l.name, service, f.Name())
		l.scripts[service][f.Name()] = sf

		ls := lua.NewState()

		// Allow importing without typing entire path
		ls.DoString(fmt.Sprintf("package.path = './%s/lua/?.lua;' .. package.path", l.Folder))
		if err := ls.DoFile(sf); err != nil {
			return err
		}
		l.canHandleStates[service][f.Name()] = ls
	}

	return nil
}

//GetConnection returns a connection for the given ip-address, if no connection exists yet, create it.
func (l *luaScripter) GetConnection(service string, conn net.Conn) scripter.ConnectionWrapper {
	ip := getConnIP(conn)

	sConn, ok := l.connections[ip]
	if !ok {
		sConn = &luaConn{
			conn:    conn,
			scripts: map[string]map[string]*lua.LState{},
		}
		l.connections[ip] = sConn
	} else {
		sConn.conn = conn
	}

	sConn.lastUsed = time.Now()

	if !sConn.HasScripts(service) {
		sConn.AddScripts(service, l.scripts[service], l.Folder)
		scripter.SetBasicMethods(l, sConn, service)
	}

	return &scripter.ConnectionStruct{Service: service, Conn: sConn}
}

// CanHandle checks whether scripter can handle incoming connection for the peeked message
// Returns true if there is one script able to handle the connection
func (l *luaScripter) CanHandle(service string, message string) bool {
	for _, ls := range l.canHandleStates[service] {
		canHandle, err := callCanHandle(ls, message)
		if err != nil {
			log.Errorf("%s", err)
		} else if canHandle {
			return true
		}
	}

	return false
}

// GetScripts return the scripts for this scripter
func (l *luaScripter) GetScripts() map[string]map[string]string {
	return l.scripts
}

// GetScriptFolder return the folder where the scripts are located for this scripter
func (l *luaScripter) GetScriptFolder() string {
	return fmt.Sprintf("%s/%s", l.Folder, l.name)
}

// CleanConnections Check all connections removing all that haven't been used for more than 60 minutes to open up memory
func (l *luaScripter) CleanConnections() {
	count := 0
	total := len(l.connections)

	if total == 0 {
		return
	}

	for key, connection := range l.connections {
		if time.Since(connection.GetLastUsed()) > l.CleanupTimer*time.Minute { //The connection hasn't been used for more than 60 minutes
			count++
			delete(l.connections, key)
		}
	}
	log.Infof("Cleaning connections, %d of %d connections were cleaned, %d remaining", count, total, total-count)
}

// getConnIP retrieves the IP from a connection's remote address
func getConnIP(conn net.Conn) string {
	s := strings.Split(conn.RemoteAddr().String(), ":")
	s = s[:len(s)-1]
	return strings.Join(s, ":")
}
