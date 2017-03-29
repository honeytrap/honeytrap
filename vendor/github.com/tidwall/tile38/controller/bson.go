package controller

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"sync/atomic"
	"time"
)

func bsonID() string {
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b, uint32(time.Now().Unix()))
	copy(b[4:], bsonMachine)
	binary.BigEndian.PutUint32(b[8:], atomic.AddUint32(&bsonCounter, 1))
	binary.BigEndian.PutUint16(b[7:], bsonProcess)
	return hex.EncodeToString(b)
}

var (
	bsonProcess = uint16(os.Getpid())
	bsonMachine = func() []byte {
		host, err := os.Hostname()
		if err != nil {
			b := make([]byte, 3)
			if _, err := io.ReadFull(rand.Reader, b); err != nil {
				panic("random error: " + err.Error())
			}
			return b
		}
		hw := md5.New()
		hw.Write([]byte(host))
		return hw.Sum(nil)[:3]
	}()
	bsonCounter = func() uint32 {
		b := make([]byte, 4)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			panic("random error: " + err.Error())
		}
		return binary.BigEndian.Uint32(b)
	}()
)
