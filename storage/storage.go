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
package storage

import (
	"log"
	"path/filepath"

	"github.com/dgraph-io/badger"
)

type storage interface {
	Get(string) error
	Set(string, []byte) error
}

var db *badger.DB
var dataDir string

// SetDataDir
func SetDataDir(s string) {
	if db != nil {
		return
	}

	dataDir = s
	db = MustDB()
}

// MustDB
func MustDB() *badger.DB {
	opts := badger.DefaultOptions

	p := filepath.Join(dataDir, "badger.db")
	opts.Dir = p
	opts.ValueDir = p

	for _, fn := range PlatformOptions {
		fn(&opts)
	}

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

// Storage interface
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
}

// Namespace sets the namespace prefix
func Namespace(namespace string) (*badgeStorage, error) {
	prefix := make([]byte, len(namespace)+1)

	_ = copy(prefix, namespace)

	prefix[len(namespace)] = byte('.')

	return &badgeStorage{
		db: db,
		ns: prefix,
	}, nil
}

type badgeStorage struct {
	db *badger.DB

	ns []byte
}

func (s *badgeStorage) Get(key string) ([]byte, error) {
	val := []byte{}

	k := append(s.ns, key...)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}

		v, err := item.Value()
		if err != nil {
			return err
		}

		val = v
		return nil
	})

	return val, err
}

func (s *badgeStorage) Set(key string, data []byte) error {
	k := append(s.ns, key...)

	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(k, data)
		return err
	})
}

