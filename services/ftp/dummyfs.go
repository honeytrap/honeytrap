package ftp

import (
	"io"
	"time"
)

var thisYear, _, _ = time.Now().Date()

var lsList = []FileInfo{
	// UNIX ls -l style
	{"drwxr-xr-x", "110", "1002", 0, "mydir", EntryTypeFolder, time.Date(2009, time.December, 2, 0, 0, 0, 0, time.UTC)},
	{"-rwxr-xr-x", "110", "1002", 1234, "myfile", EntryTypeFile, time.Date(2009, time.December, 2, 0, 0, 0, 0, time.UTC)},
	{"lrwxrwxrwx", "root", "root", 0, "bin -> usr/bin", EntryTypeLink, time.Date(thisYear, time.January, 25, 0, 17, 0, 0, time.UTC)},
}

/*
type lines struct {
	line      string
	name      string
	size      uint64
	entryType EntryType
	time      time.Time
}
*/

type FileInfo struct {
	mode      string
	owner     string
	group     string
	size      int
	name      string
	entryType EntryType
	time      time.Time
}

func (f *FileInfo) Owner() string {
	return f.owner
}

func (f *FileInfo) Group() string {
	return f.group
}

func (f *FileInfo) Mode() string {
	return f.mode
}

func (f *FileInfo) Size() int {
	return f.size
}

func (f *FileInfo) Name() string {
	return f.name
}

type DummyFS struct {
	fs []FileInfo
}

func (d *DummyFS) Init(conn *Conn) {
	d.fs = lsList
}

func (d *DummyFS) Stat(path string) (*FileInfo, error) {
	return &d.fs[0], nil
}

func (d *DummyFS) ChangeDir(path string) error {
	return nil
}

func (d *DummyFS) ListDir(path string, fn func(FileInfo) error) error {
	return nil
}

func (d *DummyFS) DeleteDir(path string) error {
	return nil
}

func (d *DummyFS) DeleteFile(path string) error {
	return nil
}

func (d *DummyFS) Rename(from_path, to_path string) error {
	return nil
}

func (d *DummyFS) MakeDir(path string) error {
	return nil
}

func (d *DummyFS) GetFile(path string, n int64) (int64, io.ReadCloser, error) {
	return 0, nil, nil
}

func (d *DummyFS) PutFile(path string, r io.Reader, b bool) (int64, error) {
	return 0, nil
}
