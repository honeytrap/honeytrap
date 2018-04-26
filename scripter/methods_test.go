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
	"reflect"
	"testing"
	"time"
)

//TestGetRemoteAddr tests the retrieval of the remote address from a connection
func TestGetRemoteAddr(t *testing.T) {
	got := getRemoteAddr(connectionWrapper.Conn)()

	expected := "pipe"

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "getRemoteAddr", got, expected)
	}
}

//TestGetLocalAddr tests the retrieval of the local address from a connection
func TestGetLocalAddr(t *testing.T) {
	got := getLocalAddr(connectionWrapper.Conn)()

	expected := "pipe"

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "getRemoteAddr", got, expected)
	}
}

//TestGetDatetime tests the datetime retrieval in unix timestamp
func TestGetDatetime(t *testing.T) {
	ct := time.Now()

	got := getDatetime()()

	expected := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
		ct.Year(), ct.Month(), ct.Day(),
		ct.Hour(), ct.Minute(), ct.Second())

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "getDatetime", got, expected)
	}
}

//TestGetFileDownload tests the file download functionality from a connection
func TestGetFileDownload(t *testing.T) {
	got := getFileDownload(connectionWrapper.Conn, "test")()

	expected := "no"

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "getFileDownload", got, expected)
	}
}

//TestChannelSend tests the channel send functionality to be used on a connection
func TestChannelSend(t *testing.T) {
	c, err := pushers.Dummy()
	if err != nil {
		t.Fatal(err)
	}

	dummy, err := Dummy("test", WithChannel(c))
	if err != nil {
		t.Fatal(err)
	}

	channelSend(dummy, connectionWrapper.Conn, "test")()
}

//TestDoLog tests the logging functionality on a connection
func TestDoLog(t *testing.T) {
	// logTypes := []string { "critical", "debug", "error", "info", "notice", "warning", "fatal", "panic" }

	doLog(connectionWrapper.Conn, "test")()
}

//TestGetFolder tests the retrieval of the scripter folder of a connection
func TestGetFolder(t *testing.T) {
	dummy, err := Dummy("test")
	if err != nil {
		t.Fatal(err)
	}

	got := getFolder(dummy)()

	expected := "test"

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "getFolder", got, expected)
	}
}

//TestSetBasicMethods tests the setting of the basic methods in Lua with Go
func TestSetBasicMethods(t *testing.T) {
	dummy, err := Dummy("test")
	if err != nil {
		t.Fatal(err)
	}

	SetBasicMethods(dummy, connectionWrapper.Conn, "test")
}
