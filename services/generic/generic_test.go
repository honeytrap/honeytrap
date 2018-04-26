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
package generic

import (
	"testing"
	"net"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/honeytrap/honeytrap/scripter"
	_ "github.com/honeytrap/honeytrap/scripter/lua"
	"github.com/honeytrap/honeytrap/services"
	"context"
	"github.com/honeytrap/honeytrap/pushers"
	"strings"
	"net/http/httptest"
	"bufio"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"reflect"
	"bytes"
)

type Config struct {
	Scripters map[string]toml.Primitive `toml:"scripter"`
}

// TestWithoutScript checks whether a wrong initialization of the generic service gives an error
func TestWithoutScript(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	s := Generic().(*genericService)

	//CanHandle the connection
	go func(conn net.Conn) {
		if err := s.Handle(nil, conn); err == nil {
			t.Fatal(errors.New("Expected missing scripter error"))
		}
	}(server)
}

// TestSimpleWrite checks whether a simple write from the client gives a successful response
func TestSimpleWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	configString := "[scripter.lua]\r\n" +
		"type=\"lua\"\r\n" +
		"folder=\"../../test-scripts\"\r\n"

	configLua := &Config{}
	if _, err := toml.Decode(configString, configLua); err != nil {
		t.Error(err)
	}

	scFunc, ok := scripter.Get("lua")
	if !ok {
		t.Errorf("failed to retrieve scripter func")
	}

	sc, err := scFunc("lua", scripter.WithConfig(configLua.Scripters["lua"]))
	if err != nil {
		t.Error(err)
	}

	s := Generic(services.WithScripter("generic", sc)).(*genericService)

	c, _ := pushers.Dummy()
	s.SetChannel(c)

	//CanHandle the connection
	go func(conn net.Conn) {
		if ok := s.CanHandle(nil); !ok {
			t.Fatal(errors.New("CanHandle failed to return statement"))
		}
	}(server)

	//Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(context.TODO(), server); err != nil {
			log.Errorf("%s", err)
			t.Fatal(err)
		}
	}(server)

	test := []byte("test")
	if _, err := client.Write(test); err != nil {
		t.Error(err)
	}

	buffer := make([]byte, 255)
	if n, err := client.Read(buffer[:]); err != nil {
		t.Error(err)
	} else {
		buffer = buffer[:n]
	}

	if !bytes.Equal(test, buffer) {
		t.Errorf("Test failed: got %+#v, expected %+#v", buffer, test)
		return
	}

	EOF := []byte("EOF")
	if _, err := client.Write(EOF); err != nil {
		t.Error(err)
	}
}

// TestHTTPRequest checks whether a HTTP request will response accordingly
func TestHTTPRequest(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	configString := "[scripter.lua]\r\n" +
		"type=\"lua\"\r\n" +
		"folder=\"../../test-scripts\"\r\n"

	configLua := &Config{}
	if _, err := toml.Decode(configString, configLua); err != nil {
		t.Error(err)
	}

	scFunc, ok := scripter.Get("lua")
	if !ok {
		t.Errorf("failed to retrieve scripter func")
	}

	sc, err := scFunc("lua", scripter.WithConfig(configLua.Scripters["lua"]))
	if err != nil {
		t.Error(err)
	}

	s := Generic(services.WithScripter("generic", sc)).(*genericService)

	c, _ := pushers.Dummy()
	s.SetChannel(c)

	//CanHandle the connection
	go func(conn net.Conn) {
		if ok := s.CanHandle(nil); !ok {
			t.Fatal(errors.New("CanHandle failed to return statement"))
		}
	}(server)

	//Handle the connection
	go func(conn net.Conn) {
		if err := s.Handle(context.TODO(), server); err != nil {
			log.Errorf("%s", err)
			t.Fatal(err)
		}
	}(server)

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"username":"test","password":"test"}`))
	if err := req.Write(client); err != nil {
		t.Error(err)
	}

	rdr := bufio.NewReader(client)

	resp, err := http.ReadResponse(rdr, req)
	if err != nil {
		t.Error(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	got := map[string]interface{}{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Error(err)
	}

	expected := map[string]interface{}{}
	if err := json.Unmarshal([]byte("{\"login\": \"success\"}"), &expected); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "login", got, expected)
		return
	}
}
