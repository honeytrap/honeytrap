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
	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/scripter"
	"github.com/pkg/errors"
	"github.com/yuin/gopher-lua"
	"net"
	"os"
	"reflect"
	"testing"
)

var ls scripter.Scripter
var server net.Conn
var client net.Conn

type Config struct {
	Scripters map[string]toml.Primitive `toml:"scripter"`
}

// TestMain is the setup function for the global luaScripter
func TestMain(m *testing.M) {
	configString := "[scripter.lua]\r\n" +
		"type=\"lua\"\r\n" +
		"folder=\"../../test-scripts\"\r\n"

	configLua := &Config{}
	if _, err := toml.Decode(configString, configLua); err != nil {
		return
	}

	var err error
	ls, err = New("lua", scripter.WithConfig(configLua.Scripters["lua"]))
	if err != nil {
		log.Infof("%v", err)
		return
	}

	if err := ls.Init("test"); err != nil {
		return
	}

	server, client = net.Pipe()
	defer server.Close()
	defer client.Close()

	os.Exit(m.Run())
}

// TestNew tests the success of a new luaScripter without an error
func TestNew(t *testing.T) {
	if _, err := New("lua"); err != nil {
		t.Fatal(err)
	}
}

// TestNew2 tests the success of a new luaScripter with config without an error
func TestNew2(t *testing.T) {
	configString := "[scripter.lua]\r\n" +
		"type=\"lua\"\r\n" +
		"folder=\"../../test-scripts\"\r\n"

	configLua := &Config{}
	if _, err := toml.Decode(configString, configLua); err != nil {
		t.Error(err)
	}

	if _, err := New("lua", scripter.WithConfig(configLua.Scripters["lua"])); err != nil {
		t.Fatal(err)
	}
}

// TestLuaScripter_Init tests whether the init function does not return an error with scripts
func TestLuaScripter_Init(t *testing.T) {
	configString := "[scripter.lua]\r\n" +
		"type=\"lua\"\r\n" +
		"folder=\"../../test-scripts\"\r\n"

	configLua := &Config{}
	if _, err := toml.Decode(configString, configLua); err != nil {
		t.Error(err)
	}

	luaScripter, err := New("lua", scripter.WithConfig(configLua.Scripters["lua"]))
	if err != nil {
		t.Fatal(err)
	}

	if err := luaScripter.Init("test"); err != nil {
		t.Fatal(err)
	}
}

// TestLuaScripter_Init2 tests the error given when a luaScripter is made without config
func TestLuaScripter_Init2(t *testing.T) {
	luaScripter, err := New("lua")
	if err != nil {
		t.Fatal(err)
	}

	if err := luaScripter.Init("test"); err == nil {
		t.Fatal(errors.New("expected error while folder is not set in config"))
	}
}

// TestLuaScripter_CanHandle tests whether the CanHandle works with a message
func TestLuaScripter_CanHandle(t *testing.T) {
	//CanHandle the connection
	if ok := ls.CanHandle("test", "pass"); !ok {
		t.Fatal(errors.New("CanHandle failed to return success statement"))
	}
}

// TestLuaScripter_CanHandle2 tests whether the CanHandle works with a fail message
func TestLuaScripter_CanHandle2(t *testing.T) {
	//CanHandle the connection
	if ok := ls.CanHandle("test", "fail"); ok {
		t.Fatal(errors.New("CanHandle failed to return fail statement"))
	}
}

// TestLuaScripter_GetScriptFolder tests whether the right script folder is returned
func TestLuaScripter_GetScriptFolder(t *testing.T) {
	if !reflect.DeepEqual(ls.GetScriptFolder(), "../../test-scripts/lua") {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "login", ls.GetScriptFolder(), "../../test-scripts/lua")
	}
}

// TestLuaScripter_SetChannel tests the set channel function
func TestLuaScripter_SetChannel(t *testing.T) {
	c, err := pushers.Dummy()
	if err != nil {
		t.Fatal(err)
	}
	ls.SetChannel(c)
}

// TestLuaScripter_GetChannel tests the channel retrieve function from a luaScripter
func TestLuaScripter_GetChannel(t *testing.T) {
	c := ls.GetChannel()
	if _, ok := c.(pushers.Channel); !ok {
		t.Fatal(errors.New("invalid channel return"))
	}
}

// TestLuaScripter_GetScripts tests the retrieval of the scripts function
func TestLuaScripter_GetScripts(t *testing.T) {
	got := ls.GetScripts()

	expected := map[string]map[string]string{
		"test": {
			"test.lua": "../../test-scripts/lua/test/test.lua",
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "GetScripts", got, expected)
	}
}

