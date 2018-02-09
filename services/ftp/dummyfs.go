package ftp

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
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

func (f *FileInfo) IsDir() bool {
	if f.mode[0] == 'd' {
		return true
	}

	return false
}

//return directory and filename
func breakpath(path string) (dir string, filename string) {
	if path == "" {
		return "", ""
	}

	split := strings.LastIndex(path, "/")

	if split < 0 {
		//No dir separators, treat this as a filename
		filename = path
		return
	}
	if split == len(path)-1 {
		//no filename
		dir = path[:split]
		return
	}

	filename = path[split+1:]
	if split == 0 { // dir is root
		dir = "/"
	} else {
		dir = path[:split]
	}

	return
}

type Dummyfs struct {
	f map[string][]*FileInfo

	download []byte

	dir string //Current directory
}

func NewDummyfs() *Dummyfs {
	fs := map[string][]*FileInfo{
		"/": []*FileInfo{},
	}

	return &Dummyfs{
		f:   fs,
		dir: "/",
	}
}

func (d *Dummyfs) Init() {
	d.download = []byte("very secret stuff!")
}

//Create a random datetime for use in FileInfo
func filetime() time.Time {
	t := time.Now()

	rand.Seed(int64(t.Nanosecond()))
	days := rand.Intn(30)
	seconds := days * 1000 * t.Nanosecond()

	return t.AddDate(0, -3, days).Add(time.Duration(seconds))
}

func (d *Dummyfs) makefile(path string) {
	dir, fname := breakpath(path)
	if dir == "" {
		dir = d.dir
	}

	newfile := &FileInfo{
		"-rwxr-xr-x",
		"user",
		"users",
		rand.Intn(1024 * 1024),
		fname,
		filetime(),
	}

	d.f[dir] = append(d.f[dir], newfile)
}

func (d *Dummyfs) Stat(path string) (*FileInfo, error) {
	if path == "" {
		return nil, errors.New("Stat: Empty path")
	}

	dpath := strings.TrimRight(path, "/")
	if dpath == "" { //We are in the root directory
		return &FileInfo{
			"drwxr-xr-x",
			"user",
			"users",
			0,
			"/",
			time.Now(),
		}, nil
	}

	dir, fname := breakpath(path)

	if dir == "" {
		dir = d.dir
	}

	if l, ok := d.f[dir]; ok {
		for _, f := range l {
			if fname == f.Name() {
				return f, nil
			}
		}
		return nil, errors.New("Stat: Bad Filename" + fname)
	}
	return nil, errors.New("Not a valid path: " + path)
}

func (d *Dummyfs) MakeDir(path string) error {
	dpath := strings.TrimRight(path, "/")
	parent, newdir := breakpath(dpath)

	//Put new directory under current directory
	if parent == "" {
		parent = d.dir
	}

	if newdir == "" {
		return errors.New("No new directory name given")
	}

	if _, ok := d.f[parent]; ok {
		d.f[parent] = append(d.f[parent], &FileInfo{"drwxr-xr-x", "user", "users", 0, newdir, filetime()})

		//Not in root '/' directory
		if len(parent) > 1 {
			newdir = "/" + newdir
		}

		d.f[parent+newdir] = []*FileInfo{}

		return nil
	}

	return errors.New("Not a valid path")
}

func (d *Dummyfs) CurDir() string {
	return d.dir
}

func (d *Dummyfs) ChangeDir(path string) error {
	if path == "" {
		return errors.New("Stat: Empty path")
	}

	if _, ok := d.f[path]; ok {
		d.dir = path
		return nil
	}
	return errors.New("Wrong path: " + path)
}

//Return all files in a directory
// nil on error
func (d *Dummyfs) ListDir(path string) []*FileInfo {

	if path == "" {
		//return current directory
		return d.f[d.dir]
	}

	stat, err := d.Stat(path)
	if err != nil {
		return nil
	}

	if stat.IsDir() {
		return d.f[path]
	}

	return nil
}

func (d *Dummyfs) DeleteDir(path string) error {
	if path == "/" { //can not delete root path
		return errors.New("DeleteDir: can not delete root '/'")
	}

	if _, ok := d.f[path]; ok {
		delete(d.f, path)
		return nil
	}
	return errors.New("Not a valid path: " + path)
}

func (d *Dummyfs) DeleteFile(path string) error {
	dir, fname := breakpath(path)

	if files, ok := d.f[dir]; ok {

		for i, f := range files {
			if fname == f.Name() {
				files[i] = files[len(files)-1]
				//files[len(files)-1] = nil
				files = files[:len(files)-1]
				d.f[dir] = files
				return nil
			}
		}
	}
	return errors.New("Could not delete file, wrong path")
}

func (d *Dummyfs) Rename(from_path, to_path string) error {
	return errors.New("Rename not implemented!")
}

// path: file to send
// n: offset to start sending from: NOT USED
// returns filesize, the file content and error
func (d *Dummyfs) GetFile(path string, n int64) (int64, io.ReadCloser, error) {
	if _, err := d.Stat(path); err != nil {
		return 0, nil, err
	}

	dl := ioutil.NopCloser(bytes.NewReader(d.download))
	return int64(len(d.download)), dl, nil
}

func (d *Dummyfs) PutFile(path string, r io.Reader, appendata bool) (int64, error) {
	d.makefile(path)

	return 0, nil
}
