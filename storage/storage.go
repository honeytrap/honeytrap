/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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
