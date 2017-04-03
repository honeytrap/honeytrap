package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/honeytrap/honeytrap/pushers/message"
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

// Get returns the giving buckets based on the provided cursor point and size.
// If the `from` and `length` are -1 then all keys and values are returned, else
// the provided range will be used.
func (d *Bolted) Get(bucket []byte, from int, length int, fx func([]message.Event)) error {
	return d.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(bucket)
		cu := bu.Cursor()

		var list []message.Event

		// Retrieve all values in bucket.
		if from < 0 && length < 0 {
			for k, v := cu.First(); k != nil; k, v = cu.Next() {

				// Probably some subbucket.
				if v == nil {
					continue
				}

				var item message.Event
				if err := json.Unmarshal(v, &item); err != nil {
					return err
				}

				list = append(list, item)
			}

			// Call the pending callback with event slice.
			fx(list)

			return nil
		}

		if length < 0 {
			for k, v := cu.Seek(parseInt(uint64(from))); k != nil; k, v = cu.Next() {

				// Probably some subbucket.
				if v == nil {
					continue
				}

				var item message.Event
				if err := json.Unmarshal(v, &item); err != nil {
					return err
				}

				list = append(list, item)
			}

			// Call the pending callback with event slice.
			fx(list)

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

			var item message.Event
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}

			list = append(list, item)

			counter++
		}

		// Call the pending callback with event slice.
		fx(list)

		return nil
	})
}

// Save attempts to save the series of passed in events into the underline db.
func (d *Bolted) Save(bucket []byte, events ...message.Event) error {
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
