package ftp

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/honeytrap/honeytrap/services/filesystem"
)

type Fs struct {
	*filesystem.Htfs
}

func NewFileDriver(f *filesystem.Htfs) *Fs {
	return &Fs{f}
}

func (ftp *Fs) Init() {
}

func (ftp *Fs) Stat(path string) (os.FileInfo, error) {
	p := ftp.RealPath(path)

	info, err := os.Lstat(p)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ftp *Fs) ChangeDir(path string) error {

	return ftp.ChangeDir(path)
}

func (ftp *Fs) ListDir(path string) []os.FileInfo {
	p := ftp.RealPath(path)

	dir, err := os.Open(p)
	if err != nil {
		return []os.FileInfo{}
	}

	list, err := dir.Readdir(-1)
	if err != nil {
		return []os.FileInfo{}
	}

	return list
}

func (ftp *Fs) DeleteDir(path string) error {
	p := ftp.RealPath(path)

	info, err := os.Lstat(p)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("FtpFs: not a directory")
	}

	return os.Remove(p)
}

func (ftp *Fs) DeleteFile(path string) error {
	p := ftp.RealPath(path)
	return os.Remove(p)
}

func (ftp *Fs) Rename(from, to string) error {
	frompath := ftp.RealPath(from)
	topath := ftp.RealPath(to)

	return os.Rename(frompath, topath)
}

func (ftp *Fs) MakeDir(path string) error {
	p := ftp.RealPath(path)

	return os.Mkdir(p, 0770)
}

func (ftp *Fs) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
	p := ftp.RealPath(path)

	of, err := os.Open(p)
	if err != nil {
		return 0, nil, err
	}

	info, err := of.Stat()
	if err != nil {
		return 0, nil, err
	}

	of.Seek(offset, io.SeekEnd)

	return info.Size(), of, nil
}

func (ftp *Fs) PutFile(path string, data io.Reader, appendData bool) (int64, error) {
	p := ftp.RealPath(path)

	var isExist bool
	f, err := os.Lstat(p)
	if err == nil {
		isExist = true
		if f.IsDir() {
			return 0, errors.New("A dir has the same name")
		}
	} else {
		if os.IsNotExist(err) {
			isExist = false
		} else {
			return 0, errors.New(fmt.Sprintln("Put File error:", err))
		}
	}

	if appendData && !isExist {
		appendData = false
	}

	if !appendData {
		if isExist {
			err = os.Remove(p)
			if err != nil {
				return 0, err
			}
		}
		f, err := os.Create(p)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		bytes, err := io.Copy(f, data)
		if err != nil {
			return 0, err
		}
		return bytes, nil
	}

	of, err := os.OpenFile(p, os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return 0, err
	}
	defer of.Close()

	_, err = of.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	bytes, err := io.Copy(of, data)
	if err != nil {
		return 0, err
	}

	return bytes, nil
}

func (ftp *Fs) CurDir() string {
	return ftp.Cwd()
}
