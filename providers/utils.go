package providers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

type FileCloser struct {
	*os.File
	path string
}

func (f *FileCloser) Close() error {
	ec := f.File.Close()
	log.Info("Will Remove %s", f.path)
	ex := os.Remove(f.path)

	if ex == nil {
		return ec
	}

	return ex
}

func NewFileCloser(path string) (*FileCloser, error) {
	ff, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return &FileCloser{ff, path}, nil
}

//BufferCloser closes a byte.Buffer
type BufferCloser struct {
	*bytes.Buffer
}

//NewBufferCloser returns a new closer for a bytes.Buffer
func NewBufferCloser(bu *bytes.Buffer) *BufferCloser {
	return &BufferCloser{bu}
}

//Close resets the internal buffer
func (b *BufferCloser) Close() error {
	b.Buffer.Reset()
	return nil
}

//GzipWalker walks a path and turns it into a tar written into a bytes.Buffer
func GzipWalker(file string, tmp io.Writer) error {
	f, err := os.Open(file)

	if err != nil {
		return err
	}

	defer f.Close()

	//gzipper
	gz := gzip.NewWriter(tmp)
	defer gz.Close()

	log.Info("Will Copy data from file to gzip")
	_, err = io.Copy(gz, f)

	return err
}

//TarWalker walks a path and turns it into a tar written into a bytes.Buffer
func TarWalker(rootpath string, w io.Writer) error {
	tz := tar.NewWriter(w)
	defer tz.Close()

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("Error while walking path %s: ", path, err.Error())
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		np, err := filepath.Rel(rootpath, path)
		if err != nil {
			return err
		}

		fl, err := os.Open(path)
		if err != nil {
			return err
		}

		defer fl.Close()

		var h *tar.Header
		if h, err = tar.FileInfoHeader(info, ""); err != nil {
			return err
		}

		h.Name = np

		if err := tz.WriteHeader(h); err != nil {
			return err
		}

		if _, err := io.Copy(tz, fl); err != nil {
			return err
		}

		// TODO: should be pushed to channel, then in channel it can be filtered, tarred or whatever. Just push the path
		return nil
	}

	err := filepath.Walk(rootpath, walkFn)
	if err != nil {
		log.Error("Error occured walking dir %s with Error: (%+s)", rootpath, err.Error())
		return err
	}

	return nil
}
