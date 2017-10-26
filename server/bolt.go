// +build ignore

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
package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// Bolted defines a structure which saves delivered events into a giving boltDB
// database.
type Bolted struct {
	name string
	db   *bolt.DB
}

// NewBolted returns a new instance of a Bolted type.
func NewBolted(dbName string, buckets ...string) (*Bolted, error) {
	db, err := bolt.Open(fmt.Sprintf("%s.db", dbName), 0600, &bolt.Options{
		Timeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	var b Bolted
	b.name = dbName
	b.db = db

	// Create buckets for db.
	if terr := b.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}

		return nil
	}); terr != nil {
		return nil, terr
	}

	return &b, nil
}

// GetSize returns the giving size of the total items in a given bucket.
func (d *Bolted) GetSize(bucket []byte) (int, error) {
	var total int

	if terr := d.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		total = int(bu.Stats().KeyN)
		return nil
	}); terr != nil {
		return -1, terr
	}

	return total, nil
}

// Get returns the giving buckets based on the provided cursor point and size.
// If the `from` and `length` are -1 then all keys and values are returned, else
// the provided range will be used.
func (d *Bolted) Get(bucket []byte, from int, length int) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	// var total int

	if err := d.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		cu := bu.Cursor()

		// Retrieve all values in bucket.
		if from < 0 && length < 0 {
			for k, v := cu.First(); k != nil; k, v = cu.Next() {

				// Probably some subbucket.
				if v == nil {
					continue
				}

				var item map[string]interface{}
				if err := json.Unmarshal(v, &item); err != nil {
					return err
				}

				list = append(list, item)
			}

			return nil
		}

		if length < 0 {
			for k, v := cu.Seek(parseInt(uint64(from))); k != nil; k, v = cu.Next() {

				// Probably some subbucket.
				if v == nil {
					continue
				}

				var item map[string]interface{}
				if err := json.Unmarshal(v, &item); err != nil {
					return err
				}

				list = append(list, item)
			}

			return nil
		}

		var counter int

		for k, v := cu.Seek(parseInt(uint64(from))); k != nil; k, v = cu.Next() {
			// Probably some subbucket.
			if v == nil {
				continue
			}

			if counter >= length {
				break
			}

			var item map[string]interface{}
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}

			list = append(list, item)

			counter++
		}

		// Call the pending callback with event slice.

		return nil
	}); err != nil {
		return nil, err
	}

	return list, nil
}

// Save attempts to save the series of passed in events into the underline db.
func (d *Bolted) Save(bucket []byte, events ...map[string]interface{}) error {
	if events == nil {
		return nil
	}

	return d.db.Update(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)

		for _, event := range events {

			// TODO: Should we find a different encoding format for this?
			// Is this is Op expensive?
			buff, err := json.Marshal(event)
			if err != nil {
				return err
			}

			nextID, _ := bu.NextSequence()
			if terr := bu.Put(parseInt(nextID), buff); terr != nil {
				return terr
			}
		}

		return nil
	})
}

// Close closes the db and ends the session being used.
func (d *Bolted) Close() error {
	return d.db.Close()
}

//================================================================================

// parseInt returns a uint8 slice version of a given int value.
func parseInt(b uint64) []byte {
	bl := make([]byte, 8)
	binary.BigEndian.PutUint64(bl, b)
	return bl
}
