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
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"net"
	"os"
	"reflect"
	"testing"
)

var connectionWrapper *ConnectionStruct
var scrConn ScrConn

func TestMain(m *testing.M) {
	basepath = "../test-scripts/"

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	scrConn = &dummyConn{
		conn: server,
	}

	connectionWrapper = &ConnectionStruct{Conn: scrConn, Service: "test"}

	os.Exit(m.Run())
}

// TestRegister tests the register of a scripter
func TestRegister(t *testing.T) {
	Register("dummy", Dummy)

	scripterFunc, ok := Get("dummy")
	if !ok {
		t.Fatal(fmt.Errorf("unable to retrieve scripter function"))
	}

	if _, err := scripterFunc("dummy"); err != nil {
		t.Fatal(err)
	}
}

// TestGet tests the successful retrieval of a registered scripter
func TestGet(t *testing.T) {
	Register("dummy", Dummy)

	if _, ok := Get("dummy"); !ok {
		t.Fatal(fmt.Errorf("unable to retrieve scripter function"))
	}
}

// TestGet2 tests the unsuccessful retrieval of an unregistered scripter
func TestGet2(t *testing.T) {
	Register("dummy", Dummy)

	if _, ok := Get("dummy2"); ok {
		t.Fatal(fmt.Errorf("able to retrieve unavailable scripter function"))
	}
}

// TestGetAvailableScripterNames tests the list of available scripters by name
func TestGetAvailableScripterNames(t *testing.T) {
	Register("dummy", Dummy)

	expected := []string{
		"dummy",
	}
	got := GetAvailableScripterNames()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "TestGetAvailableScripterNames", got, expected)
	}
}

// TestWithChannel tests the channel setup of a scripter
func TestWithChannel(t *testing.T) {
	Register("dummy", Dummy)

	scripterFunc, ok := Get("dummy")
	if !ok {
		t.Fatal(fmt.Errorf("unable to retrieve scripter function"))
	}

	expected, _ := pushers.Dummy()

	scripter, err := scripterFunc("test", WithChannel(expected))
	if err != nil {
		t.Fatal(err)
	}

	got := scripter.GetChannel()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "TestWithChannel", got, expected)
	}
}

// TestWithConfig test the config setup of a scripter
func TestWithConfig(t *testing.T) {
	if _, err := Dummy("dummy", WithConfig(toml.Primitive{})); err != nil {
		t.Fatal(err)
	}
}

// TestReloadScripts test the reload scripts of a scripter
func TestReloadScripts(t *testing.T) {
	dummy, err := Dummy("dummy")
	if err != nil {
		t.Fatal(err)
	}

	if err := ReloadScripts(dummy); err != nil {
		t.Fatal(err)
	}
}

// TestReloadAllScripters test the overall reload scripts of scripters
func TestReloadAllScripters(t *testing.T) {
	dummy, err := Dummy("dummy")
	if err != nil {
		t.Fatal(err)
	}

	scripters := map[string]Scripter{
		"dummy": dummy,
	}
	if err := ReloadAllScripters(scripters); err != nil {
		t.Fatal(err)
	}
}
