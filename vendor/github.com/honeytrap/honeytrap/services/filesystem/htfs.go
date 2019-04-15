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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// HTFS - Honeytrap filesystem:

// Sandboxed filesystem.

type Htfs struct {
	root string //absolute path on host
	cwd  string //current working directory relative to root

	//dataSize   int64 //virtual size left in filesystem (Bytes)
	//filesCount int //count of files in the filesystem
}

//Return path is always in filesystem.
func (f *Htfs) RealPath(path string) string {

	//relative path, prefix with cwd. Prevents escaping the filesystem too
	var abspath string

	if !filepath.IsAbs(path) {
		abspath = filepath.Join(f.cwd, path)
	} else {
		abspath = filepath.Clean(path)
	}

	return filepath.Join(f.root, abspath)
}

func (f *Htfs) Cwd() string {

	return f.cwd
}

func (f *Htfs) ChangeDir(path string) error {

	rpath := f.RealPath(path)

	d, err := os.Lstat(rpath)
	if err != nil {
		return err
	}

	if d.IsDir() {
		rel, err := filepath.Rel(f.root, rpath)
		if err != nil {
			return err
		}
		f.cwd = filepath.Join(string(filepath.Separator), rel)
		return nil
	}

	return fmt.Errorf("Not a directory: %s", path)
}

func New(base, serviceName, serviceRoot string) (*Htfs, error) {

	if serviceName == "" { //We need this
		return nil, errors.New("New: No service name")
	}

	if broot := filepath.Clean(base); broot == "." { //no path given, use honeytrap starting dir
		ht, err := os.Executable()
		if err != nil {
			return nil, err
		}
		base = filepath.Dir(ht)
	}

	var root string

	if serviceRoot == "" {
		newroot, err := makeRoot(filepath.Join(base, serviceName))
		if err != nil {
			return nil, err
		}
		root = newroot
	} else {
		root = filepath.Join(base, serviceName, serviceRoot)
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, fmt.Errorf("Bad root path: %s", root)
	}

	return &Htfs{
		root: root,
		cwd:  string(filepath.Separator),
	}, nil
}

func makeRoot(base string) (string, error) {
	u, err := genUniqueName()
	if err != nil {
		return "", err
	}

	realpath := filepath.Join(base, u)

	if err := os.MkdirAll(realpath, 0700); err != nil {
		return "", err
	}

	return realpath, nil
}

func genUniqueName() (string, error) {
	hash := sha256.New()
	if _, err := io.CopyN(hash, rand.Reader, 50); err != nil {
		return "", err
	}

	uniq := hex.EncodeToString(hash.Sum(nil))

	return uniq[:15], nil
}
