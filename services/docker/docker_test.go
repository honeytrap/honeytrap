package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
)

type Test struct {
}

var tests = []struct {
	Name             string
	ReqMethod        string
	ReqPath          string
	ReqBody          string
	ExpectedJSONKeys []string
	ExpectedStatus   int
}{

	{
		Name:             "root_version",
		ReqMethod:        "GET",
		ReqPath:          "/version",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{"Platform", "Components"},
	},
	{
		Name:             "version",
		ReqMethod:        "GET",
		ReqPath:          "/1.41/version",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{"Platform", "Components"},
	},
	{
		Name:             "root_info",
		ReqMethod:        "GET",
		ReqPath:          "/info",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{"ID", "Containers"},
	},
	{
		Name:             "info",
		ReqMethod:        "GET",
		ReqPath:          "/1.41/info",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{"ID", "Containers"},
	},
	{
		Name:             "containers",
		ReqMethod:        "GET",
		ReqPath:          "/1.41/containers/json",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{},
	},
	{
		Name:             "containers_create",
		ReqMethod:        "POST",
		ReqPath:          "/1.41/containers/create",
		ReqBody:          "",
		ExpectedStatus:   201,
		ExpectedJSONKeys: []string{"Id", "Warnings"},
	},
	{
		Name:             "containers_kill",
		ReqMethod:        "POST",
		ReqPath:          "/1.41/containers/e90e34656806/kill",
		ReqBody:          "",
		ExpectedStatus:   204,
		ExpectedJSONKeys: []string{},
	},
	{
		Name:             "containers_start",
		ReqMethod:        "POST",
		ReqPath:          "/1.41/containers/e90e34656806/start",
		ReqBody:          "",
		ExpectedStatus:   204,
		ExpectedJSONKeys: []string{},
	},
	{
		Name:             "images",
		ReqMethod:        "GET",
		ReqPath:          "/1.41/images/json",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{},
	},

	{
		Name:             "images_create",
		ReqMethod:        "POST",
		ReqPath:          "/1.41/images/create?fromImage=ubuntu&tag=latest",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{},
	},
	{
		Name:             "images_no_tag",
		ReqMethod:        "POST",
		ReqPath:          "/1.24/images/create?fromImage=ubuntu",
		ReqBody:          "",
		ExpectedStatus:   200,
		ExpectedJSONKeys: []string{},
	},
}

func TestDocker(t *testing.T) {
	//Create Servicer
	s := Docker()

	// Create channel
	s.SetChannel(pushers.MustDummy())

	for _, tst := range tests {

		t.Run(tst.Name, func(t *testing.T) {

			//Create a pipe
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()

			go s.Handle(context.TODO(), server)

			req := httptest.NewRequest(tst.ReqMethod, tst.ReqPath, strings.NewReader(tst.ReqBody))
			if err := req.Write(client); err != nil {
				t.Error(err)
			}

			rdr := bufio.NewReader(client)

			resp, err := http.ReadResponse(rdr, req)
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode != tst.ExpectedStatus {
				t.Fatalf("Test %s failed: got status %+#v, expected status %+#v", tst.Name, resp.StatusCode, tst.ExpectedStatus)
			}

			//getting stuck here
			body, _ := ioutil.ReadAll(resp.Body)

			if len(tst.ExpectedJSONKeys) > 0 {
				got := map[string]interface{}{}
				if err := json.Unmarshal(body, &got); err != nil {
					t.Error(err)
				}

				for _, key := range tst.ExpectedJSONKeys {
					if _, ok := got[key]; !ok {
						t.Fatalf("Test %s failed: did not find key %+#v", tst.Name, key)
					}
				}
			}
		})
	}

}
