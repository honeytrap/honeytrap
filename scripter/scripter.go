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
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/abtester"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/op/go-logging"
	"net"
)

var (
	scripters = map[string]func(string, ...ScripterFunc) (Scripter, error){}
)
var log = logging.MustGetLogger("scripter")

//Register the scripter instance
func Register(key string, fn func(string, ...ScripterFunc) (Scripter, error)) func(string, ...ScripterFunc) (Scripter, error) {
	scripters[key] = fn
	return fn
}

type ScripterFunc func(Scripter) error

//Get a scripter instance
func Get(key string) (func(string, ...ScripterFunc) (Scripter, error), bool) {
	if fn, ok := scripters[key]; ok {
		return fn, true
	}

	return nil, false
}

//GetAvailableScripterNames gets all scripters that are registered
func GetAvailableScripterNames() []string {
	var out []string
	for key := range scripters {
		out = append(out, key)
	}
	return out
}

func WithChannel(eb pushers.Channel) ScripterFunc {
	return func(s Scripter) error {
		s.SetChannel(eb)
		return nil
	}
}

//Scripter interface that implements basic scripter methods
type Scripter interface {
	Init(string) error
	GetConnection(service string, conn net.Conn) ConnectionWrapper
	CanHandle(service string, message string) bool
	SetChannel(c pushers.Channel)
	GetChannel() pushers.Channel
	GetScripts() map[string]map[string]string
	GetScriptFolder() string
}

//ConnectionWrapper interface that implements the basic method that a connection should have
type ConnectionWrapper interface {
	GetScrConn() ScrConn
	Handle(message string) (string, error)
	SetStringFunction(name string, getString func() string) error
	SetFloatFunction(name string, getFloat func() float64) error
	SetVoidFunction(name string, doVoid func()) error
	GetParameters(params []string) (map[string]string, error)
}

//ScrConn wraps a connection and exposes methods to interact with the connection and scripter
type ScrConn interface {
	GetConn() net.Conn
	SetStringFunction(name string, getString func() string, service string) error
	SetFloatFunction(name string, getFloat func() float64, service string) error
	SetVoidFunction(name string, doVoid func(), service string) error
	GetParameters(params []string, service string) (map[string]string, error)
	HasScripts(service string) bool
	AddScripts(service string, scripts map[string]string, folder string)
	Handle(service string, message string) (*Result, error)
	GetConnectionBuffer() *bytes.Buffer
}

//Result struct which allows the result to be a string, an empty string and a nil value
//The nil value can be used to indicate that lua has no value to return
type Result struct {
	Content string
}

//ScrAbTester exposes methods to interact with the AbTester
type ScrAbTester interface {
	GetAbTester() abtester.Abtester
}

//WithConfig returns a function to attach the config to the scripter
func WithConfig(c toml.Primitive) ScripterFunc {
	return func(scr Scripter) error {
		return toml.PrimitiveDecode(c, scr)
	}
}

// ReloadScripts reloads the scripts from the scripter
func ReloadScripts(s Scripter) {
	for service := range s.GetScripts() {
		if err := s.Init(service); err != nil {
			log.Errorf("error init service: %s", err)
		} else {
			log.Infof("successfully updated service: %s", service)
		}
	}
}

// ReloadAllScripters reloads all scripts from scripters
func ReloadAllScripters(scripters map[string]Scripter) {
	for _, script := range scripters {
		ReloadScripts(script)
	}
}
