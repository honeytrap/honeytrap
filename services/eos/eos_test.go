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
package eos

import (
	"bufio"
	"context"
	"io/ioutil"
	"net"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"net/http"

	"encoding/json"

	"github.com/honeytrap/honeytrap/pushers"
)

type Test struct {
	Name     string
	Method   string
	Path     string
	Req      string
	Expected string
}

var tests = []Test{
	Test{
		Name:     "list_keys",
		Method:   "GET",
		Path:     "/v1/wallet/list_keys",
		Req:      ``,
		Expected: `[["EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV","5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"]]`,
	},
}

func TestEOS(t *testing.T) {
	c := EOS()
	c.SetChannel(pushers.MustDummy())

	for _, tst := range tests {
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()

		go c.Handle(context.TODO(), server)

		req := httptest.NewRequest(tst.Method, tst.Path, strings.NewReader(tst.Req))
		if err := req.Write(client); err != nil {
			t.Error(err)
		}

		rdr := bufio.NewReader(client)

		resp, err := http.ReadResponse(rdr, req)
		if err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)

		var got interface{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Error(err)
		}

		var expected interface{}
		if err := json.Unmarshal([]byte(tst.Expected), &expected); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Test %s failed: got %+#v, expected %+#v", tst.Name, got, expected)
			return
		}
	}
}
