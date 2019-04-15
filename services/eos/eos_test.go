// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
