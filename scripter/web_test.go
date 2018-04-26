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
	"encoding/json"
)

//TestHandleRequests tests the handle requests general functionality
func TestHandleRequests(t *testing.T) {
	var request []byte
	dummy, err := Dummy("test")

	if err != nil {
		t.Fatal(err)
	}

	scripters := map[string]Scripter { "test": dummy }

	request, err = json.Marshal(map[string]interface{}{ "action": "script_reload" })
	if _, err = HandleRequests(scripters, request); err != nil {
		t.Fatal(err)
	}

	request, err = json.Marshal(map[string]interface{}{ "action": "script_read" })
	if _, err = HandleRequests(scripters, request); err != nil {
		t.Fatal(err)
	}

	request, err = json.Marshal(map[string]interface{}{ "action": "script_put", "path": "/lua/test/test_file.lua", "file": "test" })
	if _, err = HandleRequests(scripters, request); err != nil {
		t.Fatal(err)
	}

	request, err = json.Marshal(map[string]interface{}{ "action": "script_delete", "path": "/lua/test/test_file.lua", "file": "test" })
	if _, err = HandleRequests(scripters, request); err != nil {
		t.Fatal(err)
	}
}

//TestHandleScriptPut tests the script put
func TestHandleScriptPut(t *testing.T) {
	response, err := handleScriptPut(map[string]interface{}{ "path": "/lua/test/test_file.lua", "file": "test" })
	if err != nil {
		t.Fatal(err)
	}

	log.Infof("%v", response)
}

//TestHandleScriptRead tests the script reader
func TestHandleScriptRead(t *testing.T) {
	handleScriptRead(map[string]interface{}{ "dir": "" })
}

//TestHandleScriptDelete testss the script delete
func TestHandleScriptDelete(t *testing.T) {
	handleScriptDelete(map[string]interface{}{ "path": "/lua/test/test_file.lua" })
}

//TestHandleScriptReload tests the script reload
func TestHandleScriptReload(t *testing.T) {
	dummy, err := Dummy("test")
	if err != nil {
		t.Fatal(err)
	}

	scripters := map[string]Scripter { "test": dummy }

	handleScriptReload(scripters)
}
