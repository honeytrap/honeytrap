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
package server

import (
	"testing"
)

func TestBigPortToAddr(t *testing.T) {
	addr, proto, port, err := ToAddr("tcp/60000")
	if err != nil {
		t.Fatal(err)
	}

	if addr.String() != ":60000" {
		t.Errorf("Expected :60000 but got %s", addr)
	}
	if proto != "tcp" {
		t.Errorf("Expected tcp but got %s", proto)
	}
	if port != 60000 {
		t.Errorf("Expected 60000 but got %d", port)
	}
}

func TestIncorrectSeparatorToAddr(t *testing.T) {
	_, _, _, err := ToAddr("tcp:8080")
	if err == nil {
		t.Errorf("No error thrown with incorrect separator")
	}
}


func TestUnknownProtoToAddr(t *testing.T) {
	_, _, _, err := ToAddr("tdp:8080")
	if err == nil {
		t.Errorf("No error thrown with incorrect protocol")
	}
}
