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
package ldap

import "testing"

func TestIsLogin(t *testing.T) {
	s := Server{
		login: "tester",
	}

	got := s.isLogin()

	if got != true {
		t.Errorf("isLogin (login name: %s) Got %v, Want true", s.login, got)
	}

	s.login = ""
	got = s.isLogin()

	if got != false {
		t.Errorf("isLogin (login name: '') Got %v, Want false", got)
	}
}

func TestDSEGet(t *testing.T) {
	str := "HT"

	d := &DSE{
		VendorName: []string{str},
	}

	got := d.Get()

	if name, ok := got.Attrs["vendorName"]; ok {
		if name[0] != str {
			t.Errorf("*DSE.Get Got %s, Want %s", name, str)
		}
	} else {
		t.Errorf("*DSE.Get Key: vendorName not in AttributeMap")
	}
}