// TestLuaScripter_GetConnection tests whether a new connection gets the right struct back
func TestLuaScripter_GetConnection(t *testing.T) {
	conn := ls.GetConnection("test", client)
	got := getConnIP(conn.GetScrConn().GetConn())

	expected := getConnIP(client)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "GetConnection", got, expected)
	}
}

// TestLuaScripter_GetConnection2 tests whether a existing connection gets the right struct back
func TestLuaScripter_GetConnection2(t *testing.T) {
	conn := ls.GetConnection("test", client)
	got := getConnIP(conn.GetScrConn().GetConn())

	expected := getConnIP(client)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "GetConnection", got, expected)
	}
}

// TestLuaConn_Handle tests the handle method on a connection
func TestLuaConn_Handle(t *testing.T) {
	conn := ls.GetConnection("test", client)

	result, err := conn.GetScrConn().Handle("test", "test")
	if err != nil {
		t.Fatal(err)
	}

	got := result.Content
	expected := "test"
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_Handle", got, expected)
	}
}

// TestLuaConn_callHandle tests if the connection can be handled by the script
func TestLuaConn_callHandle(t *testing.T) {
	luaState := lua.NewState()
	luaState.DoString("function handle(message) return message end")

	result, err := callHandle(luaState, "test")
	if err != nil {
		t.Fatal(err)
	}

	got := result.Content
	expected := "test"
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_callHandle", got, expected)
	}
}

// TestLuaConn_AddScripts tests whether a script can be added to a connection
func TestLuaConn_AddScripts(t *testing.T) {
	conn := ls.GetConnection("test", client)

	err := conn.GetScrConn().AddScripts("test", map[string]string{"test_other.lua": "../../test-scripts/lua/test_other/test_other.lua"}, "../../test-scripts")
	if err != nil {
		t.Fatal(err)
	}

	got := conn.GetScrConn().HasScripts("test")
	expected := true
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_AddScripts", got, expected)
	}
}

// TestLuaConn_SetStringFunction test the string functions from Go into Lua
func TestLuaConn_SetStringFunction(t *testing.T) {
	conn := ls.GetConnection("test", client)

	err := conn.GetScrConn().SetStringFunction("parameterTest", func() string {
		return "test"
	}, "test")

	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetScrConn().Handle("test", "test")
	if err != nil {
		t.Fatal(err)
	}

	got := result.Content
	expected := "testtest"
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_Handle", got, expected)
	}
}

// TestLuaConn_SetFloatFunction test the float functions from Go into Lua
func TestLuaConn_SetFloatFunction(t *testing.T) {
	conn := ls.GetConnection("test", client)

	err := conn.GetScrConn().SetFloatFunction("parameterTest", func() float64 {
		return 2
	}, "test")

	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetScrConn().Handle("test", "test")
	if err != nil {
		t.Fatal(err)
	}

	got := result.Content
	expected := "test2"
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_Handle", got, expected)
	}
}

// TestLuaConn_SetVoidFunction tests the void functions from Go in Lua
func TestLuaConn_SetVoidFunction(t *testing.T) {
	conn := ls.GetConnection("test", client)

	err := conn.GetScrConn().SetVoidFunction("parameterTest", func() {
		got, err := conn.GetScrConn().GetParameters([]string{"key", "value"}, "test")
		if err != nil {
			t.Fatal(err)
		}

		expected := map[string]string{"key": "key", "value": "value"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_GetParameters", got, expected)
		}
	}, "test")

	if err != nil {
		t.Fatal(err)
	}

	if _, err := conn.GetScrConn().Handle("test", "test"); err != nil {
		t.Fatal(err)
	}
}

// TestLuaConn_GetParameters test the parameter retrieval from Lua in Go
func TestLuaConn_GetParameters(t *testing.T) {
	conn := ls.GetConnection("test", client)

	err := conn.GetScrConn().SetStringFunction("parameterTest", func() string {
		got, err := conn.GetScrConn().GetParameters([]string{"key", "value"}, "test")
		if err != nil {
			t.Fatal(err)
		}

		expected := map[string]string{"key": "key", "value": "value"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Test %s failed: got %+#v, expected %+#v", "LuaConn_GetParameters", got, expected)
		}

		return "test"
	}, "test")

	_, err = conn.GetScrConn().Handle("test", "test")
	if err != nil {
		t.Fatal(err)
	}
}
