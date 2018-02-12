package ftp

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestBreakPath(t *testing.T) {
	cases := []struct {
		in, dname, fname string
	}{
		{"/path/to/file", "/path/to", "file"},
		{"/path/to/file/", "/path/to/file", ""},
		{"/myfile", "/", "myfile"},
		{"/", "/", ""},
		{"", "", ""},
	}

	for _, c := range cases {
		d, f := breakpath(c.in)
		if d != c.dname && f != c.fname {
			t.Errorf("breakpath(%s) dir is [%s], want [%s] file is [%s], want [%s]",
				c.in, d, c.dname, f, c.fname)
		}
	}
}

func Testmakefile(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	name := "testing"
	fs.makefile(name)
	if fs.f["/"][0].Name() != name {
		t.Error("makefile: file not in filesystem")
	}
	log.Debugf("fs: %v", fs)
}

func TestStat(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	fs.makefile("myfile")

	cases := []struct {
		in, want string
	}{
		{"/myfile", "myfile"},
		{"/", "/"},
		{"", ""},
		{"/Not/A/Valid/Path", ""},
	}

	for _, c := range cases {
		finfo, err := fs.Stat(c.in)

		if finfo == nil && err == nil {
			t.Error("Stat returns nil but not an error")
		}
		if err == nil && finfo.Name() != c.want {
			t.Errorf("Stat(%s) want %s, got %s", c.in, c.want, finfo.Name())
		}
	}
}

func TestMakeDir(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	cases := []struct {
		in, want string
	}{
		{"/newdir", "/newdir"},
		{"/mydir/mynewdir", ""},
		{"/newdir/testing/", "/newdir/testing"},
		{"/not/a/valid/parent", ""},
		{"", ""},
	}

	for _, c := range cases {
		err := fs.MakeDir(c.in)
		if _, ok := fs.f[c.want]; !ok && err == nil {
			t.Errorf("MakeDir(%s) Error", c.in)
		}
	}

}

func TestCurDir(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	want := "/"
	if got := fs.CurDir(); got != want {
		t.Errorf("CurDir() got %s, want %s", got, want)
	}
}

func TestChangeDir(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	cases := []struct {
		in    string
		valid bool
	}{
		{"/", true},
		{"/mydir/", false},
		{"", false},
		{"/not/valid", false},
		{"not/valid", false},
	}

	for _, c := range cases {
		err := fs.ChangeDir(c.in)

		if err == nil && !c.valid {
			t.Errorf("ChangeDir(%s) No error but is not valid", c.in)
		}

		if err != nil && c.valid {
			t.Errorf("ChangeDir(%s) is valid but gives Error", c.in)
		}
	}
}

func TestListDir(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	// /
	// |- test1
	// |- test2
	// |- [testdir]
	//        |- test3

	fs.makefile("test1")
	fs.makefile("test2")
	if err := fs.MakeDir("/testdir"); err != nil {
		t.Error("ListDir: Could not set up test filesystem")
	}
	fs.ChangeDir("/testdir")
	fs.makefile("test3")

	finfo := fs.ListDir("/")

	if len(finfo) != 3 {
		t.Errorf("ListDir: wrong number of elements: %v", finfo)
	}

	for _, e := range finfo {
		switch e.Name() {
		case "test1":
		case "test2":
		case "test3":
		case "testdir":
		default:
			t.Errorf("ListDir: Wrong listing: %v", e.Name())
		}
	}

	finfo = fs.ListDir("/testdir")

	if len(finfo) != 1 {
		t.Errorf("ListDir: wrong number of elements: %v", finfo)
	}

	if finfo[0].Name() != "test3" {
		t.Errorf("ListDir: Wrong listing: %v", finfo)
	}
}

func TestDeleteDir(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	if err := fs.MakeDir("/testing"); err != nil {
		t.Error("DeleteDir: Could not set up test filesystem")
	}

	cases := []struct {
		in    string
		valid bool
	}{
		{"/testing", true},
		{"/not/valid", false},
		{"", false},
		{"/", true},
	}

	for _, c := range cases {
		err := fs.DeleteDir(c.in)

		if c.valid {
			if _, ok := fs.f[c.in]; err == nil && ok {
				t.Errorf("DeleteDir(%s) Not deleted but throws no error", c.in)
			}
		} else {
			if err == nil {
				t.Errorf("DeleteDir(%s) Bad path but no error", c.in)
			}
		}
	}
}

func TestDeleteFile(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	delfile := "/test"
	fs.makefile(delfile)

	if err := fs.DeleteFile(delfile); err != nil {
		t.Errorf("DeleteFile(%s) gives Error. None expected: %s", delfile, err.Error())
	}

	if _, err := fs.Stat(delfile); err == nil {
		t.Errorf("DeleteFile(%s) Not deleted!", delfile)
	}

	if err := fs.DeleteFile("Not/Valid"); err == nil {
		t.Error("DeleteFile(Not/Valid) Expected an error, but none given")
	}
}

/*
func TestRename(t *testing.T) {
	fs := &Dummyfs{
		f:   dfs,
		dir: "/",
	}

	from := "/myfile"
	to := "/changed"

	fs.makefile(from)

	if err := fs.Rename(from, to); err != nil {
		t.Errorf("Rename(%s, %s) gives Error: %s", from, to, err.Error())
	}

	if _, err := fs.Stat(to); err != nil {
		t.Errorf("%s is not in filesystem: %s", to, err.Error())
	}
}
*/

func TestGetFile(t *testing.T) {
	fs := &Dummyfs{
		f:        map[string][]*FileInfo{"/": []*FileInfo{}},
		dir:      "/",
		download: []byte("very secret stuff!"),
	}

	sendfile := "/secrets.txt"

	fs.makefile(sendfile)

	if filesz, newfile, _ := fs.GetFile(sendfile, int64(0)); newfile != nil {
		var buf bytes.Buffer

		if newsz, err := io.Copy(&buf, newfile); err != nil {
			t.Errorf("GetFile read error: %s", err.Error())
		} else if newsz != filesz {
			t.Errorf("GetFile filesize: %d is not read size: %d", filesz, newsz)
		}

		newfile.Close()
	} else {
		t.Error("GetFile returned a nil reader")
	}
}

func TestPutFile(t *testing.T) {
	fs := &Dummyfs{
		f:   map[string][]*FileInfo{"/": []*FileInfo{}},
		dir: "/",
	}

	putfile := "/uplo.ad"
	if _, err := fs.Stat(putfile); err == nil {
		t.Errorf("Wrong test file: %s already exists", putfile)
	}

	//Create a dummy io.Reader
	upload := (*os.File)(nil)

	if _, err := fs.PutFile(putfile, upload, false); err != nil {
		t.Errorf("PutFile put error: %s", err.Error())
	}

	if _, err := fs.Stat(putfile); err != nil {
		t.Errorf("PutFile %s not in filesystem", putfile)
	}
}
