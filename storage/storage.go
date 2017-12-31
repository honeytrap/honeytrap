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
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/dgraph-io/badger"
)

type storage interface {
	Get(string) error
	Set(string, []byte) error
}

var db = MustDB()

func HomeDir() string {
	var err error
	var usr *user.User
	if usr, err = user.Current(); err != nil {
		panic(err)
	}

	p := path.Join(usr.HomeDir, ".honeytrap")

	_, err = os.Stat(p)

	switch {
	case err == nil:
		break
	case os.IsNotExist(err):
		if err = os.Mkdir(p, 0755); err != nil {
			panic(err)
		}
	default:
		panic(err)
	}

	return p
}

func MustDB() *badger.DB {
	opts := badger.DefaultOptions

	p := HomeDir()
	p = filepath.Join(p, "badger.db")

	opts.Dir = p
	opts.ValueDir = p

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return db
}

type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
}

func Namespace(namespace string) (*badgeStorage, error) {
	return &badgeStorage{
		db: db,
	}, nil
}

type badgeStorage struct {
	db *badger.DB
}

func (s *badgeStorage) Get(key string) ([]byte, error) {
	val := []byte{}

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
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
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), data)
		return err
	})
}
