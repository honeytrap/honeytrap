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
package files

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("honeytrap:utils/files")

// FileCloser defines a struct which implements the io.Closer for a file object
// which removes the path when closed.
type FileCloser struct {
	*os.File
	path string
}

// Close calls the file.Close method and removes the file.
func (f *FileCloser) Close() error {
	ec := f.File.Close()
	log.Infof("Will Remove %s", f.path)
	ex := os.Remove(f.path)

	if ex == nil {
		return ec
	}

	return ex
}

// NewFileCloser returns a new instance of the FileCloser.
func NewFileCloser(path string) (*FileCloser, error) {
	ff, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return &FileCloser{ff, path}, nil
}

// BufferCloser closes a byte.Buffer
type BufferCloser struct {
	*bytes.Buffer
}

// NewBufferCloser returns a new closer for a bytes.Buffer
func NewBufferCloser(bu *bytes.Buffer) *BufferCloser {
	return &BufferCloser{bu}
}

//Close resets the internal buffer
func (b *BufferCloser) Close() error {
	b.Buffer.Reset()
	return nil
}

// GzipWalker walks a path and turns it into a tar written into a bytes.Buffer
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

// TarWalker walks a path and turns it into a tar written into a bytes.Buffer
func TarWalker(rootpath string, w io.Writer) error {
	tz := tar.NewWriter(w)
	defer tz.Close()

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Error while walking path %s: %s", path, err.Error())
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
		log.Errorf("Error occurred walking dir %s with Error: (%+s)", rootpath, err.Error())
		return err
	}

	return nil
}
