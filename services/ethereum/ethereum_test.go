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
package ethereum

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
	Req      string
	Expected string
}

var tests = []Test{
	Test{
		Name: "eth_getBlockByNumber",
		Req:  `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false], "id":583367}`,
		Expected: `{
"id":583367,
"jsonrpc":"2.0",
"result": {
    "number": "0x1b4",
    "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
    "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
    "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
    "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
    "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
    "difficulty": "0x027f07",
    "totalDifficulty":  "0x027f07",
    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "size":  "0x027f07",
    "gasLimit": "0x9f759",
    "minGasPrice": "0x9f759",
    "gasUsed": "0x9f759",
    "timestamp": "0x54e34e8e",
    "transactions": [],
    "uncles": []
  }
}`,
	},
}

func TestEthereum(t *testing.T) {
	c := Ethereum()
	c.SetChannel(pushers.MustDummy())

	for _, tst := range tests {
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()

		go c.Handle(context.TODO(), server)

		req := httptest.NewRequest("POST", "/", strings.NewReader(tst.Req))
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
		if err := json.Unmarshal([]byte(tst.Expected), &expected); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("Test %s failed: got %+#v, expected %+#v", tst.Name, got, expected)
			return
		}
	}
}
