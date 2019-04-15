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
package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRealPath(t *testing.T) {
	h := &Htfs{
		root: "/tmp/test",
		cwd:  "/",
	}

	cases := []struct {
		in, want string
	}{
		{"", "/tmp/test"},
		{"a/", "/tmp/test/a"},
		{"a/b", "/tmp/test/a/b"},
		{"..", "/tmp/test"},
		{".", "/tmp/test"},
		{"../../new", "/tmp/test/new"},
	}

	for _, c := range cases {
		got := h.RealPath(c.in)
		if got != c.want {
			t.Errorf("TestRealPath: got %s, want %s", got, c.want)
		}
	}
}

func TestNew(t *testing.T) {

	want := "/tmp/test/sroot"

	if err := os.MkdirAll(want, 0777); err != nil {
		t.Fatalf("Can not create test directories: %s", want)
	}

	got, err := New("/tmp", "test", "sroot")
	if err != nil {
		t.Error(err.Error())
	}

	if got.root != want {
		t.Errorf("New set wrong directory, got %s want %s", got.root, want)
	}

	got, err = New("/tmp", "test", "")
	if err != nil {
		t.Error(err.Error())
	}

	gotbase := filepath.Join(got.root, "..")
	wantbase := filepath.Join(want, "..")

	if wantbase != gotbase {
		t.Errorf("New set wrong directory, got %s want %s", gotbase, wantbase)
	}

	_ = os.RemoveAll(want)
	_ = os.RemoveAll(got.root)
}

func TestChangeDir(t *testing.T) {
	want := "/tmp/test/sroot"

	if err := os.MkdirAll(want, 0777); err != nil {
		t.Fatalf("Can not create test directories: %s", want)
	}

	h, err := New("/tmp", "test", "sroot")
	if err != nil {
		t.Fatal(err)
	}

	dirs := filepath.Join(h.root, "a/b/c")

	err = os.MkdirAll(dirs, 0777)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		in, want string
	}{
		//Order is important here!
		{"a", "/a"},
		{"/a/b/c", "/a/b/c"},
		{"..", "/a/b"},
		{"", "/a/b"},
		{".", "/a/b"},
		{"/..", "/"},
	}

	for _, c := range cases {
		err = h.ChangeDir(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if c.want != h.Cwd() {
			t.Errorf("TestChangeDir want %s got %s", c.want, h.Cwd())
		}
	}

	//Test on error behaviour
	err = h.ChangeDir("/etc")

	if err == nil {
		t.Error("TestChangeDir returns no error on wrong input")
	}

	if h.Cwd() == "/etc" {
		t.Error("TestChangeDir: cwd outside filesystem")
	}
	_ = os.RemoveAll(dirs)
}
