package ftp

import (
	"errors"
	"io"
	"strings"
	"time"
)

type FileInfo struct {
	mode  string
	owner string
	group string
	size  int
	name  string
	mtime time.Time
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

func (f *FileInfo) ModTime() time.Time {
	return f.mtime
}

type DummyFS struct {
	f map[string][]FileInfo

	dir string //Current directory
}

func (d *DummyFS) Init(conn *Conn) {
	d.f = map[string][]FileInfo{
		"/": []FileInfo{
			FileInfo{
				"drwxr-xr-x",
				"user",
				"users",
				0,
				"mydir",
				time.Date(2018, time.January, 20, 9, 0, 0, 0, time.UTC),
			},
			FileInfo{
				"-rwxrwxrwx",
				"user",
				"users",
				1024,
				"myfile",
				time.Date(2018, time.January, 19, 11, 0, 0, 0, time.UTC),
			},
		},
		"/mydir": []FileInfo{
			FileInfo{
				"-rwxrwxrwx",
				"user",
				"users",
				5623,
				"diary.txt",
				time.Now(),
			},
			FileInfo{
				"-rwxrwxrwx",
				"user",
				"users",
				2812,
				"passwords",
				time.Date(2018, time.January, 11, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	d.dir = "/"
}

//return directory and filename
func breakpath(path string) (dir string, filename string) {
	if path == "" {
		return "", ""
	}

	split := strings.LastIndex(path, "/")

	if split < 0 {
		//No dir separators, treat this as a filename
		return "", path
	}

	filename = path[split+1:]
	dir = path[:split]

	return
}

func (d *DummyFS) Stat(path string) (*FileInfo, error) {
	dir, fname := breakpath(path)

	if fname == "" {
		return nil, errors.New("No filename")
	}

	if l, ok := d.f[dir]; ok {
		if len(fname) > 0 {
			for _, lname := range l {
				if fname == lname.Name() {
					return &lname, nil
				}
			}
		}
	}
	return nil, errors.New("Not a valid path: " + path)
}

func (d *DummyFS) ChangeDir(path string) error {
	if _, ok := d.f[path]; ok {
		d.dir = path
		return nil
	}
	return errors.New("Wrong path: " + path)
}

//Return a single file or all files in a directory
func (d *DummyFS) ListDir(path string) []FileInfo {

	if path == "" {
		return d.f[d.dir]
	}

	dir, fname := breakpath(path)

	if l, ok := d.f[dir]; ok {
		if fname == "" {
			return l
		} else {
			for _, lname := range l {
				if fname == lname.Name() {
					return []FileInfo{lname}
				}
			}
		}
	}
	return nil
}

func (d *DummyFS) DeleteDir(path string) error {
	if _, ok := d.f[path]; ok {
		delete(d.f, path)
		return nil
	}
	return errors.New("Not a valid path: " + path)
}

func (d *DummyFS) DeleteFile(path string) error {
	return nil
}

func (d *DummyFS) Rename(from_path, to_path string) error {
	return nil
}

func (d *DummyFS) MakeDir(path string) error {
	dpath := strings.TrimRight(path, "/")
	dir, name := breakpath(dpath)

	if dir == "" {
		dir = d.dir
	}

	if name == "" {
		return errors.New("No directory name given")
	}

	if parent, ok := d.f[dir]; ok {
		parent = append(parent, FileInfo{"drwxr-xr-x", "user", "users", 0, name, time.Now()})
		d.f[dpath] = []FileInfo{}
	} else {
		return errors.New("Not a valid parent directory")
	}

	return nil
}

func (d *DummyFS) CurDir() string {
	return d.dir
}

func (d *DummyFS) GetFile(path string, n int64) (int64, io.ReadCloser, error) {
	return 0, nil, nil
}

func (d *DummyFS) PutFile(path string, r io.Reader, b bool) (int64, error) {
	return 0, nil
}
