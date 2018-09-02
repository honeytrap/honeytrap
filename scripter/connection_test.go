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
	"testing"
	"reflect"
)

//TestConnectionStruct_GetScrConn tests the get scripter connection function on a connection wrapper
func TestConnectionStruct_GetScrConn(t *testing.T) {
	got := connectionWrapper.GetScrConn()

	expected := scrConn

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "GetScrConn", got, expected)
	}
}

//TestConnectionStruct_GetParameters tests the get parameter function on a connection wrapper
func TestConnectionStruct_GetParameters(t *testing.T) {
	got, err := connectionWrapper.GetParameters([]string { "key", "value" })
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string(nil)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test %s failed: got %+#v, expected %+#v", "GetScrConn", got, expected)
	}
}

//TestConnectionStruct_Handle tests the handle function on a connection wrapper
func TestConnectionStruct_Handle(t *testing.T) {
	connectionWrapper.Handle("test")
}

//TestConnectionStruct_SetFloatFunction tests the float function on a connection wrapper
func TestConnectionStruct_SetFloatFunction(t *testing.T) {
	connectionWrapper.SetFloatFunction("getFloatFunctionTest", func() float64 {
		return 0
	})
}

//TestConnectionStruct_SetStringFunction tests the string function on a connection wrapper
func TestConnectionStruct_SetStringFunction(t *testing.T) {
	connectionWrapper.SetStringFunction("getStringFunctionTest", func() string {
		return ""
	})
}

//TestConnectionStruct_SetVoidFunction tests the void function on a connection wrapper
func TestConnectionStruct_SetVoidFunction(t *testing.T) {
	connectionWrapper.SetVoidFunction("getVoidFunctionTest", func() {

	})
}
