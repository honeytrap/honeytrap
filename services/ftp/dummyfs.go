package ftp

import "io"

type DummyFS struct {
}

type FileInfo struct {
	owner string
	group string
}

func (d *DummyFS) Init(conn *Conn) {
}

func (d *DummyFS) Stat(path string) (FileInfo, error) {
	return FileInfo{}, nil
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
